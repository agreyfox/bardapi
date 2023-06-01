package bingimg

import (
	"testing"
)

// TestNewBingImage is a function that tests the NewBingImage function
func TestNewBingImageClient(t *testing.T) {

	bi, _ := NewImageGen(Cookie_U, false, "test.log")
	img, err := bi.getImages("Film still of an elderly wise yellow man playing chess, medium shot, mid-shot")
	if err != nil {
		t.Errorf("error %s", err)
	}
	for _, item := range img {
		t.Logf(item)
	}
	bi.MakeThumbnail(true)
	bi.saveImages(img, ".", "img")
}
