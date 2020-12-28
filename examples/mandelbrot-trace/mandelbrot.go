// mandelbrot example code adapted from Francesc Campoy's mandelbrot package.
// https://github.com/campoy/mandelbrot
package main

import (
	"flag"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"sync"

	"github.com/pkg/profile"
)

func main() {
	defer profile.Start(profile.TraceProfile, profile.ProfilePath(".")).Stop()

	var (
		height  = flag.Int("h", 1024, "height of the output image in pixels")
		width   = flag.Int("w", 1024, "width of the output image in pixels")
		mode    = flag.String("mode", "seq", "mode: seq, px, row, workers")
		workers = flag.Int("workers", 1, "number of workers to use")
	)
	flag.Parse()

	const output = "mandelbrot.png"

	// open a new file
	f, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}

	// create the image
	c := make([][]color.RGBA, *height)
	for i := range c {
		c[i] = make([]color.RGBA, *width)
	}

	img := &img{
		h: *height,
		w: *width,
		m: c,
	}

	switch *mode {
	case "seq":
		// 1 G
		seqFillImg(img)
	case "px":
		// (height x width) 1'048'576 G's	=> 1024 * 1024 (1<<20)
		oneToOneFillImg(img)
	case "row":
		// (width) 1024 G's	=> (1<<10)
		onePerRowFillImg(img)
	case "workers":
		nWorkersFillImg(img, *workers)
	default:
		panic("unknown mode")
	}

	// and encoding it
	if err := png.Encode(f, img); err != nil {
		log.Fatal(err)
	}
}

type img struct {
	h, w int
	m    [][]color.RGBA
}

func (m *img) At(x, y int) color.Color { return m.m[x][y] }
func (m *img) ColorModel() color.Model { return color.RGBAModel }
func (m *img) Bounds() image.Rectangle { return image.Rect(0, 0, m.h, m.w) }

// sequential
func seqFillImg(m *img) {
	for i, row := range m.m {
		for j := range row {
			fillPixel(m, i, j)
		}
	}
}

// one goroutine per pixel
func oneToOneFillImg(m *img) {
	var wg sync.WaitGroup
	wg.Add(m.h * m.w) // 1024 * 1024 (1<<20)
	for i, row := range m.m {
		for j := range row {
			go func(i, j int) { // 1G => 1 pixel
				fillPixel(m, i, j)
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()
}

// one goroutine per row of pixels
func onePerRowFillImg(m *img) {
	var wg sync.WaitGroup
	wg.Add(m.h) // 1024 (1<<10)
	for i := range m.m {
		go func(i int) { // 1G => 1024 pixels (row)
			for j := range m.m[i] {
				fillPixel(m, i, j)
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func nWorkersFillImg(m *img, workers int) {
	c := make(chan int, 1024)
	var wg sync.WaitGroup
	wg.Add(workers)                // 1 (default)
	for i := 0; i < workers; i++ { // G: (1 * workers)
		go func() {
			for i := range c { // channel (row)
				for j := range m.m[i] { // col
					fillPixel(m, i, j)
				}
			}
			wg.Done()
		}()
	}

	for i := range m.m {
		c <- i
	}
	close(c)
	wg.Wait()
}

func fillPixel(m *img, x, y int) {
	const n = 1000
	const Limit = 2.0
	Zr, Zi, Tr, Ti := 0.0, 0.0, 0.0, 0.0
	Cr := (2*float64(x)/float64(n) - 1.5)
	Ci := (2*float64(y)/float64(n) - 1.0)

	for i := 0; i < n && (Tr+Ti <= Limit*Limit); i++ {
		Zi = 2*Zr*Zi + Ci
		Zr = Tr - Ti + Cr
		Tr = Zr * Zr
		Ti = Zi * Zi
	}
	paint(&m.m[x][y], Tr, Ti)
}

func paint(c *color.RGBA, x, y float64) {
	n := byte(x * y * 2)
	c.R, c.G, c.B, c.A = n, n, n, 255
}
