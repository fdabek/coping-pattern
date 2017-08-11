package main

import (
	"fmt"
	"github.com/llgcode/draw2d/draw2dimg"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dpdf"
	"image"
	"image/color"
	"image/png"
	"io"
	"math"
	"net/http"
	_ "net/http/pprof"
	"strconv"
)

func init() {
	http.HandleFunc("/png", handler)
	http.Handle("/", http.FileServer(http.Dir("./static")))
}

type pt struct {
	w    float64
	d    float64
	d_in float64
}

func computeD(r float64, R float64, D float64, phi_deg int, theta float64) float64 {
	phi := float64(phi_deg) / 360.0 * 2*math.Pi
	x_disp := (r*math.Sin(theta))
	d := D
	if math.Abs(x_disp) < R {
		d -= (math.Sqrt(R*R - x_disp*x_disp))
	}
	
	if phi_deg != 90 {
		d += (r - r*math.Cos(theta)) / math.Tan(phi)
	}
	return d
}

func computeIntersection(r float64, t float64, R float64, D float64, phi_deg int) (pts []pt) {	
	for w := float64(0.0); w < r*2*math.Pi; w += 0.025 {
		theta := float64(w/r);
		d_out := computeD(r, R, D, phi_deg, theta);
		d_in := computeD(r - t, R, D, phi_deg, theta);
		pts = append(pts, pt{w, d_out, d_in})
	}
	return pts
}

func dumpText(writer http.ResponseWriter, pts []pt) {
	for _, p := range pts {
		io.WriteString(writer, fmt.Sprintf("%f %f\n", p.w, p.d));
	}
}

func DrawPattern(gc draw2d.GraphicContext,
	r float64,
	R float64,
	D float64,
	phi_deg int,
	pts []pt,
	scale float64) {
	gc.SetStrokeColor(color.Gray{0x00})
	gc.SetLineWidth(0.02 * scale)

	// outside profile:
	gc.MoveTo(0, 0)
	for _,p := range pts {
		gc.LineTo(p.d * scale, p.w * scale)
	}
	gc.Stroke()

	// inside profile:
	gc.SetStrokeColor(color.Gray{0xaa})
	gc.MoveTo(0, 0)
	for _,p := range pts {
		gc.LineTo(p.d_in * scale, p.w * scale)
	}
	gc.Stroke()


	// quarter rotation lines:
	gc.SetStrokeColor(color.Gray{0x00})
	for i := float64(1.0); i <= 4; i++ {
		theta := float64(math.Pi*i/2.0)
		quarter_d := computeD(r, R, D, int(phi_deg), theta)
		quarter_w := theta*r
		gc.MoveTo(0, quarter_w * scale)
		gc.LineTo(quarter_d * scale, quarter_w * scale)
		
	}
	gc.Stroke()
}

func dumpPng(
	writer http.ResponseWriter,
	r float64,
	R float64,
	D float64,
	phi_deg int,
	pts []pt) {

	dest := image.NewRGBA(image.Rect(0, 0, 1000, 1000))
	gc := draw2dimg.NewGraphicContext(dest)

	DrawPattern(gc,r,R,D,phi_deg,pts, 100.0)

	// text output does not appear to be portable across PNG/PDF
	draw2d.SetFontFolder(".")
	gc.SetFontData(draw2d.FontData{"go", draw2d.FontFamilyMono, draw2d.FontStyleBold})
	gc.SetFontSize(12)
	gc.SetFillColor(color.Gray{0x00})
	gc.FillStringAt(fmt.Sprintf("Edge of pattern is %0.2fin from center of tube", D),
		10, 30)
	gc.Fill()

	png.Encode(writer, dest)
}

func dumpPdf(
	writer http.ResponseWriter,
	r float64,
	R float64,
	D float64,
	phi_deg int,
	pts []pt) {
	
	dest := draw2dpdf.NewPdf("L", "in", "Letter")
	gc := draw2dpdf.NewGraphicContext(dest)

	dest.SetFont("Courier", "B", 12)
	dest.Cell(0.5, 0.5, fmt.Sprintf("Edge of pattern is %0.2fin from center of tube", D))

	DrawPattern(gc,r,R,D,phi_deg,pts,1.0)
	e := dest.Output(writer)
	if e != nil {
		fmt.Fprint(writer, "Error: ", e);
	}
}

func handler(writer http.ResponseWriter, req *http.Request) {
	if (req.ParseForm() != nil) {
		fmt.Fprint(writer, "Error parsing request")
		return
	}

	R,_ := strconv.ParseFloat(req.FormValue("R"), 64)
	r,_ := strconv.ParseFloat(req.FormValue("r"), 64)
	phi_deg,_ := strconv.ParseFloat(req.FormValue("phi"), 64)
	thick,_ := strconv.ParseFloat(req.FormValue("t"), 64);
	format := req.FormValue("f");

	R = R/2;
	r = r/2;
	
	D := math.Trunc(R + 4)
	pts := computeIntersection(r, thick, R, D, int(phi_deg))

	if format == "text" {
		dumpText(writer, pts)
	} else if format == "png" {
		dumpPng(writer, r, R, D, int(phi_deg), pts)
	} else if format == "pdf" {
		dumpPdf(writer, r, R, D, int(phi_deg), pts)
	}

}
