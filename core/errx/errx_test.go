package errx

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	err := newErr(WithMsg("test"))
	fmt.Printf("%#v\n", err)
	// println(err.Detail["stack"].(string))
}
