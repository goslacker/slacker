package errx

import (
	"errors"
	"github.com/goslacker/slacker/core/slogx"
	"log/slog"
	"os"
	"testing"
)

func TestNormal(t *testing.T) {
	opts := &slog.HandlerOptions{
		AddSource:   true,
		ReplaceAttr: slogx.Replacer,
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, opts)))
	err := errors.New("test error")
	err = Wrap(err, "test error2")
	err = Wrap(err, "test error3")
	//fmt.Printf("%v\n", err)
	//slog.Error("tests", "err", err)
	//slog.Error("test", "err", err)
	println(err.Error())
}
