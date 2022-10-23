package main

import (
	"bytes"
	"fmt"
	"os"
	"testing"

	"image"
	"image/gif"
)

func LoadGif(t testing.TB, name string) *image.Paletted {
	t.Helper()

	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("failed to open input gif: %s", err)
	}

	img, err := gif.Decode(f)
	if err != nil {
		t.Fatalf("failed to decode input git: %s", err)
	}

	return img.(*image.Paletted)
}

func Test_compressGif(t *testing.T) {
	loadGifs := func(path string) []*image.Paletted {
		var imgs []*image.Paletted
		for i := 0; i <= 6; i++ {
			imgs = append(imgs, LoadGif(t, fmt.Sprintf("%s/%d.gif", path, i)))
		}
		return imgs
	}

	images := loadGifs("testdata/gif/raw")
	wants := loadGifs("testdata/gif/compressed")

	compressGif(images)

	var want, actual bytes.Buffer

	for i := range images {
		//f, _ := os.Create(fmt.Sprintf("testdata/gif/compressed/%d.gif", i))
		//gif.Encode(f, images[i], nil)

		if err := gif.Encode(&want, wants[i], nil); err != nil {
			t.Errorf("failed to encode want image[%d]: %s", i, err)
			continue
		}
		if err := gif.Encode(&actual, images[i], nil); err != nil {
			t.Errorf("failed to encode actual image[%d]: %s", i, err)
			continue
		}

		if !bytes.Equal(want.Bytes(), actual.Bytes()) {
			t.Errorf("unexpected image[%d]", i)
		}
	}
}
