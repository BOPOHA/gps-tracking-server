package tools

import (
	"reflect"
	"testing"
)

func AssertEqual(t *testing.T, expected interface{}, fact interface{}) {
	if !reflect.DeepEqual(expected, fact) {
		t.Fatalf("Assert error, expected:\n%+v\nfact:\n%+v", expected, fact)
	}
}
