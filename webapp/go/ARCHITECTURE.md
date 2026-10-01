# webapp/go のアーキテクチャ

ISUCON13 の参考実装 (isupipe) を、挙動を変えずに Clean Architecture の層に分け直したもの。
Go で Clean Architecture を実践することが目的なので、ここには「どこに何を置くか」と「なぜそうしたか」を残す。

## 層と依存の向き

```
domain  ←  usecase (+ usecase/repository)  ←  infra/*
                                           ←  interfaces/http/handler
```

右側のパッケージが左側のパッケージに依存する (import する)。main パッケージ (`main.go`, `usecases.go`) だけが全ての層を知っていて、組み立てを行う。

| パッケージ | 役割 |
|---|---|
| `domain` | エンティティ・値オブジェクト・型付き ID・ドメインのルール。他の層にも DB にも依存しない |
| `usecase` | ユースケース。トランザクションの境界を持つ。外部 (DNS・初期化スクリプト・ログ) への要求は interface として定義する |
| `usecase/repository` | repository の interface と `Querier` / `TxManager` / `ErrNotFound` |
| `infra/mysql` | repository と `TxManager` の MySQL 実装、DB への接続 (`Config` / `ConfigFromEnv` / `Open`) |
| `infra/powerdns` | `usecase.DNSRecordRegistrar` の実装 (pdnsutil を呼ぶ) |
| `infra/script` | `usecase.Initializer` の実装 (初期化スクリプトを呼ぶ) |
| `interfaces/http/handler` | HTTP の入口。リクエストの解釈、usecase の呼び出し、レスポンスへの変換、ルーティング、セッション |
| `main` | 設定の読み込みと、repository・usecase・handler の組み立て |

依存の向きは golangci-lint の depguard で強制している (`.golangci.yml`)。

- `domain` は `usecase` / `infra` / `interfaces` / `database/sql` に依存しない
- `usecase` (`usecase/repository` を含む) は `infra` / `interfaces` / `database/sql` / `sqlx` / MySQL ドライバに依存しない
- `infra` は `interfaces` に依存しない
- `interfaces` は `infra` と `usecase/repository` に依存しない (データには usecase を通してアクセスする)

## 設計の方針

### interface は使う側に置く

usecase が外部に求めるもの (repository・DNS の登録・初期化・ログ) は、全て usecase 側で interface として定義し、infra がそれを実装する。
Clean Architecture の原典でデータアクセスの interface がユースケース層に置かれていることと、Go の「interface は使う側で定義する」慣習に合わせている。
repository だけは数が多く、usecase の型と名前を分けたいので `usecase/repository` にまとめている。

### トランザクションは usecase が持つ

- usecase のメソッドは `TxManager.RunInTx` の中で repository を呼ぶ。コミット・ロールバックの判断は usecase が行う
- repository のメソッドは実行先の `repository.Querier` (DB またはトランザクション) を引数で受け取る
  - トランザクションの境界がコード上で見えることと、repository を 1 メソッド (1 SQL) ずつテストしやすいことを優先した
- `Querier.ExecContext` は `sql.Result` ではなく `repository.Result` を返し、usecase 側を `database/sql` から切り離している

### repository はテーブル単位

- 1 つのテーブルに対するクエリを 1 つの repository にまとめる。統計のような画面単位の集約 repository は作らない
- repository のメソッドは 1 つの SQL を実行し、テーブルの行に対応する domain のエンティティ (`domain.User` など) を返す

### 詳細の組み立ては usecase の Filler が行う

複数のテーブルを組み合わせた読み取り用のモデル (`domain.UserDetail` / `LivestreamDetail` / `LivecommentDetail` / `ReactionDetail` / `LivecommentReportDetail`) は、
usecase の `XxxFiller` (`UserFiller` / `LivestreamFiller` など) が repository のメソッドを組み合わせて組み立てる。

