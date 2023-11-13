package main

import (
	"testing"

	"github.com/ursaserver/ursa"
)

func TestSanitizeRateString(t *testing.T) {
	type Test struct {
		rateString string
		expected   string
	}
	tests := []Test{
		{rateString: "3/minute", expected: "3/minute"},
		{rateString: "3/ minute", expected: "3/minute"},
		{rateString: " 3/ minut e", expected: "3/minute"},
	}
	for _, test := range tests {
		if got := SanitizeRateString(test.rateString); got != test.expected {
			t.Errorf("expected %v got %v for %v", test.expected, got, test.rateString)
		}
	}
}

func TestRateString(t *testing.T) {
	type Test struct {
		rateString string
		expected   ursa.Rate
		error      bool
	}
	tests := []Test{
		{rateString: "30 / minute", expected: ursa.NewRate(30, ursa.Minute)},
		{rateString: "1/hour", expected: ursa.NewRate(1, ursa.Hour)},
		{rateString: "1/day", expected: ursa.NewRate(1, ursa.Day)},
		{rateString: "-1/day", error: true},                          // No negative capacity should be allowed
		{rateString: "0/day", error: true},                           // Capacity has to be > 0
		{rateString: "1/second", error: true},                        // Second as a unit isn't allowed
		{rateString: "1.5/Minute", error: true},                      // No float should be allowed
		{rateString: "5/HOUR", expected: ursa.NewRate(5, ursa.Hour)}, // Any case for unit of time is allowed
		{rateString: "5/hOUr", expected: ursa.NewRate(5, ursa.Hour)}, // Mixed casing should be allowed
	}
	for _, test := range tests {
		rate, err := RateStringToRate(test.rateString)
		gotError := err != nil
		expectedError := test.error
		if gotError != expectedError {
			t.Errorf("expected error %v got error %v for rate string %v", expectedError, gotError, test.rateString)
		} else if rate != test.expected {
			t.Errorf("expected ratea %v got %v", test.expected, rate)
		}
	}
}
