package reflectx

import (
	"fmt"
	"reflect"
	"strings"
)

func FieldByNameCaseInsensitivity(v reflect.Value, name string) (ret reflect.Value) {
	return v.FieldByNameFunc(func(fieldName string) bool {
		return strings.EqualFold(fieldName, name)
	})
}

func IsKind(target any, kind reflect.Kind) bool {
	value := reflect.ValueOf(target)
	return IsKindValue(value, kind)
}

func IsKindValue(value reflect.Value, kind reflect.Kind) bool {
	return value.Kind() == kind
}

func Indirect(v reflect.Value, fillZero bool) (ret reflect.Value) {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			if fillZero && v.CanSet() {
				v.Set(reflect.New(v.Type().Elem()))
			} else {
				return
			}
		}
		v = v.Elem()
	}

	return v
}

func SetValue(dstValue reflect.Value, srcValue reflect.Value) (err error) {
	dstValue = Indirect(dstValue, true)
	srcValue = Indirect(srcValue, false)
	if dstValue.Type() != srcValue.Type() {
		err = fmt.Errorf("type <%s> and <%s> are not same", dstValue.Type().String(), srcValue.Type().String())
		return
	}

	dstValue.Set(srcValue)

	return
}

func NewIndirectType(t reflect.Type) (r reflect.Value) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	r = reflect.New(t)
	return
}
