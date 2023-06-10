package bard

import "testing"

func TestNewBardClient(t *testing.T) {
	bard := NewBard("")
	rsp, err := bard.SendMessage("Hello")
	if err != nil {
		t.Errorf(err.Error())
	}
	t.Logf("%+v", rsp.Choices[0].Answer)

}
