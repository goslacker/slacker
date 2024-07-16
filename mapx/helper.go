package mapx

import "reflect"

func Merge[T any](ms ...T) (ret T) {
	value := reflect.ValueOf(ms[0])
	if value.Kind() != reflect.Map {
		panic("params is not a map")
	}

	valueRet := reflect.MakeMap(value.Type())

	for _, m := range ms {
		value := reflect.ValueOf(m)
		for _, key := range value.MapKeys() {
			valueRet.SetMapIndex(key, value.MapIndex(key))
		}
	}

	return valueRet.Interface().(T)
}
