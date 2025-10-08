package tool

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/goslacker/slacker/core/reflectx"
	"github.com/tidwall/gjson"
)

// SimpleMapFuncBack 将src转换成dest处理后再转回来
func SimpleMapFuncBack[S any, D any](src S, f func(dest D) (err error)) (err error) {
	var dest D
	err = SimpleMap(&dest, src)
	if err != nil {
		return
	}
	err = f(dest)
	if err != nil {
		return
	}
	err = SimpleMap(&src, dest)
	return
}

func SimpleMap(dst any, src any) (err error) {
	return SimpleMapValue(reflect.ValueOf(dst), reflect.ValueOf(src), "root")
}

func SimpleMapValue(dst reflect.Value, src reflect.Value, fieldName string) (err error) {
	src = reflectx.Indirect(src, false)
	switch src.Kind() {
	case reflect.Struct, reflect.Invalid:
		return StructValueTo(dst, src)
	case reflect.Slice:
		return SliceValueTo(dst, src, fieldName)
	case reflect.String:
		return StringValueTo(dst, src, fieldName)
	case reflect.Map:
		return MapValueTo(dst, src)
	default:
		reflectx.SetValue(dst, src)
		return
	}
}

func MapValueTo(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.String:
		return StructValueToString(dst, src)
	case reflect.Map:
		return MapValueToMap(dst, src)
	default:
		//err = fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
		return
	}
}

func StringValueTo(dst reflect.Value, src reflect.Value, fieldName string) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Slice:
		if _, ok := dst.Interface().([]byte); ok {
			dst.SetBytes([]byte(src.String()))
		} else {
			if src.String() == "" {
				return
			}
			r := gjson.Parse(src.String())
			if r.IsArray() {
				err = json.Unmarshal([]byte(src.String()), dst.Addr().Interface())
				if err == nil {
					return
				}
			}

			if dst.Type().Elem().Kind() != reflect.String { //不是字符串切片, 不用加引号
				err = json.Unmarshal([]byte("["+src.String()+"]"), dst.Addr().Interface())
				if err == nil {
					return
				}
			}

			strSlice := strings.Split(src.String(), ",")
			err = reflectx.SetValue(dst, reflect.ValueOf(strSlice))
			if err != nil {
				err = fmt.Errorf("2set value directly failed: %w[src=%s, fieldName=%s]", err, src.String(), fieldName)
			}
			return
		}
	case reflect.Struct, reflect.Map:
		if src.String() == "" {
			return
		}
		err = json.Unmarshal([]byte(src.String()), dst.Addr().Interface())
		if err != nil {
			err = fmt.Errorf("string to struct/map json.Unmarshal failed: %w[src=%s, fieldName=%s]", err, src.String(), fieldName)
		}
	default:
		err = reflectx.SetValue(dst, src)
		if err != nil {
			err = fmt.Errorf("1set value directly failed: %w[src=%s, fieldName=%s]", err, src.String(), fieldName)
		}
	}
	return
}

func StructValueTo(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Struct:
		return StructValueToStruct(dst, src)
	case reflect.String:
		return StructValueToString(dst, src)
	case reflect.Slice:
		return StructValueToSlice(dst, src)
	default:
		//err = fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
		return
	}
}

func StructValueToStruct(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	src = reflectx.Indirect(src, false)
	if !src.IsValid() {
		return
	}
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldStruct := src.Type().Field(i)
		if !srcFieldStruct.IsExported() {
			continue
		}
		if srcFieldStruct.Anonymous {
			dstField := dst.FieldByName(srcFieldStruct.Name)
			if !dstField.IsValid() {
				err = StructValueToStruct(dst, srcField)
			} else {
				if dstField.Kind() == srcField.Kind() && dstField.Kind() != reflect.Pointer {
					if !dstField.CanSet() {
						continue
					}
					err = reflectx.SetValue(dstField, srcField)
					if err != nil {
						err = fmt.Errorf("SetValue failed: %w[srcFieldName=%s]", err, srcFieldStruct.Name)
					}
				} else {
					err = StructValueToStruct(dstField, srcField)
				}
			}
			if err != nil {
				return
			}
			continue
		}

		dstField := reflectx.FieldByNameCaseInsensitivity(dst, srcFieldStruct.Name)
		if dstField.IsValid() {
			err = SimpleMapValue(dstField, srcField, srcFieldStruct.Name)
			if err != nil {
				return
			}
		}
	}
	return
}

func SliceValueTo(dst, src reflect.Value, fieldName string) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Slice:
		return SliceValueToSlice(dst, src, fieldName)
	case reflect.String:
		return StructValueToString(dst, src)
	case reflect.Struct:
		return SliceValueToStruct(dst, src)
	default:
		return fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
	}
}

func SliceValueToStruct(dst reflect.Value, src reflect.Value) (err error) {
	if src.Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("slice2struct failed: unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
	}
	bytes := src.Bytes()
	if len(bytes) == 0 {
		return
	}
	err = json.Unmarshal(bytes, dst.Addr().Interface())
	return
}

func SliceValueToSlice(dst reflect.Value, src reflect.Value, fieldName string) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)
	dstItemType := dst.Type().Elem()
	if dst.IsNil() {
		dst.Set(reflect.MakeSlice(dst.Type(), 0, src.Len()))
	}
	for i := 0; i < src.Len(); i++ {
		var dstItem reflect.Value
		if dst.Len()-1 >= i {
			dstItem = reflectx.Indirect(dst.Index(i), false)
			err = SimpleMapValue(dstItem, src.Index(i), fieldName)
			if err != nil {
				return
			}
		} else {
			dstItem = reflect.New(dstItemType)
			err = SimpleMapValue(dstItem.Elem(), src.Index(i), fieldName)
			if err != nil {
				return
			}
			dst.Set(reflect.Append(dst, dstItem.Elem()))
		}

	}
	return
}

func MapValueToMap(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	src = reflectx.Indirect(src, false)
	if dst.IsNil() {
		dst.Set(reflect.MakeMap(dst.Type()))
	}
	for _, key := range src.MapKeys() {
		dst.SetMapIndex(key, src.MapIndex(key))
	}
	return
}

func StructValueToSlice(dst reflect.Value, src reflect.Value) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)
	if dst.Type().Elem().Kind() != reflect.Uint8 {
		return fmt.Errorf("slice2struct failed: unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
	}
	if !src.IsValid() {
		dst.Set(reflect.ValueOf([]byte("null")))
		return
	}

	var result []byte
	if s, ok := src.Interface().(interface{ MapToString() string }); ok {
		result = []byte(s.MapToString())
	} else {
		var tmp []byte
		tmp, err = json.Marshal(src.Interface())
		if err != nil {
			return
		}
		result = tmp
	}
	dst.Set(reflect.ValueOf(result))
	return
}

func StructValueToString(dst reflect.Value, src reflect.Value) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)
	if !src.IsValid() {
		dst.SetString("null")
		return
	}

	if s, ok := src.Interface().(interface{ MapToString() string }); ok {
		dst.SetString(s.MapToString())
	} else {
		var tmp []byte
		tmp, err = json.Marshal(src.Interface())
		if err != nil {
			return
		}
		dst.Set(reflect.ValueOf(string(tmp)))
	}

	return
}
