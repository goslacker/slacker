package errx

import (
	"errors"
	"fmt"
	"log/slog"
	"testing"
)

//	func Test(t *testing.T) {
//		err := newErr(WithMsg("test"))
//		fmt.Printf("%#v\n", err)
//		// println(err.Detail["stack"].(string))
//	}
func TestNormal(t *testing.T) {
	err := errors.New("test error")
	err = Wrap(err, "test error2")
	err = Wrap(err, "test error3")
	fmt.Printf("%v\n", err)
	slog.Info("test", "info", fmt.Sprintf("%v\n", err))
	//println(err.Error())
}
