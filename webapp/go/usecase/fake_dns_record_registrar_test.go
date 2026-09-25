package usecase

type fakeDNSRecordRegistrar struct {
	err error

	gotNames []string
}

func (r *fakeDNSRecordRegistrar) AddRecord(name string) error {
	r.gotNames = append(r.gotNames, name)
	return r.err
}
