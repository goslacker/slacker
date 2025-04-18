package slogx

import (
	"log/slog"
	"reflect"
)

var replacers map[any]func(groups []string, a slog.Attr) slog.Attr

func RegisterValueBasedReplacer(value any, replacer func(groups []string, a slog.Attr) slog.Attr) {
	if replacers == nil {
		replacers = make(map[any]func(groups []string, a slog.Attr) slog.Attr)
	}
	replacers[value] = replacer
}

func Replacer(groups []string, a slog.Attr) slog.Attr {
	for k, v := range replacers {
		if reflect.TypeOf(a.Value.Any()) == reflect.TypeOf(k) {
			return v(groups, a)
		}
	}
	return a
}
