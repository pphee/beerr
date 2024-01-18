package utils

import (
	"reflect"
	"testing"
)

func TestBinaryConverter(t *testing.T) {
	tests := []struct {
		name   string
		number int
		bits   int
		want   []int
	}{
		{"Zero", 0, 4, []int{0, 0, 0, 0}},
		{"One", 1, 4, []int{0, 0, 0, 1}},
		{"Two", 2, 4, []int{0, 0, 1, 0}},
		{"Eight", 8, 4, []int{1, 0, 0, 0}},
		{"Fifteen", 15, 4, []int{1, 1, 1, 1}},
		{"LargeNumber", 123, 8, []int{0, 1, 1, 1, 1, 0, 1, 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BinaryConverter(tt.number, tt.bits); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BinaryConverter(%v, %v) = %v, want %v", tt.number, tt.bits, got, tt.want)
			}
		})
	}
}
