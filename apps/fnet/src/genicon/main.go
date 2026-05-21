package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	sizes := map[string]int{
		"apps/fnet/fnos/ICON.PNG":       64,
		"apps/fnet/fnos/ICON_256.PNG":   256,
		"apps/fnet/fnos/ui/images/64.png": 64,
	}

	for path, size := range sizes {
		img := image.NewRGBA(image.Rect(0, 0, size, size))

		// Background: dark blue gradient-ish background
		bg := color.RGBA{30, 50, 100, 255}
		draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{}, draw.Src)

		// Inner rounded rect effect - slightly lighter
		margin := size / 8
		inner := image.Rect(margin, margin, size-margin, size-margin)
		innerColor := color.RGBA{50, 80, 160, 255}
		draw.Draw(img, inner, &image.Uniform{innerColor}, image.Point{}, draw.Src)

		// Draw "FN" text using simple pixel drawing (8x8-ish per char)
		letterColor := color.RGBA{255, 255, 255, 255}
		charSize := size / 3
		startX := size/2 - charSize
		startY := size/2 - charSize/2

		// F
		drawLetter(img, 'F', startX, startY, charSize, letterColor)
		// N
		drawLetter(img, 'N', startX+charSize+size/16, startY, charSize, letterColor)

		f, _ := os.Create(path)
		png.Encode(f, img)
		f.Close()
	}
}

func drawLetter(img *image.RGBA, letter byte, x, y, charSize int, c color.Color) {
	thick := charSize / 6
	if thick < 1 {
		thick = 1
	}

	switch letter {
	case 'F':
		// Vertical line
		rect(img, x, y, x+thick, y+charSize, c)
		// Top horizontal
		rect(img, x, y, x+charSize, y+thick, c)
		// Middle horizontal
		rect(img, x, y+charSize/2-thick/2, x+charSize*2/3, y+charSize/2+thick/2, c)

	case 'N':
		// Left vertical
		rect(img, x, y, x+thick, y+charSize, c)
		// Right vertical
		rect(img, x+charSize-thick, y, x+charSize, y+charSize, c)
		// Diagonal
		for i := 0; i <= charSize; i++ {
			diagX := x + thick + i*(charSize-thick*2)/charSize
			diagY := y + i
			for d := -thick / 2; d <= thick/2; d++ {
				px := diagX + d
				py := diagY
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}

func rect(img *image.RGBA, x1, y1, x2, y2 int, c color.Color) {
	for y := y1; y < y2; y++ {
		for x := x1; x < x2; x++ {
			if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
				img.Set(x, y, c)
			}
		}
	}
}
