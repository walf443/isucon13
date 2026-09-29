package domain

type LivecommentID = ID[Livecomment]

// ParseLivecommentID は 10 進数の文字列をライブコメントの ID として読み取る。
func ParseLivecommentID(s string) (LivecommentID, error) {
	return ParseID[Livecomment](s)
}

type Livecomment struct {
	ID           LivecommentID `db:"id"`
	UserID       UserID        `db:"user_id"`
	LivestreamID LivestreamID  `db:"livestream_id"`
	Comment      string        `db:"comment"`
	Tip          int64         `db:"tip"`
	CreatedAt    int64         `db:"created_at"`
}

// LivecommentDetail はコメントしたユーザ・ライブ配信を含めたライブコメントの情報。
type LivecommentDetail struct {
	ID         LivecommentID
	User       UserDetail
	Livestream LivestreamDetail
	Comment    string
	Tip        int64
	CreatedAt  int64
}

type LivecommentReportID = ID[LivecommentReport]

type LivecommentReport struct {
	ID            LivecommentReportID `db:"id"`
	UserID        UserID              `db:"user_id"`
	LivestreamID  LivestreamID        `db:"livestream_id"`
	LivecommentID LivecommentID       `db:"livecomment_id"`
	CreatedAt     int64               `db:"created_at"`
}

// LivecommentReportDetail は報告したユーザ・報告されたライブコメントを含めた報告の情報。
type LivecommentReportDetail struct {
	ID          LivecommentReportID
	Reporter    UserDetail
	Livecomment LivecommentDetail
	CreatedAt   int64
}
