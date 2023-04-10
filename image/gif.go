package image

import (
	"image"
	"image/draw"
)

func PalettedToRGBA(last *image.RGBA, in *image.Paletted) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, 16, 16))

	draw.Draw(out, out.Bounds(), last, last.Rect.Min, draw.Src)
	draw.Draw(out, in.Bounds(), in, in.Rect.Min, draw.Over)

	return out
}
