package node

var ErrDeep = &DeepError{}

type DeepError struct {
  Err error
}

func (e *DeepError) Error() string {
  return e.Err.Error()
}

func (e *DeepError) Unwrap() error {
  return e.Err
}

func (e *DeepError) Is(target error) bool {
  return target == ErrDeep
}
