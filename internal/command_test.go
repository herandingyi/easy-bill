package internal

import (
	"fmt"
	"testing"
)

func TestToAaCmd(t *testing.T) {
	cmdStr := "/a tk 100,hcx u"
	cmd, err := ToAaCmd(cmdStr)
	if err != nil {
		t.Fatal(err)
		return
	}
	if fmt.Sprint(cmd.Names) != "[tk hcx]" {
		t.Fatal("names not match")
	}
}
