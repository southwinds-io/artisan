package data

import (
	"fmt"
	"testing"
)

func TestLoadBuildFile(t *testing.T) {
	bf, err := LoadBuildFile("test/build.yaml")
	if err != nil {
		t.Fatal(err)
	}
	t1 := bf.Fx("test1")
	if t1 != nil {
		fmt.Println(t1.Name)
	}
}
