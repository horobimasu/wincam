package win_cam

import (
	"bytes"
	"image"
	"image/png"
)

const (
	_NV12_CHROMA_OFFSET          float64 = 128
	_NV12_RED_OFFSET             float64 = 1.370705
	_NV12_GREEN_FROM_RED_OFFSET  float64 = 0.698001
	_NV12_GREEN_FROM_BLUE_OFFSET float64 = 0.337633
	_NV12_BLUE_OFFSET            float64 = 1.732446
)

func ConvertToPNG(pixels []byte, width int, height int) (*bytes.Buffer, error) {
	picture := image.NewRGBA(image.Rect(0, 0, width, height))

	for y := range height {
		for x := range width {
			i := (y*width + x) * 3

			picture.Pix[(y*width+x)*4+0] = pixels[i+0]
			picture.Pix[(y*width+x)*4+1] = pixels[i+1]
			picture.Pix[(y*width+x)*4+2] = pixels[i+2]
			picture.Pix[(y*width+x)*4+3] = 255
		}
	}

	var buffer bytes.Buffer

	err := png.Encode(&buffer, picture)
	if err != nil {
		return nil, err
	}

	return &buffer, nil
}

func ParseNV12(pixels []byte, width int, height int) []byte {
	newPixels := make([]byte, width*height*3)

	lumma := pixels[:width*height]
	chroma := pixels[width*height:]

	for y := range height {
		for x := range width {
			block := (y/2)*width + (x/2)*2

			blue := float64(chroma[block]) - _NV12_CHROMA_OFFSET
			red := float64(chroma[block+1]) - _NV12_CHROMA_OFFSET
			alpha := float64(lumma[y*width+x])

			i := (y*width + x) * 3

			newPixels[i+0] = byte(clamp(alpha + _NV12_RED_OFFSET*red))
			newPixels[i+1] = byte(clamp(alpha - _NV12_GREEN_FROM_RED_OFFSET*red - _NV12_GREEN_FROM_BLUE_OFFSET*blue))
			newPixels[i+2] = byte(clamp(alpha + _NV12_BLUE_OFFSET*blue))
		}
	}

	return newPixels
}

func ParseYUY2(pixels []byte, width int, height int) []byte {
	newPixels := make([]byte, width*height*3)

	for y := range height {
		for x := 0; x < width; x += 2 {
			base := (y*width + x) * 2

			leftAlpha := float64(pixels[base+0])
			blue := float64(pixels[base+1]) - 128

			rightAlpha := float64(pixels[base+2])
			red := float64(pixels[base+3]) - 128

			i0 := (y*width + x) * 3

			newPixels[i0+0] = byte(clamp(leftAlpha + 1.370705*red))
			newPixels[i0+1] = byte(clamp(leftAlpha - 0.698001*red - 0.337633*blue))
			newPixels[i0+2] = byte(clamp(leftAlpha + 1.732446*blue))

			i1 := i0 + 3

			if x+1 < width {
				newPixels[i1+0] = byte(clamp(rightAlpha + 1.370705*red))
				newPixels[i1+1] = byte(clamp(rightAlpha - 0.698001*red - 0.337633*blue))
				newPixels[i1+2] = byte(clamp(rightAlpha + 1.732446*blue))
			}
		}
	}

	return newPixels
}

func clamp(num float64) float64 {
	if num < 0 {
		return 0
	}

	if num > 255 {
		return 255
	}

	return num
}
