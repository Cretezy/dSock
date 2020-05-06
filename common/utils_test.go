// Simple tests in here for simple functions
package common_test

import (
	"github.com/Cretezy/dSock/common"
	"testing"
)

func TestRemoveString(t *testing.T) {
	removed := common.RemoveString([]string{"a", "b", "a"}, "a")
	if len(removed) != 2 || removed[0] != "b" || removed[1] != "a" {
		t.Fatal()
	}
}

func TestUniqueString(t *testing.T) {
	unique := common.UniqueString([]string{"a", "b", "a"})
	if len(unique) != 2 || unique[0] != "a" || unique[1] != "b" {
		t.Fatal()
	}
}

func TestRemoveEmpty(t *testing.T) {
	removed := common.RemoveEmpty([]string{"a", "b", ""})
	if len(removed) != 2 || removed[0] != "a" || removed[1] != "b" {
		t.Fatal()
	}
}

func TestRandomString(t *testing.T) {
	random := common.RandomString(8)
	if len(random) != 8 {
		t.Fatal()
	}
}
