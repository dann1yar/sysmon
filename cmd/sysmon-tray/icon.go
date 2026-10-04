package main

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
)

func statusIcon(pct float64) []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	var c color.RGBA
	switch {
	case pct < 50:
		c = color.RGBA{R: 0x4c, G: 0xaf, B: 0x50, A: 0xff} // green
	case pct < 80:
		c = color.RGBA{R: 0xff, G: 0xb3, B: 0x00, A: 0xff} // amber
	default:
		c = color.RGBA{R: 0xe5, G: 0x39, B: 0x35, A: 0xff} // red
	}

	cx, cy, r := float64(size)/2, float64(size)/2, float64(size)/2-2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx, dy := float64(x)+0.5-cx, float64(y)+0.5-cy
			if dx*dx+dy*dy <= r*r {
				img.Set(x, y, c)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
