package main

import (
	"bufio"
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"strconv"
	"strings"

	"github.com/disintegration/gift"
)

var white color.Gray
var black color.Gray
var opt jpeg.Options

func init() {
	image.RegisterFormat("jpeg", "jpeg", jpeg.Decode, jpeg.DecodeConfig)
}

func main() {
	// Initialize vars
	white.Y = 255
	black.Y = 0
	opt.Quality = 100
	reader := bufio.NewReader(os.Stdin)

	// Ask user for file path to source image
	fmt.Println("Welcome: Please input the filepath for the image you want to convert into a D&D map.")
	filename := readLine(reader)

	// Open file, convert to grayscale and invert
	img := openFile(filename)
	fmt.Println("One moment...")

	// Make a range of binary images with thresholds 100 - 200 intervals of 10
	// save these images in threshold_samples
	for i := 100; i <= 200; i += 10 {
		// Generate filename
		var buffer bytes.Buffer
		buffer.WriteString("threshold_samples/")
		buffer.WriteString(strconv.Itoa(i))
		buffer.WriteString(".jpg")
		filename := buffer.String()
		// Make the image and print it
		sample := binary(img, uint8(i))
		printImage(sample, filename)
	}

	// Ask user to review generated samples and select a threshold
	fmt.Println("Please review the images in the folder called \"threshold_samples\" and select the desired image.")
	fmt.Println("For best results, choose an image where the dice are mostly completely white, but there is little to no noise(fuzz) around the edges of the image")
	fmt.Println("Input the number in the chosen file's name here (for example input 170 if 170.jpg looks best)")
	threshold, err := strconv.Atoi(readLine(reader))
	if err != nil {
		panic(err)
	}
	fmt.Println("One moment...")

	// Generate binary image with selected threshold
	img = binary(img, uint8(threshold))
	img = erode(img)

	// Make a range of images dialated 0 - 30 times in intervals of 5
	// save to smooth_samples
	smoothdst := dilate(img)
	for i := 0; i <= 30; i++ {
		smoothdst = dilate(smoothdst)

		if i%5 == 0 {
			// Generate filename
			var buffer bytes.Buffer
			buffer.WriteString("smooth_samples/")
			buffer.WriteString(strconv.Itoa(i))
			buffer.WriteString(".jpg")
			filename := buffer.String()

			printImage(smoothdst, filename)
		}
	}
	// Ask user to review samples and select a smooth value
	fmt.Println("Please review the sample images in \"smooth_samples\" and select the best one. This will be the final shape of your map.")
	fmt.Println("As before, type the number from the filename of your favorite image. Or if none have been bloated enough feel free to guess a larger number.")
	// Dialate image desired number of times and erase smooth_samples
	smoothNum, err := strconv.Atoi(readLine(reader))
	if err != nil {
		panic(err)
	}
	for i := 0; i <= smoothNum; i++ {
		img = dilate(img)
	}

	// Colorize image
	g := gift.New(gift.ColorFunc(func(r0 float32, g0 float32, b0 float32, a0 float32) (r, g, b, a float32) {
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
	final := image.NewRGBA(g.Bounds(img.Bounds()))
	g.Draw(final, img)
	// Output final result
	out, err := os.Create("result.jpg")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = jpeg.Encode(out, final, &opt)
	if err != nil {
		panic(err)
	}
}

func readLine(reader *bufio.Reader) string {
	line, _ := reader.ReadString('\n')
	line = strings.Replace(line, "\n", "", -1)
	return line
}

func openFile(filename string) *image.Gray {
	// Read the file
	reader, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	// Decode into an image
	m, _, err := image.Decode(reader)
	if err != nil {
		panic(err)
	}

	// Convert to Grayscale and invert
	g := gift.New(
		gift.Grayscale(),
		gift.Invert())
	dst := image.NewGray(g.Bounds(m.Bounds()))
	g.Draw(dst, m)

	return dst
}

func binary(src *image.Gray, threshold uint8) *image.Gray {
	dst := image.NewGray(src.Bounds())
	for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
		for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
			intensity := src.GrayAt(x, y)
			if intensity.Y > threshold {
				dst.SetGray(x, y, white)
			} else {
				dst.SetGray(x, y, black)
			}
		}
	}
	return dst
}

func printImage(src *image.Gray, filename string) {
	out, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = jpeg.Encode(out, src, &opt)
	if err != nil {
		panic(err)
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
