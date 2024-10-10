package mapx

import (
	"net/http"
	"reflect"
	"testing"
)

func TestMerge(t *testing.T) {
	// Test case with two empty maps
	a := make(map[string]int)
	b := make(map[string]int)
	expected := make(map[string]int)
	result := Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}

	// Test case with one map having some elements
	a = map[string]int{"a": 1, "b": 2}
	b = make(map[string]int)
	expected = map[string]int{"a": 1, "b": 2}
	result = Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}

	// Test case with both maps having some overlapping elements
	a = map[string]int{"a": 1, "b": 2}
	b = map[string]int{"b": 3, "c": 4}
	expected = map[string]int{"a": 1, "b": 3, "c": 4}
	result = Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}
}

func TestMerge2(t *testing.T) {
	// Test case with two empty maps
	a := make(http.Header)
	b := make(http.Header)
	expected := make(http.Header)
	result := Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}

	// Test case with one map having some elements
	a = http.Header{"a": {"b", "c"}}
	b = make(http.Header)
	expected = http.Header{"a": {"b", "c"}}
	result = Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}

	// Test case with both maps having some overlapping elements
	a = http.Header{"a": {"b", "c"}}
	b = http.Header{"b": {"b", "c"}}
	expected = http.Header{"a": {"b", "c"}, "b": {"b", "c"}}
	result = Merge(a, b)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("Merge(%v, %v) = %v, expected %v", a, b, result, expected)
	}
}
