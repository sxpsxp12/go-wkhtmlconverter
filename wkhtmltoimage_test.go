package wkhtmltopdf

import (
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func newTestImageGenerator(tb testing.TB) *ImageGenerator {
	pImage, err := NewImageGenerator()
	if err != nil {
		tb.Fatal(err)
	}
	pImage.Format.Set("png")
	pImage.Height.Set(100)
	pImage.Width.Set(100)
	pImage.Quality.Set(100)

	return pImage
}

func expectedImageArgString() string {
	return "--format png --height 100 --width 100 --quality 100 --quiet - -"
}

func TestImageGenerator_Args(t *testing.T) {
	pImg := newTestImageGenerator(t)
	assert.Equal(t, expectedImageArgString(), pImg.ArgString())
	log.Println(pImg.ArgString())
}

func TestImageGenerator_WriteFileFromFile(t *testing.T) {
	pImg, err := NewImageGenerator()
	if err != nil {
		t.Fatal(err)
	}
	err = pImg.CreateFromFile("testdata/htmlsimple.html")
	if err != nil {
		t.Fatal(err)
	}
	err = pImg.WriteFile("testdata/file_htmlsimple.png")
	if err != nil {
		t.Fatal(err)
	}
}

func TestImageGenerator_WriteFileFromIO(t *testing.T) {
	pImg, err := NewImageGenerator()
	if err != nil {
		t.Fatal(err)
	}
	f, err := os.OpenFile("testdata/htmlsimple.html", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}

	err = pImg.CreateFromIOReader(f)
	if err != nil {
		t.Fatal(err)
	}
	err = pImg.WriteFile("testdata/io_htmlsimple.png")
	if err != nil {
		t.Fatal(err)
	}
}
