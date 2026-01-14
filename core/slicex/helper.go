package slicex

import (
	"context"
	"errors"
	"reflect"
	"slices"
	"strings"

	"github.com/goslacker/slacker/core/reflectx"
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

func Map[T any, R any](s []T, f func(item T) (R, error)) ([]R, error) {
	return MapFilter(s, func(item T) (r R, ok bool, err error) {
		r, err = f(item)
		if err != nil {
			return
		}
		ok = true
		return
	})
}

func MapIdx[T any, R any](s []T, f func(idx int, item T) (R, error)) ([]R, error) {
	return MapFilterIdx(s, func(idx int, item T) (r R, ok bool, err error) {
		r, err = f(idx, item)
		if err != nil {
			return
		}
		ok = true
		return
	})
}

// MustMap maps the slice to a new slice by applying the function to each element.
func MustMap[T any, R any](s []T, f func(item T) R) []R {
	r, _ := Map(s, func(item T) (R, error) {
		return f(item), nil
	})
	return r
}

func MustMapIdx[T any, R any](s []T, f func(idx int, item T) R) []R {
	r, _ := MapIdx(s, func(idx int, item T) (R, error) {
		return f(idx, item), nil
	})
	return r
}

func MapFilter[T any, R any](s []T, f func(item T) (R, bool, error)) ([]R, error) {
	return MapFilterIdx(s, func(idx int, item T) (R, bool, error) {
		return f(item)
	})
}

func MapFilterIdx[T any, R any](s []T, f func(idx int, item T) (R, bool, error)) ([]R, error) {
	ret := make([]R, 0, len(s))
	for idx, item := range s {
		tmp, ok, err := f(idx, item)
		if err != nil {
			return nil, err
		}
		if ok {
			ret = append(ret, tmp)
		}
	}
	return ret, nil
}

func MustMapFilter[T any, R any](s []T, f func(item T) (R, bool)) []R {
	r, _ := MapFilter(s, func(item T) (r R, ok bool, err error) {
		r, ok = f(item)
		return
	})
	return r
}

func FilterCtx[T any](ctx context.Context, s []T, f func(item T) bool) []T {
	var tmp []T
	for _, item := range s {
		select {
		case <-ctx.Done():
			return nil
		default:
			if f(item) {
				tmp = append(tmp, item)
			}
		}
	}
	return tmp
}

func Filter[T any](s []T, f func(item T) bool) []T {
	return FilterCtx(context.Background(), s, f)
}

// ToMap convert a slice to a map, with specified key and value using the function.
func ToMap[S ~[]E, E comparable, K comparable, V any](s S, f func(item E) (key K, value V)) map[K]V {
	m := make(map[K]V, len(s))
	for _, item := range s {
		key, value := f(item)
		m[key] = value
	}

	return m
}

func Index[S ~[]E, E comparable](s S, v ...E) int {
	if len(v) == 0 || len(s) == 0 || len(v) > len(s) {
		return -1
	}

	for idx := range s {
		if s[idx] == v[0] {
			same := true
			for i := range v {
				if v[i] != s[idx+i] {
					same = false
				}
			}
			if same {
				return idx
			}
		}
	}
	return -1
}

var ExitBatch = errors.New("exit batch")

func Batch[S ~[]E, E comparable](s S, batch int, f func(piece S) error) (err error) {
	var idx int
	for {
		start := idx * batch
		if start >= len(s) {
			break
		}
		end := start + batch
		if end > len(s) {
			end = len(s)
		}
		if start == end {
			break
		}

		err = f(s[start:end])
		if err != nil {
			if errors.Is(err, ExitBatch) {
				err = nil
			}
			return
		}
		idx++
	}
	return
}

// EqualIgnoreOrder 判断两个切片是否相等，忽略元素顺序
func EqualIgnoreOrder[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for _, item := range a {
		if !slices.Contains(b, item) {
			return false
		}
	}
	return true
}
