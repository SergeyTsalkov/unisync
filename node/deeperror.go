package node

type DeepError struct {
	Err error
}

func (e *DeepError) Error() string {
	return e.Err.Error()
}

func (e *DeepError) Unwrap() error {
	return e.Err
}
