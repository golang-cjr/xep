package main

import (
	"image"
	"image/color"
	"image/png"
	"io"
	"math/rand"
)

const AvatarSize = 24

func Splatter(avatar *image.RGBA, nameBytes []byte, pixelColor color.RGBA) {

	// A somewhat random number based on the username.
	var nameSum int64
	for i := range nameBytes {
		nameSum += int64(nameBytes[i])
	}

	// Use said number to keep random-ness deterministic for a given name
	rand.Seed(nameSum)

	// Make the "splatter"
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			if ((x + y) % 2) == 0 {
				if rand.Intn(2) == 1 {
					avatar.SetRGBA(x, y, pixelColor)
				}
			}
		}
	}

	// Mirror left half to right half
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			if x < AvatarSize/2 {
				avatar.Set(AvatarSize-x-1, y, avatar.At(x, y))
			}
		}
	}

	// Mirror top to bottom
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			if y < AvatarSize/2 {
				avatar.Set(x, AvatarSize-y-1, avatar.At(x, y))
			}
		}
	}
}

func PaintBG(avatar *image.RGBA, bgColor color.RGBA) {
	for y := 0; y < AvatarSize; y++ {
		for x := 0; x < AvatarSize; x++ {
			avatar.SetRGBA(x, y, bgColor)
		}
	}
}

func CalcPixelColor(nameBytes []byte) (pixelColor color.RGBA) {
	pixelColor.A = 255

	var mutator = byte((len(nameBytes) * 4))

	pixelColor.R = nameBytes[0] * mutator
	pixelColor.G = nameBytes[1] * mutator
	pixelColor.B = nameBytes[2] * mutator

	return
}

func CalcBGColor(nameBytes []byte) (bgColor color.RGBA) {
	bgColor.A = 255

	var mutator = byte((len(nameBytes) * 2))

	bgColor.R = nameBytes[0] * mutator
	bgColor.G = nameBytes[1] * mutator
	bgColor.B = nameBytes[2] * mutator

	return
}

func ico(wr io.Writer) {
	nameBytes := []byte("golang@conference.jabber.ru/xep")
	avatar := image.NewRGBA(image.Rect(0, 0, AvatarSize, AvatarSize))
	PaintBG(avatar, CalcBGColor(nameBytes))
	Splatter(avatar, nameBytes, CalcPixelColor(nameBytes))
	png.Encode(wr, avatar)
}
