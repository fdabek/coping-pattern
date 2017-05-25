package main

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"math"
)

func main() {
	dest := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(2)

	const kVertOffset float64 = 500.0
	const kPi float64 = 3.14159

	gc.MoveTo(0, kVertOffset)

	R := 1.5
	r := 1.125
	phi := 72.0 / 360.0 * 2*kPi
	cosphi := math.Cos(phi)

	var w float64 = 0
	for ; w < r*2*kPi; w += 0.1 {
                x_disp := (r*math.Sin(w/r))
		d := R - math.Sqrt(R*R - x_disp*x_disp) - (r - r*math.Cos(w/r))*cosphi
		fmt.Println(w * 100, kVertOffset + d * 100)
		gc.LineTo(w * 100, kVertOffset + d*100)
	}
	gc.Stroke()

	draw2dimg.SaveToPngFile("cope.png", dest)
}
