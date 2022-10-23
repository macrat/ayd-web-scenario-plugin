package main

import (
	"bytes"
	"os"
	"testing"

	"image"
	"image/draw"
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
	images := []*image.Paletted{
		LoadGif(t, "testdata/gif/raw/0.gif"),
		LoadGif(t, "testdata/gif/raw/1.gif"),
		LoadGif(t, "testdata/gif/raw/2.gif"),
		LoadGif(t, "testdata/gif/raw/3.gif"),
	}
	wants := []*image.Paletted{
		LoadGif(t, "testdata/gif/compressed/0.gif"),
		LoadGif(t, "testdata/gif/compressed/1.gif"),
		LoadGif(t, "testdata/gif/compressed/2.gif"),
		LoadGif(t, "testdata/gif/compressed/3.gif"),
	}

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

func Benchmark_compressGif(b *testing.B) {
	images := []*image.Paletted{
		LoadGif(b, "testdata/gif/raw/0.gif"),
		LoadGif(b, "testdata/gif/raw/1.gif"),
		LoadGif(b, "testdata/gif/raw/2.gif"),
		LoadGif(b, "testdata/gif/raw/3.gif"),
	}

	var target []*image.Paletted
	for _, img := range images {
		target = append(target, image.NewPaletted(img.Bounds(), Palette))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		for i := range images {
			draw.Draw(target[i], images[i].Bounds(), images[i], image.ZP, draw.Over)
		}
		b.StartTimer()
		compressGif(target)
	}
}
