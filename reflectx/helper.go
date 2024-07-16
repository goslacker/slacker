package reflectx

import "reflect"

func FindField(v reflect.Value, name string) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v.FieldByName(name)
}

func IsKind(target any, kind reflect.Kind) bool {
	value := reflect.ValueOf(target)
	return IsKindValue(value, kind)
}

func IsKindValue(value reflect.Value, kind reflect.Kind) bool {
	return value.Kind() == kind
}
