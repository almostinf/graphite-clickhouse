package load_avg

import (
	"fmt"
	"testing"
)

func TestWeight(t *testing.T) {
	tests := []struct {
		n    int
		l    float64
		want int64
	}{
		// n : 100
		{n: 100, l: 0, want: 200},
		{n: 100, l: 0.1, want: 199},
		{n: 100, l: 0.11, want: 199},
		{n: 100, l: 0.2, want: 169},
		{n: 100, l: 0.5, want: 130},
		{n: 100, l: 0.9, want: 104},
		{n: 100, l: 1, want: 100},
		{n: 100, l: 1.1, want: 36},
		{n: 100, l: 1.9, want: 12},
		{n: 100, l: 2, want: 1},
		{n: 100, l: 9, want: 1},
		{n: 100, l: 10, want: 1},
		{n: 100, l: 20, want: 1},
		// n : 1000
		{n: 1000, l: 0, want: 2000},
		{n: 1000, l: 0.1, want: 1999},
		{n: 1000, l: 0.11, want: 1999},
		{n: 1000, l: 0.2, want: 1698},
		{n: 1000, l: 0.5, want: 1301},
		{n: 1000, l: 0.9, want: 1045},
		{n: 1000, l: 1, want: 1000},
		{n: 1000, l: 1.1, want: 357},
		{n: 1000, l: 1.9, want: 120},
		{n: 1000, l: 2, want: 1},
		{n: 1000, l: 3, want: 1},
		{n: 1000, l: 4, want: 1},
		{n: 1000, l: 9, want: 1},
		{n: 1000, l: 10, want: 1},
		{n: 1000, l: 20, want: 1},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("%d#%f", tt.n, tt.l), func(t *testing.T) {
			if got := Weight(tt.n, tt.l); got != tt.want {
				t.Errorf("Weight(%d, %f) = %v, want %v", tt.n, tt.l, got, tt.want)
			}
		})
	}
}
