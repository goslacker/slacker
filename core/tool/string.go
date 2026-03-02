package tool

import (
	"math/rand"
)

// 字符串类型常量
const (
	// 纯字母
	typeAlpha = 1
	// 纯数字
	typeNumeric = 2
	// 字母+数字
	typeAlphaNumeric = 3
	// 字母+数字+符号
	typeAll = 4
)

type RandomStringOption struct {
	// 字符串类型，可选值：typeAlpha, typeNumeric, typeAlphaNumeric, typeAll
	StrType int
}

// 字符集
const (
	lowercase = "abcdefghijklmnopqrstuvwxyz"
	uppercase = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	numeric   = "0123456789"
	symbols   = "!@#$%^&*()_+-=[]{}|;':,.<>?"
)

// RandomStringWithTypeNumeric 设置随机字符串类型为纯数字
func RandomStringWithTypeNumeric() func(*RandomStringOption) {
	return func(opt *RandomStringOption) {
		opt.StrType = typeNumeric
	}
}

// RandomStringWithTypeAlphaNumeric 设置随机字符串类型为字母+数字
func RandomStringWithTypeAlphaNumeric() func(*RandomStringOption) {
	return func(opt *RandomStringOption) {
		opt.StrType = typeAlphaNumeric
	}
}

// RandomStringWithTypeAll 设置随机字符串类型为字母+数字+符号
func RandomStringWithTypeAll() func(*RandomStringOption) {
	return func(opt *RandomStringOption) {
		opt.StrType = typeAll
	}
}

// GenerateRandomString 生成指定长度和类型的随机字符串
// length: 字符串长度
func RandomString(length int, opts ...func(*RandomStringOption)) string {
	if length <= 0 {
		return ""
	}

	opt := &RandomStringOption{
		StrType: typeAlpha,
	}
	for _, o := range opts {
		o(opt)
	}

	var charset string

	switch opt.StrType {
	case typeAlpha:
		charset = lowercase + uppercase
	case typeNumeric:
		charset = numeric
	case typeAlphaNumeric:
		charset = lowercase + uppercase + numeric
	case typeAll:
		charset = lowercase + uppercase + numeric + symbols
	default:
		charset = lowercase + uppercase + numeric // 默认字母+数字
	}

	result := make([]byte, length)
	charsetLength := len(charset)

	for i := 0; i < length; i++ {
		result[i] = charset[rand.Intn(charsetLength)]
	}

	return string(result)
}
