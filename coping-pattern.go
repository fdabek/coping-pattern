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
	_ "net/http/pprof"
	"strconv"
	"log"
	"net/http"
	"os"
)

func main() {
	http.HandleFunc("/png", handler)
	http.Handle("/", http.FileServer(http.Dir("./static")))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	err := http.ListenAndServe(":"+port, nil);
	if err != nil {
	   log.Fatal(err)
	}
}

type pt struct {
	w    float64
	cut_d    float64
	cut_d_in float64
	angle_d float64
}

// r: cut tube radius
// R: other tube radius
// theta: angle around cut tube "clockwise" (radians)
func computeCutDelta(r float64, R float64, theta float64) float64 {
	// first we're going to consider how much to displace the cut edge of
	// the pattern from "D" due to the "cutting action" of the other
	// tube. We are walking around the cut tube (with theta) and computing
	// the displacement from the center of the cuting tube. Imagine a
	// triangle with the base of length x_disp at the cuttin tube,
	// hypotenuse of length R. The displacment is just the other leg:
	x_disp := (r*math.Sin(theta))
	if math.Abs(x_disp) < R {
		 return (math.Sqrt(R*R - x_disp*x_disp))
	} else {
		return 0
	}

}

// phi_deg: angle between tubes (degress)
// r: cut tube radius
// theta: angle around cut tube "clockwise" (radians)
func computeAngleDelta(phi_deg int, r float64, theta float64) float64 {
	// if tubes are perpendicular, no adjustment
	if phi_deg == 90 { return 0 }

	// convert to radians here so we don't need to match aginst pi/2
	phi := float64(phi_deg) / 360.0 * 2*math.Pi

	// return displacment due to the angle between the tubes (if it's not
	// 90). This is the distance a plane tilted at phi_deg degrees cuts
	// into the cut tube at a given position (theta) along the pattern.
	//
	// Computed as the leg of a triange with hypotenuse of r and the other
	// leg as the "y distance" above (or below) the center of the cut tube.
	return (r*math.Cos(theta)) / math.Tan(phi)
}

func computeIntersection(r float64, t float64, R float64, phi_deg int) (pts []pt) {
	// for every point in the linear space of the pattern (w) compute the
	// corresponding angle (theta) and the displacment at that angle.
	for w := float64(0.0); w < r*2*math.Pi; w += 0.005 {
		theta := float64(w/r)
		cut_d_out := computeCutDelta(r, R, theta)
		cut_d_in := computeCutDelta(r - t, R, theta)
		angle_d := computeAngleDelta(phi_deg, r, theta)
		pts = append(pts, pt{w, cut_d_out, cut_d_in, angle_d})
	}
	return pts
}

func dumpText(writer http.ResponseWriter, pts []pt) {
	for _, p := range pts {
		io.WriteString(writer, fmt.Sprintf("%f %f\n", p.w, p.cut_d + p.angle_d));
	}
}

func LeftEdge(p pt, D float64, offset float64) float64 {
	return offset - p.angle_d
}

func RightEdge(p pt, D float64, offset float64) float64 {
	return D + offset - p.angle_d - p.cut_d
}

func DrawPattern(gc draw2d.GraphicContext,
	r float64,
	R float64,
	D float64,
	phi_deg int,
	pts []pt,
	scale float64,
	border float64) {
	gc.Save()
	gc.Translate(border * scale, border * scale)
	gc.Scale(scale, scale);
	gc.SetStrokeColor(color.Gray{0x00})
	gc.SetLineWidth(0.02)

	// offset the pattern to make room for the edge curve:
	offset := D/4

	/*
	// inside profile. The profile is displaced by the angle of the tube
	// and the cutting action of the other tube and the widge of the pattern (D)
	gc.SetStrokeColor(color.Gray{0xaa})
	gc.MoveTo(offset + D - pts[0].cut_d_in - pts[0].angle_d, pts[0].w)
	for _,p := range pts {
		gc.LineTo(offset + D - p.cut_d_in - p.angle_d, p.w)
	}
	gc.Stroke()
	*/
	
	// outside profile:
	gc.SetStrokeColor(color.Gray{0x00})
	gc.MoveTo(RightEdge(pts[0], D, offset), pts[0].w)
	for _,p := range pts {
		gc.LineTo(RightEdge(p,  D, offset), p.w)
	}
	gc.Stroke()

	// now the left edge. This curve is displaced only by the angle between the tubes:
	gc.MoveTo(LeftEdge(pts[0], D, offset), pts[0].w)
	for _,p := range pts {
		gc.LineTo(LeftEdge(p, D, offset), p.w)
	}
	gc.Stroke()

	// quarter rotation lines:
	indices := []int{0, len(pts) / 4, len(pts) / 2, 3 * len(pts) / 4, len(pts) - 1}
	for _, index := range indices {
		gc.MoveTo(LeftEdge(pts[index], D, offset), pts[index].w)
		gc.LineTo(RightEdge(pts[index], D, offset),  pts[index].w)
	}
	
	
	// and the alignment notches every inch
	for i := offset; i <= offset + D - pts[0].cut_d - pts[0].angle_d; i++ {
		// top:
		gc.MoveTo(i, 0)
		gc.LineTo(i, 0.125)

		// bottom:
		bottom_w := math.Pi*2 * r
		gc.MoveTo(i, bottom_w)
		gc.LineTo(i, bottom_w - 0.125)
	}

	gc.Stroke()
	gc.Restore()
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

	DrawPattern(gc, r, R, D, phi_deg, pts, 100.0, 0.1)

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
	
	dest := draw2dpdf.NewPdf("P", "in", "Letter")
	gc := draw2dpdf.NewGraphicContext(dest)

	dest.SetFont("Courier", "B", 12)
	dest.Cell(0.5, 0.5, fmt.Sprintf("Edge of pattern is %0.2fin from center of tube", D))

	DrawPattern(gc,r,R,D,phi_deg,pts,1.0,1.0)
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
	pts := computeIntersection(r, thick, R, int(phi_deg))

	if format == "text" {
		writer.Header().Set("Content-type", "text/plain")
		dumpText(writer, pts)
	} else if format == "png" {
		writer.Header().Set("Content-type", "image/png")
		dumpPng(writer, r, R, D, int(phi_deg), pts)
	} else if format == "pdf" {
		writer.Header().Set("Content-type", "application/pdf")
		dumpPdf(writer, r, R, D, int(phi_deg), pts)
	} else {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("500 - Invalid output format"))
	}

}
