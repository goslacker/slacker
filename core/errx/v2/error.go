package errx

import (
	"errors"
	"fmt"
	"github.com/goslacker/slacker/core/slogx"
	"log/slog"
	"runtime"
	"strings"
)

func init() {
	slogx.RegisterValueBasedReplacer(&Error{}, SlogAttrReplacer)
}

func WithSkip(skip int) func(*ErrorOption) {
	return func(option *ErrorOption) {
		option.Skip = skip
	}
}

func WithDetail(detail map[string]any) func(*ErrorOption) {
	return func(option *ErrorOption) {
		option.Detail = detail
	}
}

func Wrap(err error, message ...string) error {
	if len(message) == 0 {
		return wrap(err, err.Error(), 1)
	} else {
		return wrap(err, message[0], 1)
	}
}

func wrap(err error, message string, skip int) error {
	e := New(message, WithSkip(skip+1))
	e.(*Error).err = err
	return e
}

func Errorf(template string, args ...interface{}) error {
	err := fmt.Errorf(template, args...)
	msg := err.Error()
	switch x := err.(type) {
	case interface{ Unwrap() error }:
		err = x.Unwrap()
	case interface{ Unwrap() []error }:
		err = errors.Join(x.Unwrap()...)
	}
	return wrap(err, msg, 1)
}

func New(message string, opts ...func(*ErrorOption)) error {
	option := &ErrorOption{
		Detail: make(map[string]any),
	}
	for _, opt := range opts {
		opt(option)
	}
	_, file, line, _ := runtime.Caller(option.Skip + 1)
	return &Error{
		message: message,
		file:    file,
		line:    line,
		detail:  option.Detail,
	}
}

type ErrorOption struct {
	Skip   int
	Detail map[string]any
}

type Error struct {
	message string
	err     error
	file    string
	line    int
	detail  map[string]any
}

func (e *Error) Stack() []string {
	if e.err == nil {
		return []string{fmt.Sprintf("%s <%s:%d>", e.message, e.file, e.line)}
	} else {
		if x, ok := e.err.(*Error); ok {
			return append(x.Stack(), fmt.Sprintf("%s <%s:%d>", e.message, e.file, e.line))
		} else {
			return []string{e.err.Error(), fmt.Sprintf("%s <%s:%d>", e.message, e.file, e.line)}
		}
	}
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) Format(s fmt.State, c rune) {
	stack := e.Stack()
	switch c {
	case 'v':
		msg := strings.Join(stack, "\n")
		_, _ = s.Write([]byte(msg))
	default:
		_, _ = s.Write([]byte(e.Error()))
	}
}

func (e *Error) Unwrap() error {
	return e.err
}

func SlogAttrReplacer(groups []string, a slog.Attr) slog.Attr {
	if x, ok := a.Value.Any().(*Error); ok {
		return slog.Attr{
			Key:   a.Key,
			Value: slog.AnyValue(x.Stack()),
		}
	}
	return a
}
