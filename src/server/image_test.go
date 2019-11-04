package main

import (
	"io/ioutil"
	"testing"

)

func TestImageResize(t *testing.T) {
	opts := ImageOptions{Width: 300, Height: 300}
	data:=Data{}
	buf, _ := ioutil.ReadAll(readFile("imaginary.jpg"))

	img, err := Resize(buf, opts,data)
	if err != nil {
		t.Errorf("Cannot process image: %s", err)
	}

	if img.Mime != "application/json" {
		t.Error("Invalid application/json MIME type")
	}

	if assertSize(img.Body, opts.Width, opts.Height) != nil {
		t.Errorf("Invalid image size, expected: %dx%d", opts.Width, opts.Height)
	}
}
