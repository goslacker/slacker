package tool

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/goslacker/slacker/core/reflectx"
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
	return SimpleMapValue(reflect.ValueOf(dst), reflect.ValueOf(src))
}

func SimpleMapValue(dst reflect.Value, src reflect.Value) (err error) {
	if (src.Kind() != reflect.Struct && src.IsZero()) || !src.IsValid() {
		return
	}
	src = reflectx.Indirect(src, false)
	switch src.Kind() {
	case reflect.Struct:
		return StructValueTo(dst, src)
	case reflect.Slice:
		return SliceValueTo(dst, src)
	case reflect.String:
		return StringValueTo(dst, src)
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
	default:
		//err = fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
		return
	}
}

func StringValueTo(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Slice:
		if _, ok := dst.Interface().([]byte); ok {
			dst.SetBytes([]byte(src.String()))
		} else {
			json.Unmarshal([]byte(src.String()), dst.Addr().Interface())
		}
	case reflect.Struct, reflect.Map:
		json.Unmarshal([]byte(src.String()), dst.Addr().Interface())
	default:
		reflectx.SetValue(dst, src)
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
	default:
		//err = fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
		return
	}
}

func StructValueToStruct(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	src = reflectx.Indirect(src, false)
	for i := 0; i < src.NumField(); i++ {
		srcField := src.Field(i)
		srcFieldStruct := src.Type().Field(i)
		if srcFieldStruct.Anonymous {
			dstField := dst.FieldByName(srcFieldStruct.Name)
			if !dstField.IsValid() {
				err = StructValueToStruct(dst, srcField)
			} else {
				if dstField.Kind() == srcField.Kind() && dstField.Kind() != reflect.Pointer {
					if !dstField.CanSet() {
						continue
					}
					err = reflectx.SetValue(dst, src)
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
			err = SimpleMapValue(dstField, srcField)
			if err != nil {
				return
			}
		}
	}
	return
}

func SliceValueTo(dst, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Slice:
		return SliceValueToSlice(dst, src)
	case reflect.String:
		return StructValueToString(dst, src)
	default:
		return fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
	}
}

func SliceValueToSlice(dst reflect.Value, src reflect.Value) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)
	dstItemType := dst.Type().Elem()
	if dst.IsNil() {
		dst.Set(reflect.MakeSlice(dst.Type(), 0, src.Len()))
	}
	for i := 0; i < src.Len(); i++ {
		var dstItem reflect.Value
		if dst.Len()-1 >= i {
			dstItem = dst.Index(i)
			err = SimpleMapValue(dstItem.Elem(), src.Index(i))
			if err != nil {
				return
			}
		} else {
			dstItem = reflect.New(dstItemType)
			err = SimpleMapValue(dstItem.Elem(), src.Index(i))
			if err != nil {
				return
			}
			dst.Set(reflect.Append(dst, dstItem.Elem()))
		}

	}
	return
}

func StructValueToString(dst reflect.Value, src reflect.Value) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)

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
