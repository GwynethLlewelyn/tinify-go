// Test suite for auxiiary functions.

package main

import (
	"testing"
)

var tests = []struct {
	hex   string
	check bool
}{
	{"#000000", true},
	{"00FF", true},
	{"0#0", false},
	{"", false},
	{"#", false},
	{"#F", false},
	{"B", false},
	{"abcdef01", true},
	{"#abcdef01", true},
	{"#abcdef0", false},
	{"#abcdefg0", false},
	{"#0f0", false},
	{"0x0F", false},
	{"10000", false},
}

// Given a series of what we consider to be valid colours in hexadecimal, test if
// we got the expected results.
func TestIsValidHex(t *testing.T) {
	for _, tc := range tests {
		if isValidHex(tc.hex) != tc.check {
			t.Fatalf("checked %q if it was valid hex or not and expected %t", tc.hex, tc.check)
		}
	}
}

func BenchmarkIsValidHex(b *testing.B) {
	for b.Loop() {
		for _, tc := range tests {
			if isValidHex(tc.hex) != tc.check {
				b.Fatalf("checked %q if it was valid hex or not and expected %t", tc.hex, tc.check)
			}
		}
	}
}

/*
func BenchmarkIsValidHexChatGPT(b *testing.B) {
	for b.Loop() {
		for _, tc := range tests {
			if isValidHexChatGPT(tc.hex) != tc.check {
				b.Fatalf("checked %q if it was valid hex or not and expected %t", tc.hex, tc.check)
			}
		}
	}
} */
