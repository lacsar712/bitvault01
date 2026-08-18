package slots

import "testing"

func TestAccept(t *testing.T) {
	if err := Enforce("img", "slot=a counter=2", []string{"gw"}); err != nil {
		t.Fatal(err)
	}
}

func TestRejectSlot(t *testing.T) {
	if err := Enforce("img", "slot=c counter=2", []string{"gw"}); err == nil {
		t.Fatal("expected reject")
	}
}
