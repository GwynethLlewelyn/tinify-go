// Test suite for main functionality

package main

import (
	"os"
	"testing"

	"github.com/gwpp/tinify-go/tinify"
	_ "github.com/joho/godotenv/autoload"
)

func TestCompressFromFile(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))
	source, err := Tinify.FromFile("./testdata/input/test.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/CompressFromFile.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Compress successful, %d tokens left", tokens)
}

func TestCompressFromBuffer(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))

	buf, err := os.ReadFile("./testdata/input/test.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	source, err := Tinify.FromBuffer(buf)
	if err != nil {
		t.Error(err)
		return
	}
	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/CompressFromBuffer.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Compress successful, %d tokens left", tokens)
}

func TestCompressFromUrl(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))
	url := "http://pic.tugou.com/realcase/1481255483_7311782.jpg"
	source, err := Tinify.FromUrl(url)
	if err != nil {
		t.Error(err)
		return
	}
	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/CompressFromUrl.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Compress successful, %d tokens left", tokens)
}

func TestResizeFromFile(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))
	source, err := Tinify.FromFile("./testdata/input/test.jpg")
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

	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/ResizeFromFile.jpg")
	if err != nil {
		t.Error(err)
		return
	}

	t.Logf("Resize successful, %d tokens left", tokens)
}

func TestResizeFromBuffer(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))

	buf, err := os.ReadFile("./testdata/input/test.jpg")
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

	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/ResizesFromBuffer.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Resize successful, %d tokens left", tokens)
}

// This ests if we're using scale with both width and height set.
func TestResizeFromBufferScaleWidthAndHeight(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))

	buf, err := os.ReadFile("./testdata/input/test.jpg")
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
		Height: 256,
		Width:  128,
	})
	// inverse logic, this *must* fail!
	if err == nil {
		t.Error("Resize with scale cannot have both width and height set!")
		return
	}

	t.Log("Resize with scale using width and height both set was correctly flagged with error", err)
}

// This ests if we're using scale with both width and height set to zero.
func TestResizeFromBufferScaleBothDimensionsZero(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))

	buf, err := os.ReadFile("./testdata/input/test.jpg")
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
		/*		Height: 256,
				Width:  128,*/
	})
	// inverse logic, this *must* fail!
	if err == nil {
		t.Error("Resize with scale cannot have both width and height set to zero!")
		return
	}

	t.Log("Resize with scale using width and height both set to zero was correctly flagged with error", err)
}

func TestResizeFromUrl(t *testing.T) {
	Tinify.SetKey(os.Getenv("TINIFY_API_KEY"))
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

	var tokens int64
	tokens, err = source.ToFileC("./testdata/output/ResizeFromUrl.jpg")
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("Resize successful, %d tokens left", tokens)
}
