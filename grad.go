package main

import (
	"image/color"
)

type GradStop struct {
	Position float64
	Color    color.RGBA
}

type ColourGrad []GradStop

func (cg ColourGrad) makePal(min, max float64, numCol int) color.Palette {
	pal := make([]color.Color, numCol, numCol)

	p := min
	pStep := (max - min) / float64(numCol)

	for i := range pal {
		pal[i] = cg.getColour(p)
		p += pStep
	}

	return color.Palette(pal)
}

func (cg ColourGrad) getColour(pos float64) color.RGBA {
	if pos <= 0.0 || len(cg) == 1 {
		return cg[0].Color
	}

	last := cg[len(cg)-1]
	if pos >= last.Position {
		return last.Color
	}

	for i, stop := range cg[1:] {
		if pos < stop.Position {
			pos = (pos - cg[i].Position) / (stop.Position - cg[i].Position)

			return colorLerp(cg[i].Color, stop.Color, pos)
		}
	}

	return last.Color
}

func colorLerp(c0, c1 color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		uint8Lerp(c0.R, c1.R, t),
		uint8Lerp(c0.G, c1.G, t),
		uint8Lerp(c0.B, c1.B, t),
		uint8Lerp(c0.A, c1.A, t),
	}
}

func uint8Lerp(a, b uint8, t float64) uint8 {
	return uint8(float64(a)*(1.0-t) + float64(b)*t)
}
