package stats

import (
	"strconv"
	"testing"
)

func TestRecordResultCategories(t *testing.T) {
	type counts struct {
		green, red, black, even, odd, low, high int
	}
	tests := []struct {
		num  string
		want counts
	}{
		{"0", counts{green: 1}},
		{"00", counts{green: 1}},
		{"1", counts{red: 1, odd: 1, low: 1}},
		{"2", counts{black: 1, even: 1, low: 1}},
		{"18", counts{red: 1, even: 1, low: 1}},
		{"19", counts{red: 1, odd: 1, high: 1}},
		{"28", counts{black: 1, even: 1, high: 1}},
		{"35", counts{black: 1, odd: 1, high: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.num, func(t *testing.T) {
			s := New(0, 0, 0, 0, 0, 0, 0, 0)
			s.RecordResult(tt.num)
			got := counts{s.GreenCount, s.RedCount, s.BlackCount, s.EvenCount, s.OddCount, s.LowCount, s.HighCount}
			if got != tt.want {
				t.Errorf("counts = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestRecordResultTotalsAcrossWheel(t *testing.T) {
	s := New(0, 0, 0, 0, 0, 0, 0, 0)
	s.RecordResult("0")
	s.RecordResult("00")
	for n := 1; n <= 36; n++ {
		s.RecordResult(strconv.Itoa(n))
	}

	if s.TotalSpins != 38 {
		t.Fatalf("TotalSpins = %d, want 38", s.TotalSpins)
	}
	checks := []struct {
		name string
		got  int
		want int
	}{
		{"green", s.GreenCount, 2},
		{"red", s.RedCount, 18},
		{"black", s.BlackCount, 18},
		{"even", s.EvenCount, 18},
		{"odd", s.OddCount, 18},
		{"low", s.LowCount, 18},
		{"high", s.HighCount, 18},
	}
	for _, c := range checks {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
}
