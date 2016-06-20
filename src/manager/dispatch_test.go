package manager

import (
	"fmt"
	"testing"
)

func TestIp(t *testing.T) {
	i, err := NewIpDatabase()
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	subnet, err := i.GetSubNet("36.110.128.35")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	fmt.Println(subnet)
}
