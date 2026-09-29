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
- repository のメソッドは 1 つの SQL を実行し、テーブルの行に対応する型 (`domain.XxxModel`) を返す

### 詳細の組み立ては usecase の Filler が行う

複数のテーブルを組み合わせた読み取り用のモデル (`domain.User` / `Livestream` / `Livecomment` / `Reaction` / `LivecommentReport`) は、
usecase の `XxxFiller` (`UserFiller` / `LivestreamFiller` など) が repository の単発のメソッドを組み合わせて組み立てる。

- `Fill(ctx, q, models)` は、行の型の一覧を受け取り、ID ごとの map で返す。並び順は呼び出し側で決める (`orderedBy` で元の順序に並べ直す)
  - 1 件だけの場合も長さ 1 のスライスで渡す。まとめて取得する作りにしておけば、将来 `IN (...)` で取得するように変えても呼び出し側は変わらない
- 同じ ID が複数含まれていても 1 回だけ取得する
- 主となる行 (ユーザ・ライブ配信など) は呼び出し側が引き、見つからない場合の 404 への変換やエラーメッセージも呼び出し側で決める
  - 付随するデータ (テーマ・配信者・タグなど) が無いのはデータ不整合なので 500 にする。Filler のエラーを 404 に変換しないこと
- Filler は他の Filler を使う (`LivestreamFiller` は配信者に `UserFiller` を使う、など)。`usecases.go` でそれぞれ 1 つ作って共有する

### SQL

- SQL は完全なリテラルで書く。カラム一覧の定数化や文字列の連結による動的な組み立てはしない
- 移行前に SQL で判定していたロジック (NG ワードの LIKE によるスパム判定など) は SQL のまま残す。Go に移すと照合順序やワイルドカードの扱いまで含めた同一性を保証できないため

### domain

- 型付き ID (`domain.ID[T]`、`domain.UserID` など) を使い、文字列からの変換は `domain.ParseXxxID` で行う
- 値オブジェクト: `Limit` (`ParseLimit` で範囲を検証して作る)、`ReservationPeriod`、`HashedPassword`、`IconHash` など
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
- 詳細の組み立てを Filler に移したことで、重複した ID の取得が 1 回になり、SQL の順序・回数が変わる場合がある (結果は同じ)
- データ不整合 (テーマなどの欠損) の場合の 500 の本文が `sql: no rows in result set` から `not found` になった

## テスト

| 対象 | 方法 |
|---|---|
| `domain` | 通常の単体テスト |
| `usecase` | repository などを fake に差し替える。fake は interface を埋め込み、テストで設定した関数に処理を委ねる (未設定のメソッドを呼ぶと panic する)。Filler は fake にせず、テスト用のデータ (`livestreamFixture`) から作った本物を使い、組み立てた結果まで確認する |
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
| 統計の集約 repository・統計 usecase の分割 | テーブル単位でクエリがまとまっていることの方が価値がある。分けても改善が小さい |
| ベンチマーカーでの確認 | 目的は Clean Architecture の実践で、環境構築の手間に見合わない |
