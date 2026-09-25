package usecase

// fakeDNSRecordRegistrar はテストで設定した関数に処理を委ねる DNSRecordRegistrar。
// メソッドが 1 つだけなのでインターフェースは埋め込まない。関数を設定せずに呼ぶと panic する。
type fakeDNSRecordRegistrar struct {
	addRecord func(name string) error
}

func (r *fakeDNSRecordRegistrar) AddRecord(name string) error {
	return r.addRecord(name)
}
