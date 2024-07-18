package slicex

import (
	"errors"
	"github.com/goslacker/slacker/extend/reflectx"
	"reflect"
	"strings"
)

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
