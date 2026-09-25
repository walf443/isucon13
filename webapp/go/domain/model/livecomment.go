package model

type LivecommentID = ID[LivecommentModel]

// ParseLivecommentID は 10 進数の文字列をライブコメントの ID として読み取る。
func ParseLivecommentID(s string) (LivecommentID, error) {
	return ParseID[LivecommentModel](s)
}

type LivecommentModel struct {
	ID           LivecommentID `db:"id"`
	UserID       UserID        `db:"user_id"`
	LivestreamID LivestreamID  `db:"livestream_id"`
	Comment      string        `db:"comment"`
	Tip          int64         `db:"tip"`
	CreatedAt    int64         `db:"created_at"`
}

// Livecomment はコメントしたユーザ・ライブ配信を含めたライブコメントの情報。
type Livecomment struct {
	ID         LivecommentID
	User       User
	Livestream Livestream
	Comment    string
	Tip        int64
	CreatedAt  int64
}

type LivecommentReportID = ID[LivecommentReportModel]

type LivecommentReportModel struct {
	ID            LivecommentReportID `db:"id"`
	UserID        UserID              `db:"user_id"`
	LivestreamID  LivestreamID        `db:"livestream_id"`
	LivecommentID LivecommentID       `db:"livecomment_id"`
	CreatedAt     int64               `db:"created_at"`
}

// LivecommentReport は報告したユーザ・報告されたライブコメントを含めた報告の情報。
type LivecommentReport struct {
	ID          LivecommentReportID
	Reporter    User
	Livecomment Livecomment
	CreatedAt   int64
}
