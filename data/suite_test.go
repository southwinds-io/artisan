package data

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"os"
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

func TestIPGroups(t *testing.T) {
	b, err := os.ReadFile("net.yaml")
	if err != nil {
		t.Fatal(err)
	}
	n := &Network{}
	err = yaml.Unmarshal(b, n)
	if err != nil {
		t.Fatal(err)
	}
	m, err := n.AllocateIPs("172.16.82.132", "172.16.82.133")
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range m {
		fmt.Printf("%v\n", info)
	}
}
