package main

import (
	"image/color"
	"image/draw"

	"time"
)

type digitmap [28]uint8

var (
	digitHeight = 7
	digitWidth  = 4
	colonWidth  = 2

	colonslice = []uint8{
		0, 0,
		1, 1,
		1, 1,
		0, 0,
		1, 1,
		1, 1,
		0, 0,
	}

	digitslice = [10]digitmap{
		{
			0, 1, 1, 0,
			1, 0, 0, 1,
			1, 0, 0, 1,
			1, 0, 0, 1,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 0,
		}, // 0
		{
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
		}, // 1
		{
			0, 1, 1, 0,
			1, 0, 0, 1,
			0, 0, 0, 1,
			0, 1, 1, 0,
			1, 0, 0, 0,
			1, 0, 0, 0,
			0, 1, 1, 1,
		}, // 2
		{
			0, 1, 1, 0,
			1, 0, 0, 1,
			0, 0, 0, 1,
			0, 1, 1, 0,
			0, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 0,
		}, // 3
		{
			0, 0, 0, 1,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
		}, // 4
		{
			0, 1, 1, 0,
			1, 0, 0, 0,
			1, 0, 0, 0,
			0, 1, 1, 0,
			0, 0, 0, 1,
			0, 0, 0, 1,
			1, 1, 1, 0,
		}, // 5
		{
			0, 1, 1, 0,
			1, 0, 0, 0,
			1, 0, 0, 0,
			0, 1, 1, 0,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 0,
		}, // 6
		{
			1, 1, 1, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 1, 0,
			0, 0, 1, 0,
			0, 0, 1, 0,
		}, // 7
		{
			0, 1, 1, 0,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 0,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 0,
		}, // 8
		{
			0, 1, 1, 0,
			1, 0, 0, 1,
			1, 0, 0, 1,
			0, 1, 1, 1,
			0, 0, 0, 1,
			0, 0, 0, 1,
			0, 0, 0, 0,
		}, // 9
	}
)

func DrawClock(dest draw.Image, t *time.Time) {

	x := 0
	y := 0

	DrawFrame(dest, x, y)

	x += 2
	y += 2

	hour := t.Hour()
	if hour >= 10 {
		DrawDigit(dest, digitslice[hour/10], x, y)
		x += 5
		DrawDigit(dest, digitslice[hour%10], x, y)

	} else {
		DrawDigit(dest, digitslice[0], x, y)
		x += 5
		DrawDigit(dest, digitslice[hour], x, y)

	}
	x += 5

	DrawColon(dest, x, y)
	x += 3

	min := t.Minute()
	if min >= 10 {
		DrawDigit(dest, digitslice[min/10], x, y)
		x += 5
		DrawDigit(dest, digitslice[min%10], x, y)
	} else {
		DrawDigit(dest, digitslice[0], x, y)
		x += 5
		DrawDigit(dest, digitslice[min], x, y)
	}
	x += 5
}

func DrawFrame(dest draw.Image, sx int, sy int) {
	mw := 32
	mh := 11

	for y, h := sy, 0; h < mh; y, h = y+1, h+1 {
		for x, w := sx, 0; w < mw; x, w = x+1, w+1 {
			dest.Set(x, y, color.Black)
		}
	}

	for x, w := sx, 0; w < mw; x, w = x+1, w+1 {
		dest.Set(x, sy, color.White)
		dest.Set(x, sy+mh, color.White)
	}

	for y, h := sy, 0; h < mh; y, h = y+1, h+1 {
		dest.Set(sx, y, color.White)
		dest.Set(sx+mw, y, color.White)
	}
}

func DrawDigit(dest draw.Image, dig digitmap, sx int, sy int) {
	c := 0
	for y, h := sy, 0; h < 7; y, h = y+1, h+1 {
		for x, w := sx, 0; w < 4; x, w = x+1, w+1 {
			if dig[c] > 0 {
				dest.Set(x, y, color.White)
			} else {
				dest.Set(x, y, color.Black)
			}
			c += 1
		}
	}
}

func DrawColon(dest draw.Image, sx int, sy int) {

	c := 0
	for y, h := sy, 0; h < 7; y, h = y+1, h+1 {

		for x, w := sx, 0; w < 2; x, w = x+1, w+1 {
			if colonslice[c] > 0 {
				dest.Set(x, y, color.White)
			} else {
				dest.Set(x, y, color.Black)
			}
			c += 1
		}
	}

}
