package errx

func WithCode(code int) func(*Error) {
	return func(e *Error) {
		e.Code = code
	}
}

func WithDetail(detail map[string]any) func(*Error) {
	return func(e *Error) {
		e.Detail = detail
	}
}

func WithErr(err error) func(*Error) {
	return func(e *Error) {
		e.err = err
	}
}

func WithMsg(msg string) func(*Error) {
	return func(e *Error) {
		e.Message = msg
	}
}

func newErr(opt ...func(*Error)) error {
	e := &Error{}
	for _, set := range opt {
		set(e)
	}
	return e
}

func New(msg string, opt ...func(*Error)) error {
	opt = append(opt, WithMsg(msg))
	return newErr(opt...)
}

func Wrap(err error, opt ...func(*Error)) error {
	if err == nil {
		return nil
	}
	opt = append(opt, WithErr(err))
	e := newErr(opt...)
	if e.(*Error).Message == "" {
		e.(*Error).Message = err.Error()
	}
	return e
}

type Error struct {
	Message string
	err     error
	Code    int
	Detail  map[string]any
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.err
}
