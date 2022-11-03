package webscenario

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/color/palette"
	"image/gif"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

const (
	SourceWidth = 600
	LineHeight  = 20
)

var (
	NoRecord = errors.New("no record")
)

var (
	Palette = append(palette.WebSafe, color.Transparent)
)

type recorderTask struct {
	Where      string
	Screenshot *[]byte
}

type Recorder struct {
	images []*image.Paletted
	ch     chan<- recorderTask
	stop   context.CancelFunc

	width, height int

	Done chan struct{}
}

func NewRecorder(ctx context.Context, width, height int) *Recorder {
	ch := make(chan recorderTask, 8)

	ctx, cancel := context.WithCancel(ctx)

	rec := &Recorder{
		ch:     ch,
		stop:   cancel,
		width:  width,
		height: height,
		Done:   make(chan struct{}),
	}

	go rec.runRecorder(ch, width, height)
	go func() {
		<-ctx.Done()
		close(ch)
	}()

	return rec
}

func parseWhere(where string) (string, int) {
	where = where[:len(where)-1]
	pos := strings.LastIndexByte(where, ':')
	line, err := strconv.Atoi(where[pos+1:])
	if err != nil {
		return where[:pos], 0
	}
	return where[:pos], line
}

func (r *Recorder) runRecorder(ch <-chan recorderTask, width, height int) {
	screenSize := image.Rect(0, 0, width, height)
	recordSize := image.Rect(0, 0, width+SourceWidth, height)

	for task := range ch {
		orig, err := png.Decode(bytes.NewReader(*task.Screenshot))
		if err != nil {
			// TODO: add error handling
			continue
		}

		img := image.NewPaletted(recordSize, Palette)
		draw.FloydSteinberg.Draw(img, screenSize, orig, image.ZP)

		where, line := parseWhere(task.Where)
		sourceImager.LoadAsImage(img, image.Rect(width, 0, recordSize.Max.X, height), where, line)

		r.images = append(r.images, img)
	}
	close(r.Done)
}

type RecordAction struct {
	ch   chan<- recorderTask
	task recorderTask
}

func (a RecordAction) Do(ctx context.Context) error {
	a.ch <- a.task
	return nil
}

func (r *Recorder) Record(where string, screenshot *[]byte) RecordAction {
	return RecordAction{
		ch: r.ch,
		task: recorderTask{
			Where:      where,
			Screenshot: screenshot,
		},
	}
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

func copyImage(dst, src *image.Paletted, rect image.Rectangle, offset image.Point) {
	for y := rect.Min.Y; y < rect.Max.Y; y++ {
		for x := rect.Min.X; x < rect.Max.X; x++ {
			dst.SetColorIndex(x+offset.X, y+offset.Y, src.ColorIndexAt(x, y))
		}
	}
}

func (r *Recorder) SaveTo(f io.Writer) error {
	if len(r.images) == 0 {
		return NoRecord
	}

	compressGif(r.images)

	g := gif.GIF{
		Image: r.images,
	}
	for range g.Image {
		g.Delay = append(g.Delay, 200)
	}
	g.Delay[len(g.Delay)-1] = 400

	return gif.EncodeAll(f, &g)
}

type SourceImager struct {
	face    font.Face
	sources map[string][]string
}

var (
	sourceImager *SourceImager
)

func init() {
	si, err := NewSourceImager()
	if err != nil {
		panic(err)
	}
	sourceImager = si
}

func NewSourceImager() (*SourceImager, error) {
	ft, err := opentype.Parse(goregular.TTF)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{
		Size: 16,
		DPI:  72,
	})

	return &SourceImager{
		face:    face,
		sources: make(map[string][]string),
	}, nil
}

func (s *SourceImager) Load(path string) ([]string, error) {
	if xs, ok := s.sources[path]; ok {
		return xs, nil
	}
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	xs := strings.Split(strings.ReplaceAll(string(f), "\r", ""), "\n")
	s.sources[path] = xs
	return xs, nil
}

func (s *SourceImager) LoadAsImage(img *image.Paletted, rect image.Rectangle, path string, line int) {
	lines, err := s.Load(path)
	if err != nil {
		return
	}

	drawer := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: s.face,
	}

	offset := line + 3 - rect.Bounds().Max.Y/LineHeight
	if offset < 1 {
		offset = 1
	}

	for i, l := range lines[offset-1:] {
		if i > rect.Bounds().Max.Y/LineHeight {
			break
		}

		drawer.Dot.X = fixed.I(rect.Min.X + LineHeight/2)
		drawer.Dot.Y = fixed.I(rect.Min.Y + LineHeight + i*LineHeight)

		if line == i+offset {
			b, _ := drawer.BoundString(l)
			draw.Draw(img, image.Rect(b.Min.X.Round()-4, b.Min.Y.Round()-2, b.Max.X.Round()+4, b.Max.Y.Round()+2), &image.Uniform{image.White}, image.ZP, draw.Src)
			drawer.Src = image.Black
		} else {
			drawer.Src = image.White
		}

		drawer.DrawString(l)
	}

	return
}