- `Fill(ctx, q, models)` は、行の型の一覧を受け取り、ID ごとの map で返す。並び順は呼び出し側で決める (`orderedBy` で元の順序に並べ直す)
  - 1 件だけの場合も長さ 1 のスライスで渡す
- 付随するデータは、テーブルごとに 1 回のクエリ (`FindAllByIDs` などの `IN (...)`) でまとめて取得する。件数が増えてもクエリの回数は変わらない
  - 渡す ID の一覧は `uniqueKeys` で重複を除き、結果は `indexBy` で ID ごとの map にしてから組み立てる
  - 一括取得のメソッドは見つからない ID を結果に含めないだけなので、欠けているかどうかは Filler が判定する
- 主となる行 (ユーザ・ライブ配信など) は呼び出し側が引き、見つからない場合の 404 への変換やエラーメッセージも呼び出し側で決める
  - 付随するデータ (テーマ・配信者・タグなど) が無いのはデータ不整合なので 500 にする
  - Filler は付随するデータが欠けている場合、`repository.ErrNotFound` としては判定されない `errMissingDetail` を返す (メッセージは `not found` のまま)。
    呼び出し側が誤って Filler のエラーに `errors.Is(err, repository.ErrNotFound)` を使っても、404 にならない
- Filler は他の Filler を使う (`LivestreamFiller` は配信者に `UserFiller` を使う、など)。`usecases.go` でそれぞれ 1 つ作って共有する

### 型で取り違えを防ぐ

同じ `string` や `int64` でも、意味が違うものには名前付きの型を付ける (`UserID` と `LivestreamID`、`Username` と `DisplayName`、`PlainPassword` と `HashedPassword` など)。
引数の取り違えがコンパイルで見つかり、シグネチャを見れば何を渡すのかがわかる。

**作り方は、値の出どころで分ける。**

| 出どころ | 作り方 |
|---|---|
| 外部からの入力 (パスパラメータ、リクエストの本文、クエリ) | `domain.ParseXxx(s) (T, error)`。失敗しうることをシグネチャで表す |
| すでに登録済みで信頼できる値 (DB から読んだ行) | 型の変換 (`domain.Username(s)`) でよい。infra が行をスキャンするときはこちら |

- Go では型の変換を禁止できないので、外部の入力に型の変換を直接使わないのは規約。型のコメントに書いておく (テストでは使ってよい)
- 外部の入力を受けるリクエストの型のフィールドは `string` / `int64` のままにして、handler が明示的に `ParseXxx` を呼ぶ。
  リクエストの型のフィールドを `domain.Username` などにすると、JSON の読み込みで検証を通らずに入ってしまう
- 検証に失敗したときのレスポンスは、そのエンドポイントの「存在しない」場合に揃える (参照 API や `login` は、不正な形のユーザ名のユーザは存在しないので 404 / 401 のまま)。
  新しく作るもの (登録など) だけ 400 にする。メッセージは固定で、入力の値は含めない
- 検証の規則は、その値が使われる先の要件から決める。例えばユーザ名はそのままサブドメイン (`<ユーザ名>.u.isucon.dev`) になるので、
  DNS のラベルとして使える形 (英数字とハイフンだけの 1〜63 文字で、先頭と末尾がハイフンでない。`isDNSLabel`) だけを受け付ける。ベンチマーカーが作る名前の形は参考にとどめ、規則の根拠にしない
  - 先頭が `-` の名前は `pdnsutil` のオプションに、`.` を含む名前は別の階層のサブドメインになりうる

**HTTP の表現は domain の型に持ち込まない。** JSON の数値の配列 (`tags`) は `[]int64` のまま受け、handler の `parseTagIDs` で `[]domain.TagID` にする。
今は件数の上限 (`maxTagCount`) だけを検証しているが、`error` を返す形にしてあるので、検証を足しても呼び出し側は変わらない。

