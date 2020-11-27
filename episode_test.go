package main

import (
	"testing"
)

func TestCalcTier(t *testing.T) {
	vals := map[string]uint{
		"1080p.AMZN":   10,
		"1080p.PROPER": 9,
		"1080p.Atmos":  9,
		"1080p.who":    8,
		"720p.AMZN":    6,
		"720p.PROPER":  5,
		"720p.nice":    4,
		"480p.AMZN":    2,
		"480p.PROPER":  1,
		"480p.nice":    0,
	}

	for f, exp := range vals {
		got := calcTier(f)
		if got != exp {
			t.Errorf("calcTier(%s) = %d; want %d", f, got, exp)
		}
	}
}
