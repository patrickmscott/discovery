package discovery

import (
	"testing"
)

func TestStrCmp(t *testing.T) {
	if strcmp("", "") != 0 {
		t.Error("empty strings should be equal")
	}
	if strcmp("a", "b") != -1 {
		t.Error("a - b should be -1")
	}
	if strcmp("a", "aa") != -1 {
		t.Error("a - aa should be -1")
	}
	if strcmp("bb", "b") != 1 {
		t.Error("bb - b should be 1")
	}
}
