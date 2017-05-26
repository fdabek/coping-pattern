package main

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"image"
	"image/color"
	"image/png"
	"math"
	"net/http"
	"strconv"
)

func init() {
    http.HandleFunc("/png", handler)
}

func handler(writer http.ResponseWriter, req *http.Request) {
	if (req.ParseForm() != nil) {
		fmt.Fprint(writer, "Error parsing request")
		return
	}

	dest := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	gc := draw2dimg.NewGraphicContext(dest)
	gc.SetStrokeColor(color.RGBA{0x44, 0x44, 0x44, 0xff})
	gc.SetLineWidth(2)

	const kVertOffset float64 = 500.0
	const kPi float64 = 3.14159

	gc.MoveTo(0, kVertOffset)

	R,_ := strconv.ParseFloat(req.FormValue("R"), 64)
	r,_ := strconv.ParseFloat(req.FormValue("r"), 64)
	phi_deg,_ := strconv.ParseFloat(req.FormValue("phi"), 64)
	phi := phi_deg / 360.0 * 2*kPi
	cosphi := math.Cos(phi)

	var w float64 = 0
	for ; w < r*2*kPi; w += 0.1 {
                x_disp := (r*math.Sin(w/r))
		d := R - math.Sqrt(R*R - x_disp*x_disp) + (r - r*math.Cos(w/r))*cosphi
		fmt.Println(w * 100, kVertOffset + d * 100)
		gc.LineTo(w * 100, kVertOffset + d*100)
	}
	gc.Stroke()

	png.Encode(writer, dest)
}
