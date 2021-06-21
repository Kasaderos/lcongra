package service

import (
	"fmt"
	"testing"
	"time"
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

func TestSignals_Flush(t *testing.T) {
	tests := []struct {
		name string
		s    Signals
	}{
		{"1", Signals{
			Signal{
				Stay, time.Now(),
			},
			Signal{
				Down, time.Now(),
			},
		},
		},
		{"2", Signals{
			Signal{
				Up, time.Now(),
			},
			Signal{
				Up, time.Now(),
			},
		},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.s.Flush()
		})
	}
}
