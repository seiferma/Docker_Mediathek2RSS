package main

import "testing"

func TestIsValidShowID(t *testing.T) {
	assertIsValidArdShowID(t, "", false)
	assertIsValidArdShowID(t, "a", true)
	assertIsValidArdShowID(t, "a1", true)
	assertIsValidArdShowID(t, "b2H", true)
	assertIsValidArdShowID(t, "b2H?", false)
	assertIsValidArdShowID(t, "a\\q", false)
}

func assertIsValidArdShowID(t *testing.T, idToTest string, expectedResult bool) {
	actual := isValidArdShowID(idToTest)
	if actual != expectedResult {
		t.Fatalf("Expected the validation of %v to return %v but got %v.", idToTest, expectedResult, actual)
	}
}
