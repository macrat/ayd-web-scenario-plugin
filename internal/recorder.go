package webscenario

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
	SourceWidth  = 1280 - 1280*480/720
	RecordHeight = 480
	LineHeight   = 20
)

var (
	Palette = append(palette.WebSafe, color.Transparent)
)

type recorderTask struct {
	Where      string
	Name       string
	IsAfter    bool
	Screenshot *[]byte
}

type Recorder struct {
	images []*image.Paletted
	ch     chan<- recorderTask
	stop   context.CancelFunc

	Done chan struct{}
}

func NewRecorder(ctx context.Context) *Recorder {
	ch := make(chan recorderTask, 8)

	ctx, cancel := context.WithCancel(ctx)

	rec := &Recorder{
		ch:   ch,
		stop: cancel,
		Done: make(chan struct{}),
	}

	go rec.runRecorder(ch)
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

func (r *Recorder) runRecorder(ch <-chan recorderTask) {
	buf := image.NewRGBA(image.Rect(0, 0, 900, RecordHeight))

	for task := range ch {
		orig, _, err := image.Decode(bytes.NewReader(*task.Screenshot))
		if err != nil {
			// TODO: add error handling
			continue
		}
		size := orig.Bounds().Max

		bounds := image.Rect(0, 0, size.X*RecordHeight/size.Y, RecordHeight)
		if buf.Bounds().Max.X < bounds.Max.X {
			buf = image.NewRGBA(bounds)
		}
		draw.Draw(buf, bounds, orig, image.ZP, draw.Src)
		draw.ApproxBiLinear.Scale(buf, bounds, orig, orig.Bounds(), draw.Src, nil)

		img := image.NewPaletted(image.Rect(0, 0, bounds.Max.X+SourceWidth, RecordHeight), Palette)
		draw.FloydSteinberg.Draw(img, bounds, buf, image.ZP)

		where, line := parseWhere(task.Where)
		sourceImager.LoadAsImage(img, image.Rect(bounds.Max.X, 0, img.Bounds().Max.X, RecordHeight), where, line)

		r.images = append(r.images, img)
	}
	close(r.Done)
}

func (r *Recorder) Close() error {
	r.stop()
	<-r.Done
	return nil
}

type RecordAction struct {
	ch   chan<- recorderTask
	task recorderTask
}

func (a RecordAction) Do(ctx context.Context) error {
	a.ch <- a.task
	return nil
}

func (r *Recorder) Record(where, taskName string, isAfter bool, screenshot *[]byte) RecordAction {
	return RecordAction{
		ch: r.ch,
		task: recorderTask{
			Where:      where,
			Name:       taskName,
			IsAfter:    isAfter,
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
	maxWidth := 0
	for _, img := range r.images {
		x := img.Bounds().Max.X - SourceWidth
		if maxWidth < x {
			maxWidth = x
		}
	}

	var images []*image.Paletted
	for _, img := range r.images {
		if img.Bounds().Max.X == maxWidth+SourceWidth {
			images = append(images, img)
		} else {
			dst := image.NewPaletted(image.Rect(0, 0, maxWidth+SourceWidth, RecordHeight), Palette)
			screenWidth := img.Bounds().Max.X - SourceWidth
			margin := (maxWidth - screenWidth) / 2
			copyImage(dst, img, image.Rect(0, 0, screenWidth, RecordHeight), image.Point{margin, 0})
			copyImage(dst, img, image.Rect(screenWidth, 0, img.Bounds().Max.X, RecordHeight), image.Point{maxWidth - screenWidth, 0})
			images = append(images, dst)
		}
	}

	compressGif(images)

	g := gif.GIF{
		Image: images,
	}
	for range g.Image {
		g.Delay = append(g.Delay, 100)
	}
	g.Delay[len(g.Delay)-1] = 500

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

	offset := line + 3 - RecordHeight/LineHeight
	if offset < 1 {
		offset = 1
	}

	for i, l := range lines[offset-1:] {
		if i > RecordHeight/LineHeight {
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
