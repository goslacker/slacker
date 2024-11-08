package slicex

import (
	"errors"
	"github.com/goslacker/slacker/core/reflectx"
	"reflect"
	"strings"
)

func Find[T any](s []T, f func(item T) bool) (T, bool) {
	for _, item := range s {
		if f(item) {
			return item, true
		}
	}

	return *new(T), false
}

func Merge[T any](ss ...[]T) []T {
	ret := make([]T, 0, len(ss))
	for _, s := range ss {
		ret = append(ret, s...)
	}
	return ret
}

func Contains[T comparable](target T, s []T) bool {
	for _, v := range s {
		if v == target {
			return true
		}
	}
	return false
}

func ContainsAny[T comparable](target []T, s []T) bool {
	for _, v := range target {
		if Contains(v, s) {
			return true
		}
	}
	return false
}

func ContainsAll[T comparable](target []T, s []T) bool {
	for _, v := range target {
		if !Contains(v, s) {
			return false
		}
	}
	return true
}

func GetFieldSlice[T any](key string, target any) []T {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
		panic(errors.New("type is not slice or array"))
	}
	finds := make([]reflect.Value, 0, v.Len())
	for i := 0; i < v.Len(); i++ {
		finds = append(finds, v.Index(i))
	}

	keys := strings.Split(key, ".")
	for _, key := range keys {
		for idx, item := range finds {
			field := reflectx.Indirect(item, false).FieldByName(key)
			if field.IsValid() {
				finds[idx] = field
			} else {
				finds[idx] = reflect.Value{}
			}
		}
	}

	result := make([]T, 0, len(finds))
	for _, find := range finds {
		result = append(result, find.Interface().(T))
	}

	return result
}

// SameItem 判断所有元素是否都相同
func SameItem[T comparable](s ...T) bool {
	for i := 0; i < len(s)-1; i++ {
		if s[i] != s[i+1] {
			return false
		}
	}

	return true
}

// Map maps the slice to a new slice by applying the function to each element.
func Map[T any, R any](s []T, f func(item T) R) []R {
	ret := make([]R, 0, len(s))
	for _, item := range s {
		ret = append(ret, f(item))
	}
	return ret
}

func Filter[T any](s []T, f func(item T) bool) []T {
	var tmp []T
	for _, item := range s {
		if f(item) {
			tmp = append(tmp, item)
		}
	}
	return tmp
}

// ToMap convert a slice to a map, with specified key and value using the function.
func ToMap[T any, K comparable, V any](s []T, f func(item T) (key K, value V)) map[K]V {
	m := make(map[K]V, len(s))
	for _, item := range s {
		key, value := f(item)
		m[key] = value
	}

	return m
}
