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

func Keys[K comparable, V any](m map[K]V) []K {
	var keys []K
	for key := range m {
		keys = append(keys, key)
	}
	return keys
}

func Values[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, value := range m {
		values = append(values, value)
	}
	return values
}

func Map[K comparable, V any, ToK comparable, ToV any](m map[K]V, f func(key K, value V) (newKey ToK, newV ToV)) map[ToK]ToV {
	ret := make(map[ToK]ToV, len(m))
	for key, value := range m {
		newKey, newV := f(key, value)
		ret[newKey] = newV
	}

	return ret
}
