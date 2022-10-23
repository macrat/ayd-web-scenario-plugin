package main

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

var (
	Palette = append(palette.WebSafe, color.Transparent)

	debugCounter = 0
)

type recorderTask struct {
	Name       string
	IsAfter    bool
	Screenshot []byte
}

type Recorder struct {
	face   font.Face
	images []*image.Paletted
	ch     chan<- recorderTask
	stop   context.CancelFunc

	Done chan struct{}
}

func NewRecorder(ctx context.Context) (*Recorder, error) {
	ft, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size: 32,
		DPI:  72,
	})

	ch := make(chan recorderTask, 8)

	ctx, cancel := context.WithCancel(ctx)

	rec := &Recorder{
		face: face,
		ch:   ch,
		stop: cancel,
		Done: make(chan struct{}),
	}

	go rec.runRecorder(ch)
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return rec, nil
}

func (r *Recorder) runRecorder(ch <-chan recorderTask) {
	for task := range ch {
		scr, _, err := image.Decode(bytes.NewReader(task.Screenshot))
		if err != nil {
			// TODO: add error handling
			continue
		}
		bounds := scr.Bounds()
		bounds.Max.Y += 32

		img := image.NewPaletted(bounds, Palette)

		bounds.Min.Y += 32
		draw.FloydSteinberg.Draw(img, bounds, scr, image.ZP)

		drawer := &font.Drawer{
			Dst:  img,
			Src:  image.White,
			Face: r.face,
		}
		b, _ := drawer.BoundString(task.Name)
		drawer.Dot.Y = fixed.I(2) - b.Min.Y
		drawer.DrawString(task.Name)

		r.images = append(r.images, img)
	}
	close(r.Done)
}

func (r *Recorder) Close() error {
	r.stop()
	<-r.Done
	return nil
}

func (r *Recorder) RecordOnce(taskName string, isAfter bool, screenshot []byte) error {
	r.ch <- recorderTask{
		Name:       taskName,
		IsAfter:    isAfter,
		Screenshot: screenshot,
	}

	return nil
}

func (r *Recorder) RecordBoth(taskName string, before, after []byte) error {
	if err := r.RecordOnce(taskName, false, before); err != nil {
		return err
	}
	return r.RecordOnce(taskName, true, after)
}

func compressGif(images []*image.Paletted) {
	width := images[0].Rect.Max.X
	height := images[0].Rect.Max.Y

	for i := len(images) - 1; i > 0; i-- {
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				if images[i].ColorIndexAt(x, y) == images[i-1].ColorIndexAt(x, y) {
					images[i].SetColorIndex(x, y, uint8(len(Palette)-1))
				}
			}
		}
	}
}

func (r *Recorder) SaveTo(f io.Writer) error {
	compressGif(r.images)

	g := gif.GIF{
		Image: r.images,
	}
	for range g.Image {
		g.Delay = append(g.Delay, 100)
	}
	g.Delay[len(g.Delay)-1] = 500

	return gif.EncodeAll(f, &g)
}
