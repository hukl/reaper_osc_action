package main

import "testing"

func Test_padString(t *testing.T) {
	testCases := []struct {
		input string
	}{
		{"action"},
		{"4077"},
		{"_S&M_INS_MARKER_PLAY"},
		{"streamdeck"},
		{"osc"},
		{"a"},
		{"abc"},
		{"1234567890"},
	}

	for _, tc := range testCases {
		result := padString(tc.input)
		// Check if the length of the byte slice is divisible by 4
		if len(result) % 4 != 0 {
			t.Errorf("Length of padded string for input %q is %d; expected to be divisible by 4", tc.input, len(result))
		}
	}
}


func Test_createOSCPacket(t *testing.T) {
	testCases := []struct {
		address  string
		argument string
	}{
		{"/action", "PLAY"},
		{"/action", "PLA"},
		{"/action", "PL"},
		{"/action", "P"},
		{"/action", "LONGER_STRING"},
		{"/action", "_S&M_INS_MARKER_PLAY"},
		{"/action", "40961"},
		{"/action", "123"},
		{"/action", "__WADDEHADDEDUDEDA__"},
	}

	for _, tc := range testCases {
		result := createOSCPacket(tc.address, tc.argument)

		if len(result) % 4 != 0 {
			t.Errorf("Length of padded string for input is %d; expected to be divisible by 4", len(result))
		}
	}

}
