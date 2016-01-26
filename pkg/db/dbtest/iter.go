package dbtest

type TestIter struct {
}

func (t *TestIter) All(result interface{}) error {
	return nil
}

func (t *TestIter) Close() error {
	return nil
}

func (t *TestIter) Err() error {
	return nil
}

func (t *TestIter) For(result interface{}, f func() error) error {
	return nil
}

func (t *TestIter) Next(result interface{}) bool {
	return false
}

func (t *TestIter) Timeout() bool {
	return false
}