**表示してはいけない値は、自分で伏せ字になる型にする。** `PlainPassword` は `String()` と `GoString()` が `[REDACTED]` を返すので、
`%v` / `%+v` / `%#v` やエラーへの埋め込みでも、パスワードがログに出ない。ハッシュ化と照合には元の値が使われる。

**保存や外への出力の境界では、素の型に戻す。**
- セッションは gob で保存するので、独自の型は登録が要る。保存する形式を変えないため、`string` / `int64` に変換して入れる (`string(user.Name)`、`int64(user.ID)`)
- DNS の登録 (`AddRecord`) には、ユーザ名とサブドメインが別の概念なので、`string(input.Name)` と明示的に変換して渡す
- レスポンスの型は `string` で持ち、`newXxx` で変換する

### SQL

- SQL は完全なリテラルで書く。カラム一覧の定数化や文字列の連結による動的な組み立てはしない
- 移行前に SQL で判定していたロジック (NG ワードの LIKE によるスパム判定など) は SQL のまま残す。Go に移すと照合順序やワイルドカードの扱いまで含めた同一性を保証できないため

### domain

- 型の名前は層での役割で分ける
  - domain のエンティティ (テーブルの行に対応する型) と値オブジェクトは接尾辞を付けない (`domain.User`、`domain.Livestream`、`domain.Limit`)
  - 関連するデータを埋めて組み立てたものは `Detail` を付ける (`domain.UserDetail`、`domain.LivestreamDetail`)。組み立てた状態を使うドメインのロジックを、そのメソッドとして書けるように domain に置いている
  - handler のレスポンス・リクエストの型は `Response` / `Request` を付けて非公開にする (`userResponse`、`postUserRequest`)
- 型付き ID (`domain.ID[T]`、`domain.UserID` など) を使い、文字列からの変換は `domain.ParseXxxID` で行う
- 値オブジェクト: `Limit` (`ParseLimit` で範囲を検証して作る)、`Username` (`ParseUsername`)、`PlainPassword`、`ReservationPeriod`、`HashedPassword`、`IconHash` など
  - 作り方と使い分けは「型で取り違えを防ぐ」を参照
- `json` タグは付けない。domain の構造をそのまま HTTP に出してしまわないため
- `db` タグは付けてよい。外すと infra に domain とほぼ同じ行の型が増えるだけなので、テーブルと形がずれるまでは domain に置く

### usecase

- 関連するメソッドを 1 つの usecase にまとめる (1 ユースケース 1 型にはしない)
- 依存する repository などがほかのメソッドと大きく異なるときだけ分ける (例: `LivestreamReservationUsecase`、`UserRegistrationUsecase`、`LivecommentReportUsecase`)
- 失敗の種類は `ErrXxx` またはエラー型で返し、handler が HTTP のステータスに変換する
  - エラーメッセージを利用者の入力から組み立てない

### handler

- Request / Response の型はパッケージ内に閉じる (非公開)。domain の型は `newXxx` などで明示的にレスポンスの型へ変換する
- エラーは `errors.Is` / `errors.AsType` で判定して `echo.NewHTTPError` にする。レスポンスは `errorResponseHandler` が `{"error":"code=N, message=..."}` の形にする
- 認証はルート単位の middleware (`requireSession`) で行い、`routes.go` で認証が必要なルートに `auth` を付ける
  - `e.Group` に middleware を付けると echo が catch-all ルートを登録し、存在しないパスが 404 でなくなるため、ルートごとに渡している
  - 付け忘れは `TestRegisterRoutes_RequireSession` が公開 API の一覧と突き合わせて検出する
- 一覧 API の `limit` は API ごとの上限を持ち、範囲外は 400 にする

### 設定

- DB に関する環境変数 (`ISUCON13_MYSQL_DIALCONFIG_*`) は `infra/mysql.ConfigFromEnv` で読む。バッチなど別の入口からも同じ設定で接続できるようにするため
- それ以外 (セッションの秘密鍵、PowerDNS のアドレス) は main で読む

