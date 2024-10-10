package fmtx

import (
	"fmt"
	"regexp"
)

// Errorf extend fmt.Errorf. same as fmt.Errorf, but only return nil when all error params are nil.
func Errorf(format string, a ...any) error {
	reg := regexp.MustCompile(`(%[a-zA-Z])|(%[#+-][a-z])|(%\.\d[a-z])`)
	matched := reg.FindAllString(format, -1)

	shouldRetrunNil := true
	for idx, item := range matched {
		if item == "%w" && a[idx] != nil {
			shouldRetrunNil = false
			break
		}
	}
	if shouldRetrunNil {
		return nil
	}

	return fmt.Errorf(format, a...)
}
