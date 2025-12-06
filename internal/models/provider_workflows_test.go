package models

import (
	"testing"
)

type TestStruct struct{}

func (t *TestStruct) TestMethod() {}

func TestGetNameFromFunction(t *testing.T) {
	ts := &TestStruct{}
	// Test Method Value (fallback path)
	name := GetNameFromFunction(ts.TestMethod)
	if name != "TestMethod" {
		t.Errorf("Method Value: expected TestMethod, got %s", name)
	}

	// Test Method Expression (reflection path)
	name = GetNameFromFunction((*TestStruct).TestMethod)
	if name != "TestMethod" {
		t.Errorf("Method Expression: expected TestMethod, got %s", name)
	}
}
