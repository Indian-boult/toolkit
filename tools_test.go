package toolkit

import "testing"

// TestTools_RandomString tests the RandomString function of the Tools struct.
//
// It checks that the function returns a random string of the specified length.
// The length of the returned string should match the specified length.
func TestTools_RandomString(t *testing.T) {
	var testTools Tools

	s := testTools.RandomString(10)
	if len(s) != 10 {
		t.Error("wrong length random string returned")
	}
}
