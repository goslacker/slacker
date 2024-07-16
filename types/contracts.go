package types

type StrOrNumber interface {
	~string |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~float64 | ~float32 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}
