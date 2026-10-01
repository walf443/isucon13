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
| `usecase/repository` | repository の interface と `Querier` (実行先を表す不透明な値) / `TxManager` / `ErrNotFound` |
| `infra/mysql` | repository と `TxManager` の MySQL 実装 (GORM)、テーブルごとの行の型、DB への接続 (`Config` / `ConfigFromEnv` / `Open`) |
| `infra/powerdns` | `usecase.DNSRecordRegistrar` の実装 (pdnsutil を呼ぶ) |
| `infra/script` | `usecase.Initializer` の実装 (初期化スクリプトを呼ぶ) |
| `interfaces/http/handler` | HTTP の入口。リクエストの解釈、usecase の呼び出し、レスポンスへの変換、ルーティング、セッション |
| `main` | 設定の読み込みと、repository・usecase・handler の組み立て |

依存の向きは golangci-lint の depguard で強制している (`.golangci.yml`)。

- `domain` は `usecase` / `infra` / `interfaces` / `database/sql` に依存しない
- `usecase` (`usecase/repository` を含む) は `infra` / `interfaces` / `database/sql` / MySQL ドライバ / GORM に依存しない
- `domain` と `interfaces` も GORM に依存しない
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
- `Querier` は usecase から見ると不透明な値で、usecase は `TxManager` から受け取って repository に渡すだけで、中身を見ない
  - 小文字のメソッド `querier()` を持つ interface で、`repository.QuerierBase` を埋め込んだ型 (infra の `querier`) だけが満たせる。関係のない値をうっかり渡すとコンパイルエラーになる
  - 実体は `*gorm.DB` (トランザクション) で、infra の repository が `gormOf` / `dbOf` で取り出す。usecase は GORM に依存しない

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

### SQL (GORM)

infra/mysql の SQL は GORM で書く。

- 単純な処理 (ID・カラムでの取得、INSERT / DELETE / UPDATE、JOIN の件数・集計) は、ビルダー (`Where` / `Find` / `Create` / `Joins` / `Count` など) で書く
  - 読み取りは必ず `Select("id, name")` のようにカラムを列挙する。`SELECT *` にしない (移行前の SQL と同じカラムだけを読み、テーブルにカラムが増えても結果が変わらないようにする)
  - `FOR UPDATE` は `clause.Locking`、`slot = slot - 1` は `gorm.Expr` で書く
- 移行前に SQL で判定していたロジック (NG ワードの LIKE によるスパム判定など) は、Go に移さず、リテラルの SQL を `Raw` / `Exec` で実行する。Go に移すと照合順序やワイルドカードの扱いまで含めた同一性を保証できないため
- 行の型はテーブルごとに infra に作る (`userRow` など)。`TableName()` を必ず明示し、カラムは `gorm:"column:..."` で対応づけ、domain の型との変換 (`toDomain`) を持たせる
  - `CreatedAt` は GORM が登録時刻を自動で設定する名前なので、`autoCreateTime:false` を付ける (usecase が渡した値をそのまま保存する)
  - 見つからない場合は `gorm.ErrRecordNotFound` を `repository.ErrNotFound` に変換する
- GORM の設定: ロガーは出力しない (移行前は SQL をログに出していなかった)、`SkipDefaultTransaction` を有効にする (トランザクションは usecase が持つ)、エラーの翻訳はしない (ドライバのエラー本文をそのまま返す。例: 重複登録の `Error 1062`)
- `IN (...)` は `findIn` を使う。ID を昇順・重複なしにして 1000 件ずつに分けて引く (MySQL のプレースホルダーは 1 つのクエリで 65535 個までのため)。分割すると結果の並び順は崩れるので、並び順が必要なメソッドは並べ直す
- 発行される SQL の文面は `gorm_test.go` のテスト (`recordSQL`) で固定している。repository を足したり変えたりしたら、そこに足す

### domain

- 型の名前は層での役割で分ける
  - domain のエンティティ (テーブルの行に対応する型) と値オブジェクトは接尾辞を付けない (`domain.User`、`domain.Livestream`、`domain.Limit`)
  - 関連するデータを埋めて組み立てたものは `Detail` を付ける (`domain.UserDetail`、`domain.LivestreamDetail`)。組み立てた状態を使うドメインのロジックを、そのメソッドとして書けるように domain に置いている
  - handler のレスポンス・リクエストの型は `Response` / `Request` を付けて非公開にする (`userResponse`、`postUserRequest`)
- 型付き ID (`domain.ID[T]`、`domain.UserID` など) を使い、文字列からの変換は `domain.ParseXxxID` で行う
- 値オブジェクト: `Limit` (`ParseLimit` で範囲を検証して作る)、`ReservationPeriod`、`HashedPassword`、`IconHash` など
- `json` タグは付けない。domain の構造をそのまま HTTP に出してしまわないため
- `db` タグは、sqlx を使っていたときの名残で、今は誰も使っていない (infra が行の型を別に持つため)。差分を小さくするために、そのまま残している。新しい型には付けなくてよい

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
- sqlx から GORM に移行したことで、発行される SQL の文面が変わった (結果は同じ)
  - `COUNT(*)` が `count(*)` になった、`GROUP BY` のカラム名にバッククォートが付いた、`LIMIT 1` が `LIMIT ?` になった、SQL の空白が保たれなくなった
  - `FindImageByUserID` は `ORDER BY id LIMIT 1` になった (同じユーザのアイコンが複数ある場合に、ID が最小のものを使うことを明示した)
  - MySQL ドライバが v1.7.1 から v1.8.1 に上がった (GORM の MySQL ドライバの要求)

## テスト

| 対象 | 方法 |
|---|---|
| `domain` | 通常の単体テスト |
| `usecase` | repository などを fake に差し替える。fake は interface を埋め込み、テストで設定した関数に処理を委ねる (未設定のメソッドを呼ぶと panic する)。Filler は fake にせず、テスト用のデータ (`fixture_test.go` の `detailFixture`) から作った本物を使い、組み立てた結果まで確認する |
| `interfaces/http/handler` | usecase を fake に差し替え、`serve` / `send` で echo にリクエストを送り、`assertResponse` で確認する。エラーレスポンスでは本文 (`wantBody`) の確認を必須にしている |
| `infra/mysql` | testcontainers で MySQL を起動し、実際のスキーマに対して 1 メソッドずつ確認する。`-short` ではスキップする。テストのデータは `insertSQL` / `execSQL` / `scanSQL` (`helper_test.go`) で SQL を直接実行して用意する。発行される SQL の文面の固定 (`recordSQL`)、7 万件の ID での一括取得、`FOR UPDATE` が本当にロックすること (別のトランザクションから `NOWAIT` で確認) など、実際の MySQL でしか分からないことも確かめている |
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
| `db` タグを domain から外す | もう使われていないが、消すと差分が大きくなるので、そのまま残している |
| 統計の集約 repository・統計 usecase の分割 | テーブル単位でクエリがまとまっていることの方が価値がある。分けても改善が小さい |
| ベンチマーカーでの確認 | 目的は Clean Architecture の実践で、環境構築の手間に見合わない |
