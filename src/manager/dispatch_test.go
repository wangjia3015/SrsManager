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
	subnet, err := i.GetSubNet("111.204.243.7")
	if err != nil {
		t.Log(err)
		t.FailNow()
	}
	fmt.Println(subnet)
}
