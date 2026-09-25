package usecase

type fakeInitializer struct {
	out []byte
	err error

	runs int
}

func (i *fakeInitializer) Initialize() ([]byte, error) {
	i.runs++
	return i.out, i.err
}
