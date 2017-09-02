package poc

import (
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"strconv"

	"github.com/disintegration/gift"
)

var white color.Gray
var black color.Gray

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func main() {
	white.Y = 255
	black.Y = 0
	threshold, err := strconv.Atoi(os.Args[2])
	if err != nil {
		panic(err)
	}
	// Read the image
	reader, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	defer reader.Close()
	out, err := os.Create("./output.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	out1, err := os.Create("./output1.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	out2, err := os.Create("./output2.jpg")
	if err != nil {
		fmt.Println(err)
		return
	}
	var opt jpeg.Options
	opt.Quality = 100

	m, _, err := image.Decode(reader)
	if err != nil {
		fmt.Println(err)
		return
	}
	bounds := m.Bounds()

	// Convert to Grayscale
	g := gift.New(
		gift.Grayscale(),
		gift.Invert())
	dst := image.NewGray(g.Bounds(bounds))
	g.Draw(dst, m)

	err = jpeg.Encode(out, dst, &opt)
	if err != nil {
		fmt.Println(err)
	}

	// Binary filter
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			intensity := dst.GrayAt(x, y)
			if intensity.Y > uint8(threshold) {
				dst.SetGray(x, y, white)
			} else {
				dst.SetGray(x, y, black)
			}
		}
	}
	// Print the result
	err = jpeg.Encode(out1, dst, &opt)
	if err != nil {
		fmt.Println(err)
	}
	// Smooth the result
	dst = erode(dst)
	dst = erode(dst)
	for i := 0; i < 15; i++ {
		dst = dilate(dst)
	}

	// Color the result
	g = gift.New(gift.ColorFunc(func(r0 float32, g0 float32, b0 float32, a0 float32) (r, g, b, a float32) {
		if r0 == 0 {
			r = .2
			g = .2
			b = .5
			a = 1
			return
		} else {
			r = .5
			g = .5
			b = 0
			a = 1
			return
		}
	}))
	dst1 := image.NewRGBA(g.Bounds(dst.Bounds()))
	g.Draw(dst1, dst)

	err = jpeg.Encode(out2, dst1, &opt)
	if err != nil {
		fmt.Println(err)
	}

}

func erode(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	var newImg *image.Gray = image.NewGray(bounds)
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			shouldBreak := false
			for i := y - 1; i <= y+1; i++ {
				for j := x - 1; j <= x+1; j++ {
					if img.GrayAt(j, i).Y == 0 {
						newImg.SetGray(x, y, black)
						shouldBreak = true
						break
					}
				}
				if shouldBreak {
					break
				}
			}
			if !shouldBreak {
				newImg.SetGray(x, y, white)
			}
		}
	}
	return newImg
}

func dilate(img *image.Gray) *image.Gray {
	bounds := img.Bounds()
	var newImg *image.Gray = image.NewGray(bounds)
	for y := bounds.Min.Y + 1; y < bounds.Max.Y-1; y++ {
		for x := bounds.Min.X + 1; x < bounds.Max.X-1; x++ {
			shouldBreak := false
			for i := y - 1; i <= y+1; i++ {
				for j := x - 1; j <= x+1; j++ {
					if img.GrayAt(j, i).Y == 255 {
						newImg.SetGray(x, y, white)
						shouldBreak = true
						break
					}
				}
				if shouldBreak {
					break
				}
			}
			if !shouldBreak {
				newImg.SetGray(x, y, black)
			}
		}
	}
	return newImg
}
