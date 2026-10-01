package usecase

// DNSRecordRegistrar はユーザのサブドメインの DNS レコードを登録するためのインターフェース。
type DNSRecordRegistrar interface {
	// AddRecord は name のサブドメインの A レコードを登録する。
	AddRecord(name string) error
}
