package wp

type Result struct {
	value interface{}
	err   error
}

func (r Result) GetValue() interface{} {
	return r.value
}

func (r Result) GetErr() error {
	return r.err
}