### 挙動を変えない

移行の目的は構造の整理なので、ステータスコード・レスポンス本文・ログ・発行する SQL は移行前と同じに保つ。
明らかな不具合の修正や、層を分けたことで生じる差は、合意の上でコミットメッセージに残す。これまでに合意した差は次のとおり。

- 認証を middleware にしたことで `POST /api/livestream/:livestream_id/reaction` の検証順が変わった (セッション無しで不正な ID の場合 400 → 403)
- 詳細の組み立てを Filler に移し、付随するデータを `IN (...)` でまとめて取得するようにした (N+1 の解消)。発行する SQL の文面・順序・回数は変わるが、結果は同じ
  - タグは紐付けの ID の昇順に並べる (`ORDER BY id` を明示した)
  - DB のエラーの 500 の本文は `failed to get themes: …` のように、まとめて取得したことを表すものになった
- データ不整合 (テーマなどの欠損) の場合の 500 の本文が `sql: no rows in result set` から `not found` になった
- ユーザ名を DNS のラベルとして検証するようにした。`POST /api/register` は、サブドメインとして使えない名前を 400 にする (これまでは DNS の登録の結果に任されていた)。
  参照 API とログインは、不正な形のユーザ名を存在しないユーザと同じ応答にしたので、ステータスは変わらない
- `POST /api/livestream/reservation` は、タグが 1000 件以上の場合 400 にする (これまで件数の上限は無かった)

## テスト

| 対象 | 方法 |
|---|---|
| `domain` | 通常の単体テスト |
| `usecase` | repository などを fake に差し替える。fake は interface を埋め込み、テストで設定した関数に処理を委ねる (未設定のメソッドを呼ぶと panic する)。Filler は fake にせず、テスト用のデータ (`fixture_test.go` の `detailFixture`) から作った本物を使い、組み立てた結果まで確認する |
| `interfaces/http/handler` | usecase を fake に差し替え、`serve` / `send` で echo にリクエストを送り、`assertResponse` で確認する。エラーレスポンスでは本文 (`wantBody`) の確認を必須にしている |
| `infra/mysql` | testcontainers で MySQL を起動し、実際のスキーマに対して 1 メソッドずつ確認する。`-short` ではスキップする |
| `infra/powerdns`, `infra/script` | 一時ディレクトリに置いた偽のコマンド・スクリプトを実行させて、渡す引数と出力・エラーの扱いを確認する |

## 静的検査と CI

- `.github/workflows/go.yml` で `go build` / `go vet` / `go test -v` / golangci-lint を並列に実行する
- golangci-lint は標準の linter に depguard (層の依存ルール) を加えている。設定は `.golangci.yml`

## パッケージが大きくなったら

今は層ごとに 1 パッケージで十分なので分けていない。テストの実行時間 (特に `infra/mysql`) などで必要になったら、各層の中で機能単位のサブパッケージに切り出す。

## 見送ったこと

| 内容 | 理由 |
|---|---|
| トランザクションを context で運ぶ | 境界が見えにくくなり、repository を 1 メソッドずつテストしにくくなる |
| スパム判定を domain に移す | SQL (LIKE) と同じ判定になることを保証しにくい |
| `db` タグを domain から外す | infra に domain とほぼ同じ型が増えるだけ |
| `Username` を構造体にして `ParseUsername` 以外では作れなくする | JSON の読み込みや DB のスキャンのためのメソッドが増える。型の変換を外部の入力に使わない規約と、レビューで足りる |
| リクエストの `tags` を `[]domain.TagID` で受ける | HTTP の表現 (JSON の数値) と domain の型は扱いが違う。変換は handler の `parseTagIDs` に置く |
| 統計の集約 repository・統計 usecase の分割 | テーブル単位でクエリがまとまっていることの方が価値がある。分けても改善が小さい |
| ベンチマーカーでの確認 | 目的は Clean Architecture の実践で、環境構築の手間に見合わない |
