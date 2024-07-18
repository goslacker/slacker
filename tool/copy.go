package tool

import (
	"fmt"
	"github.com/goslacker/slacker/extend/reflectx"
	"reflect"
)

func SimpleMap(dst any, src any) (err error) {
	return SimpleMapValue(reflect.ValueOf(dst), reflect.ValueOf(src))
}

func SimpleMapValue(dst reflect.Value, src reflect.Value) (err error) {
	if src.IsZero() {
		return
	}
	src = reflectx.Indirect(src, false)
	switch src.Kind() {
	case reflect.Struct:
		return StructValueTo(dst, src)
	case reflect.Slice:
		return SliceValueTo(dst, src)
	default:
		return reflectx.SetValue(dst, src)
	}
}

func StructValueTo(dst reflect.Value, src reflect.Value) (err error) {
	dst = reflectx.Indirect(dst, true)
	switch dst.Kind() {
	case reflect.Struct:
		return StructValueToStruct(dst, src)
	default:
		err = fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
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
				err = StructValueToStruct(dstField, srcField)
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
	default:
		return fmt.Errorf("unsupported src type <%s> to dst type <%s>", src.Type().String(), dst.Type().String())
	}
}

func SliceValueToSlice(dst reflect.Value, src reflect.Value) (err error) {
	src = reflectx.Indirect(src, false)
	dst = reflectx.Indirect(dst, false)
	dstItemType := dst.Type().Elem()
	dst.Set(reflect.MakeSlice(dst.Type(), 0, src.Len()))
	for i := 0; i < src.Len(); i++ {
		dstItem := reflect.New(dstItemType)
		err = SimpleMapValue(dstItem.Elem(), src.Index(i))
		if err != nil {
			return
		}
		dst.Set(reflect.Append(dst, dstItem.Elem()))
	}
	return
}
