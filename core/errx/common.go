package errx

import (
	"fmt"
	"runtime"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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

func WithSkip(skip int) func(*Error) {
	return func(e *Error) {
		e.skip = skip
	}
}

func newErr(opt ...func(*Error)) *Error {
	e := &Error{}
	for _, set := range opt {
		set(e)
	}

	if e.Detail == nil {
		e.Detail = make(map[string]any)
	}

	stack := make([]byte, 4096)
	n := runtime.Stack(stack, false)
	e.Detail["stack"] = string(stack[:n])

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
	if e.Message == "" {
		e.Message = err.Error()
	}
	return e
}

//type Error struct {
//	message string
//	err     error
//	line    int
//	file    string
//
//	detail map[string]any
//}

type Error struct {
	Message string
	err     error
	skip    int
	Code    int
	Detail  map[string]any
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) Format(s fmt.State, c rune) {
	switch c {
	case 'v':
		switch {
		case s.Flag('+'):
			_, _ = s.Write([]byte(e.Error()))
		case s.Flag('#'):
			_, _ = s.Write([]byte(e.Error() + "\n\n" + e.Detail["stack"].(string)))
		default:
			_, _ = s.Write([]byte(e.Error()))
		}
	}
}

func GrpcStatusError(code codes.Code, msg string) error {
	return Wrap(status.Error(code, msg))
}
