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

type pt struct {
	w    float64
	d    float64
	d_in float64
}

func computeD(r float64, R float64, D float64, phi_deg int, theta float64) float64 {
	phi := float64(phi_deg) / 360.0 * 2*math.Pi
	x_disp := (r*math.Sin(theta))
	d := D - (math.Sqrt(R*R - x_disp*x_disp))
	
	if (phi_deg != 90) {
		d += (r - r*math.Cos(theta)) / math.Tan(phi)
	}
	return d
}

func computeIntersection(r float64, t float64, R float64, D float64, phi_deg int) (pts []pt) {	
	for w := float64(0.0); w < r*2*math.Pi; w += 0.05 {
		theta := float64(w/r);
		d_out := computeD(r, R, D, phi_deg, theta);
		d_in := computeD(r - t, R, D, phi_deg, theta);
		pts = append(pts, pt{w, d_out, d_in})
	}
	return pts
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


	R,_ := strconv.ParseFloat(req.FormValue("R"), 64)
	r,_ := strconv.ParseFloat(req.FormValue("r"), 64)
	phi_deg,_ := strconv.ParseFloat(req.FormValue("phi"), 64)
	thick,_ := strconv.ParseFloat(req.FormValue("t"), 64);
	
	D := math.Trunc(R + 4)
	pts := computeIntersection(r, thick, R, D, int(phi_deg))
	
	gc.MoveTo(0, 0)
	for _,p := range pts {
		gc.LineTo(p.d * 100, p.w * 100)
	}
	gc.Stroke()

	gc.MoveTo(0, 0)
	for _,p := range pts {
		gc.LineTo(p.d_in * 100, p.w * 100)
	}
	gc.Stroke()

	png.Encode(writer, dest)
}
