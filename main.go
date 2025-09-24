package main

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/nfnt/resize"
)

func convertImageToAnsi(img image.Image, width int) string {
	bounds := img.Bounds()
	origWidth := bounds.Dx()
	origHeight := bounds.Dy()

	aspectRatio := float64(origHeight) / float64(origWidth)
	height := int(float64(width) * aspectRatio)

	resizedImg := resize.Resize(uint(width), uint(height), img, resize.Lanczos3)

	var builder strings.Builder
	resizedBounds := resizedImg.Bounds()

	for y := resizedBounds.Min.Y; y < resizedBounds.Max.Y; y += 2 {
		for x := resizedBounds.Min.X; x < resizedBounds.Max.X; x++ {
			r1, g1, b1, _ := resizedImg.At(x, y).RGBA()
			r1_8, g1_8, b1_8 := uint8(r1>>8), uint8(g1>>8), uint8(b1>>8)

			r2, g2, b2 := r1_8, g1_8, b1_8
			if y+1 < resizedBounds.Max.Y {
				r2_raw, g2_raw, b2_raw, _ := resizedImg.At(x, y+1).RGBA()
				r2, g2, b2 = uint8(r2_raw>>8), uint8(g2_raw>>8), uint8(b2_raw>>8)
			}

			ansiString := fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", r1_8, g1_8, b1_8, r2, g2, b2)

			builder.WriteString(ansiString)
		}
		builder.WriteString("\x1b[0m\n")
	}

	builder.WriteString("\x1b[0m")

	return builder.String()
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <image_path> [width]")
	}
	filePath := os.Args[1]

	outputWidth := 80
	if len(os.Args) > 2 {
		var err error
		outputWidth, err = strconv.Atoi(os.Args[2])
		if err != nil {
			log.Fatalf("Invalid width: %v", err)
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatalf("Failed to decode image: %v", err)
	}

	ansiArt := convertImageToAnsi(img, outputWidth)

	fmt.Print(ansiArt)
}
