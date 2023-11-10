package main

import (
	"testing"

//	"io/ioutil"
	"os"

	"github.com/gwpp/tinify-go/tinify"
)

const Key = "rcPZm3Zrg_1DbjYtV6AXM_-53Jg9wuWB"

func TestCompressFromFile(t *testing.T) {
	Tinify.SetKey(Key)
	source, err := Tinify.FromFile("./test.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	err = source.ToFile("./test_output/CompressFromFile.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Compress successful")
}

func TestCompressFromBuffer(t *testing.T) {
	Tinify.SetKey(Key)

	buf, err := os.ReadFile("./test.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	source, err := Tinify.FromBuffer(buf)
	if err != nil {
		t.Error(err)
		return
	}

	err = source.ToFile("./test_output/CompressFromBuffer.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Compress successful")
}

func TestCompressFromUrl(t *testing.T) {
	Tinify.SetKey(Key)
	url := "http://pic.tugou.com/realcase/1481255483_7311782.jpg"
	source, err := Tinify.FromUrl(url)
	if err != nil {
		t.Error(err)
		return
	}
	err = source.ToFile("./test_output/CompressFromUrl.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Compress successful")
}

func TestResizeFromFile(t *testing.T) {
	Tinify.SetKey(Key)
	source, err := Tinify.FromFile("./test.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	err = source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethodFit,
		Width:  100,
		Height: 100,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = source.ToFile("./test_output/ResizeFromFile.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	t.Log("Resize successful")
}

func TestResizeFromBuffer(t *testing.T) {
	Tinify.SetKey(Key)

	buf, err := os.ReadFile("./test.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	source, err := Tinify.FromBuffer(buf)
	if err != nil {
		t.Error(err)
		return
	}

	err = source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethodScale,
		Width:  200,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = source.ToFile("./test_output/ResizesFromBuffer.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Resize successful")
}

func TestResizeFromUrl(t *testing.T) {
	Tinify.SetKey(Key)
	url := "http://pic.tugou.com/realcase/1481255483_7311782.jpg"
	source, err := Tinify.FromUrl(url)
	if err != nil {
		t.Error(err)
		return
	}

	err = source.Resize(&Tinify.ResizeOption{
		Method: Tinify.ResizeMethodCover,
		Width:  300,
		Height: 100,
	})
	if err != nil {
		t.Error(err)
		return
	}

	err = source.ToFile("./test_output/ResizeFromUrl.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Log("Resize successful")
}
