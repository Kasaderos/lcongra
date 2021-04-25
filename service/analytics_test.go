package service

import (
	"fmt"
	"testing"
)

func Test_getDirection(t *testing.T) {
	tests := []struct {
		name string
		want Direction
	}{
		{"test1", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDirection("", "")
			fmt.Println(got)
		})
	}
}
