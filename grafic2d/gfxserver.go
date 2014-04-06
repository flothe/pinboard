// High-level 2D vector graphics library built on OpenVG
package grafic2d

/*
#cgo CFLAGS:   -I/opt/vc/include -I/opt/vc/include/interface/vmcs_host/linux -I/opt/vc/include/interface/vcos/pthreads
#cgo LDFLAGS:  -L/opt/vc/lib -lGLESv2 -ljpeg
#include <stdlib.h>
#include "VG/openvg.h"
#include "VG/vgu.h"
#include "EGL/egl.h"
#include "GLES/gl.h"
#include "fontinfo.h" // font information
#include "shapes.h"   // C API
*/
import "C"
import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"runtime"
	"strings"
	"unsafe"
)

type RenderObject interface {
	Begin(*GFXServer) error
	End() error
	Update(int) error
	Draw() error
}

// RGB defines the red, green, blue triple that makes up colors.
type GFXServer struct {
	DisplayWidth, DisplayHeight int
	isDebugMode                 bool
}

// RGB defines the red, green, blue triple that makes up colors.
type RGB struct {
	Red, Green, Blue uint8
}

// VGfloat defines the basic type for coordinates, dimensions and other values
type VGfloat C.VGfloat
type VGint C.VGint
type VGImage C.VGImage
type VGImageFormat C.VGImageFormat
type VGImageQuality C.VGImageQuality

// Offcolor defines the offset, color and alpha values used in gradients
// the Offset ranges from 0..1, colors as RGB triples, alpha ranges from 0..1
type Offcolor struct {
	Offset VGfloat
	RGB
	Alpha VGfloat
}

// colornames maps SVG color names to RGB triples.
var colornames = map[string]RGB{
	"aliceblue":            {240, 248, 255},
	"antiquewhite":         {250, 235, 215},
	"aqua":                 {0, 255, 255},
	"aquamarine":           {127, 255, 212},
	"azure":                {240, 255, 255},
	"beige":                {245, 245, 220},
	"bisque":               {255, 228, 196},
	"black":                {0, 0, 0},
	"blanchedalmond":       {255, 235, 205},
	"blue":                 {0, 0, 255},
	"blueviolet":           {138, 43, 226},
	"brown":                {165, 42, 42},
	"burlywood":            {222, 184, 135},
	"cadetblue":            {95, 158, 160},
	"chartreuse":           {127, 255, 0},
	"chocolate":            {210, 105, 30},
	"coral":                {255, 127, 80},
	"cornflowerblue":       {100, 149, 237},
	"cornsilk":             {255, 248, 220},
	"crimson":              {220, 20, 60},
	"cyan":                 {0, 255, 255},
	"darkblue":             {0, 0, 139},
	"darkcyan":             {0, 139, 139},
	"darkgoldenrod":        {184, 134, 11},
	"darkgray":             {169, 169, 169},
	"darkgreen":            {0, 100, 0},
	"darkgrey":             {169, 169, 169},
	"darkkhaki":            {189, 183, 107},
	"darkmagenta":          {139, 0, 139},
	"darkolivegreen":       {85, 107, 47},
	"darkorange":           {255, 140, 0},
	"darkorchid":           {153, 50, 204},
	"darkred":              {139, 0, 0},
	"darksalmon":           {233, 150, 122},
	"darkseagreen":         {143, 188, 143},
	"darkslateblue":        {72, 61, 139},
	"darkslategray":        {47, 79, 79},
	"darkslategrey":        {47, 79, 79},
	"darkturquoise":        {0, 206, 209},
	"darkviolet":           {148, 0, 211},
	"deeppink":             {255, 20, 147},
	"deepskyblue":          {0, 191, 255},
	"dimgray":              {105, 105, 105},
	"dimgrey":              {105, 105, 105},
	"dodgerblue":           {30, 144, 255},
	"firebrick":            {178, 34, 34},
	"floralwhite":          {255, 250, 240},
	"forestgreen":          {34, 139, 34},
	"fuchsia":              {255, 0, 255},
	"gainsboro":            {220, 220, 220},
	"ghostwhite":           {248, 248, 255},
	"gold":                 {255, 215, 0},
	"goldenrod":            {218, 165, 32},
	"gray":                 {128, 128, 128},
	"green":                {0, 128, 0},
	"greenyellow":          {173, 255, 47},
	"grey":                 {128, 128, 128},
	"honeydew":             {240, 255, 240},
	"hotpink":              {255, 105, 180},
	"indianred":            {205, 92, 92},
	"indigo":               {75, 0, 130},
	"ivory":                {255, 255, 240},
	"khaki":                {240, 230, 140},
	"lavender":             {230, 230, 250},
	"lavenderblush":        {255, 240, 245},
	"lawngreen":            {124, 252, 0},
	"lemonchiffon":         {255, 250, 205},
	"lightblue":            {173, 216, 230},
	"lightcoral":           {240, 128, 128},
	"lightcyan":            {224, 255, 255},
	"lightgoldenrodyellow": {250, 250, 210},
	"lightgray":            {211, 211, 211},
	"lightgreen":           {144, 238, 144},
	"lightgrey":            {211, 211, 211},
	"lightpink":            {255, 182, 193},
	"lightsalmon":          {255, 160, 122},
	"lightseagreen":        {32, 178, 170},
	"lightskyblue":         {135, 206, 250},
	"lightslategray":       {119, 136, 153},
	"lightslategrey":       {119, 136, 153},
	"lightsteelblue":       {176, 196, 222},
	"lightyellow":          {255, 255, 224},
	"lime":                 {0, 255, 0},
	"limegreen":            {50, 205, 50},
	"linen":                {250, 240, 230},
	"magenta":              {255, 0, 255},
	"maroon":               {128, 0, 0},
	"mediumaquamarine":     {102, 205, 170},
	"mediumblue":           {0, 0, 205},
	"mediumorchid":         {186, 85, 211},
	"mediumpurple":         {147, 112, 219},
	"mediumseagreen":       {60, 179, 113},
	"mediumslateblue":      {123, 104, 238},
	"mediumspringgreen":    {0, 250, 154},
	"mediumturquoise":      {72, 209, 204},
	"mediumvioletred":      {199, 21, 133},
	"midnightblue":         {25, 25, 112},
	"mintcream":            {245, 255, 250},
	"mistyrose":            {255, 228, 225},
	"moccasin":             {255, 228, 181},
	"navajowhite":          {255, 222, 173},
	"navy":                 {0, 0, 128},
	"oldlace":              {253, 245, 230},
	"olive":                {128, 128, 0},
	"olivedrab":            {107, 142, 35},
	"orange":               {255, 165, 0},
	"orangered":            {255, 69, 0},
	"orchid":               {218, 112, 214},
	"palegoldenrod":        {238, 232, 170},
	"palegreen":            {152, 251, 152},
	"paleturquoise":        {175, 238, 238},
	"palevioletred":        {219, 112, 147},
	"papayawhip":           {255, 239, 213},
	"peachpuff":            {255, 218, 185},
	"peru":                 {205, 133, 63},
	"pink":                 {255, 192, 203},
	"plum":                 {221, 160, 221},
	"powderblue":           {176, 224, 230},
	"purple":               {128, 0, 128},
	"red":                  {255, 0, 0},
	"rosybrown":            {188, 143, 143},
	"royalblue":            {65, 105, 225},
	"saddlebrown":          {139, 69, 19},
	"salmon":               {250, 128, 114},
	"sandybrown":           {244, 164, 96},
	"seagreen":             {46, 139, 87},
	"seashell":             {255, 245, 238},
	"sienna":               {160, 82, 45},
	"silver":               {192, 192, 192},
	"skyblue":              {135, 206, 235},
	"slateblue":            {106, 90, 205},
	"slategray":            {112, 128, 144},
	"slategrey":            {112, 128, 144},
	"snow":                 {255, 250, 250},
	"springgreen":          {0, 255, 127},
	"steelblue":            {70, 130, 180},
	"tan":                  {210, 180, 140},
	"teal":                 {0, 128, 128},
	"thistle":              {216, 191, 216},
	"tomato":               {255, 99, 71},
	"turquoise":            {64, 224, 208},
	"violet":               {238, 130, 238},
	"wheat":                {245, 222, 179},
	"white":                {255, 255, 255},
	"whitesmoke":           {245, 245, 245},
	"yellow":               {255, 255, 0},
	"yellowgreen":          {154, 205, 50},
}

// Initialize tne gfx server
func (gfx *GFXServer) Init() (int, int) {
	runtime.LockOSThread()
	var rh, rw C.int
	C.init(&rw, &rh)
	gfx.DisplayWidth = int(rw)
	gfx.DisplayHeight = int(rh)
	return int(rw), int(rh)
}

// Shut down the gfx server
func (gfx *GFXServer) Finish() {
	C.finish()
	runtime.UnlockOSThread()
}

// Background clears the screen with the specified solid background color using RGB triples
func (gfx *GFXServer) Background(r, g, b uint8) {
	C.Background(C.uint(r), C.uint(g), C.uint(b))
}

// BackgroundRGB clears the screen with the specified background color using a RGBA quad
func (gfx *GFXServer) BackgroundRGB(r, g, b uint8, alpha VGfloat) {
	C.BackgroundRGB(C.uint(r), C.uint(g), C.uint(b), C.VGfloat(alpha))
}

// BackgroundColor sets the background color
func (gfx *GFXServer) BackgroundColor(s string, alpha ...VGfloat) {
	c := gfx.colorlookup(s)
	if len(alpha) == 0 {
		gfx.BackgroundRGB(c.Red, c.Green, c.Blue, 1)
	} else {
		gfx.BackgroundRGB(c.Red, c.Green, c.Blue, alpha[0])
	}
}

// makestops prepares the color/stop vector
func (gfx *GFXServer) makeramp(r []Offcolor) (*C.VGfloat, C.int) {
	lr := len(r)
	nr := lr * 5
	cs := make([]C.VGfloat, nr)
	j := 0
	for i := 0; i < lr; i++ {
		cs[j] = C.VGfloat(r[i].Offset)
		j++
		cs[j] = C.VGfloat(VGfloat(r[i].Red) / 255.0)
		j++
		cs[j] = C.VGfloat(VGfloat(r[i].Green) / 255.0)
		j++
		cs[j] = C.VGfloat(VGfloat(r[i].Blue) / 255.0)
		j++
		cs[j] = C.VGfloat(r[i].Alpha)
		j++
	}
	return &cs[0], C.int(lr)
}

// FillLinearGradient sets up a linear gradient between (x1,y2) and (x2, y2)
// using the specified offsets and colors in ramp
func (gfx *GFXServer) FillLinearGradient(x1, y1, x2, y2 VGfloat, ramp []Offcolor) {
	cr, nr := gfx.makeramp(ramp)
	C.FillLinearGradient(C.VGfloat(x1), C.VGfloat(y1), C.VGfloat(x2), C.VGfloat(y2), cr, nr)
}

// FillRadialGradient sets up a radial gradient centered at (cx, cy), radius r,
// with a focal point at (fx, fy) using the specified offsets and colors in ramp
func (gfx *GFXServer) FillRadialGradient(cx, cy, fx, fy, radius VGfloat, ramp []Offcolor) {
	cr, nr := gfx.makeramp(ramp)
	C.FillRadialGradient(C.VGfloat(cx), C.VGfloat(cy), C.VGfloat(fx), C.VGfloat(fy), C.VGfloat(radius), cr, nr)
}

// FillRGB sets the fill color, using RGB triples and alpha values
func (gfx *GFXServer) FillRGB(r, g, b uint8, alpha VGfloat) {
	C.Fill(C.uint(r), C.uint(g), C.uint(b), C.VGfloat(alpha))
}

// StrokeRGB sets the stroke color, using RGB triples
func (gfx *GFXServer) StrokeRGB(r, g, b uint8, alpha VGfloat) {
	C.Stroke(C.uint(r), C.uint(g), C.uint(b), C.VGfloat(alpha))
}

// StrokeWidth sets the stroke width
func (gfx *GFXServer) StrokeWidth(w VGfloat) {
	C.StrokeWidth(C.VGfloat(w))
}

// colorlookup returns a RGB triple corresponding to the named color,
// or "rgb(r,g,b)" string. On error, return black.
func (gfx *GFXServer) colorlookup(s string) RGB {
	var rcolor = RGB{0, 0, 0}
	color, ok := colornames[s]
	if ok {
		return color
	}
	if strings.HasPrefix(s, "rgb(") {
		n, err := fmt.Sscanf(s[3:], "(%d,%d,%d)", &rcolor.Red, &rcolor.Green, &rcolor.Blue)
		if n != 3 || err != nil {
			return RGB{0, 0, 0}
		}
		return rcolor
	}
	return rcolor
}

// FillColor sets the fill color using names to specify the color, optionally applying alpha.
func (gfx *GFXServer) FillColor(s string, alpha ...VGfloat) {
	fc := gfx.colorlookup(s)
	if len(alpha) == 0 {
		gfx.FillRGB(fc.Red, fc.Green, fc.Blue, 1)
	} else {
		gfx.FillRGB(fc.Red, fc.Green, fc.Blue, alpha[0])
	}
}

// StrokeColor sets the fill color using names to specify the color, optionally applying alpha.
func (gfx *GFXServer) StrokeColor(s string, alpha ...VGfloat) {
	fc := gfx.colorlookup(s)
	if len(alpha) == 0 {
		gfx.StrokeRGB(fc.Red, fc.Green, fc.Blue, 1)
	} else {
		gfx.StrokeRGB(fc.Red, fc.Green, fc.Blue, alpha[0])
	}
}

// Start begins a picture
func (gfx *GFXServer) Start(w, h int, color ...uint8) {
	C.Start(C.int(w), C.int(h))
	if len(color) == 3 {
		gfx.Background(color[0], color[1], color[2])
	}
}

// Startcolor begins the picture with the specified color background
func (gfx *GFXServer) StartColor(w, h int, color string, alpha ...VGfloat) {
	C.Start(C.int(w), C.int(h))
	gfx.BackgroundColor(color, alpha...)
}

// End ends the picture
func (gfx *GFXServer) End() {
	C.End()
}

// SaveEnd ends the picture, saving the raw raster
func (gfx *GFXServer) SaveEnd(filename string) {
	s := C.CString(filename)
	defer C.free(unsafe.Pointer(s))
	C.SaveEnd(s)
}


func (gfx *GFXServer) CreateImage(filename string) VGImage {
	var t Timer
	t.Start()

	stream, err := os.Open(filename)
	defer stream.Close()
	if err != nil {
		log.Printf("Loading file %v failed. Error: %v\n", filename, err)
	}

	img, _, err := image.Decode(stream)
	if err != nil {
		log.Printf("Decoding file %v failed. Error: %v\n", filename, err)
	}
	r := img.Bounds()
	log.Printf("Loaded & Decoded image %v (%v px) in %v ms.", filename, r.Dx()*r.Dy(), t.TimeSinceLastCall())

	res := gfx.convGO2VGImage(img)
	log.Printf("Converted GO->VG image %v (%v px) in %v ms.", filename, r.Dx()*r.Dy(), t.TimeSinceLastCall())

	return res
}

func (gfx *GFXServer) convGO2VGImage(img image.Image) VGImage {

	// get image parameters
	bounds := img.Bounds()
	minx := C.VGint(bounds.Min.X)
	maxx := C.VGint(bounds.Max.X)
	miny := C.VGint(bounds.Min.Y)
	maxy := C.VGint(bounds.Max.Y)
	w := C.VGint(bounds.Dx())
	h := C.VGint(bounds.Dy())

	// create empty OpenVG image
	vgImage := C.vgCreateImage(C.VG_sABGR_8888, w, h, C.VG_IMAGE_QUALITY_FASTER)

	// convert the GO image to a openVG conform RGBA representation in memory
	data := make([]C.VGubyte, w*h*4)
	n := 0
	var r, g, b, a uint32
	for yp := miny; yp < maxy; yp++ {
		for xp := minx; xp < maxx; xp++ {
			r, g, b, a = img.At(int(xp), int((maxy-1)-yp)).RGBA()
			//fmt.Printf("%v,%v,%v,%v ", r,g,b,a)
			data[n] = C.VGubyte(r >> 8)
			n++
			data[n] = C.VGubyte(g >> 8)
			n++
			data[n] = C.VGubyte(b >> 8)
			n++
			data[n] = C.VGubyte(a >> 8)
			n++
		}
	}

	// initialize the OpenVG image with the data from memory
	C.vgImageSubData(vgImage, unsafe.Pointer(&data[0]), w*4, C.VG_sABGR_8888, minx, miny, w, h)

	return VGImage(vgImage)
}

// Line draws a line between two points
func (gfx *GFXServer) Line(x1, y1, x2, y2 VGfloat, style ...string) {
	C.Line(C.VGfloat(x1), C.VGfloat(y1), C.VGfloat(x2), C.VGfloat(y2))
}

// Rect draws a rectangle at (x,y) with dimesions (w,h)
func (gfx *GFXServer) Rect(x, y, w, h VGfloat, style ...string) {
	C.Rect(C.VGfloat(x), C.VGfloat(y), C.VGfloat(w), C.VGfloat(h))
}

// Rect draws a rounded rectangle at (x,y) with dimesions (w,h).
// the corner radii are at (rw, rh)
func (gfx *GFXServer) Roundrect(x, y, w, h, rw, rh VGfloat, style ...string) {
	C.Roundrect(C.VGfloat(x), C.VGfloat(y), C.VGfloat(w), C.VGfloat(h), C.VGfloat(rw), C.VGfloat(rh))
}

// Ellipse draws an ellipse at (x,y) with dimensions (w,h)
func (gfx *GFXServer) Ellipse(x, y, w, h VGfloat, style ...string) {
	C.Ellipse(C.VGfloat(x), C.VGfloat(y), C.VGfloat(w), C.VGfloat(h))
}

// Circle draws a circle centered at (x,y), with radius r
func (gfx *GFXServer) Circle(x, y, r VGfloat, style ...string) {
	C.Circle(C.VGfloat(x), C.VGfloat(y), C.VGfloat(r))
}

// Qbezier draws a quadratic bezier curve with extrema (sx, sy) and (ex, ey)
// Control points are at (cx, cy)
func (gfx *GFXServer) Qbezier(sx, sy, cx, cy, ex, ey VGfloat, style ...string) {
	C.Qbezier(C.VGfloat(sx), C.VGfloat(sy), C.VGfloat(cx), C.VGfloat(cy), C.VGfloat(ex), C.VGfloat(ey))
}

// Cbezier draws a cubic bezier curve with extrema (sx, sy) and (ex, ey).
// Control points at (cx, cy) and (px, py)
func (gfx *GFXServer) Cbezier(sx, sy, cx, cy, px, py, ex, ey VGfloat, style ...string) {
	C.Cbezier(C.VGfloat(sx), C.VGfloat(sy), C.VGfloat(cx), C.VGfloat(cy), C.VGfloat(px), C.VGfloat(py), C.VGfloat(ex), C.VGfloat(ey))
}

// Arc draws an arc at (x,y) with dimensions (w,h).
// the arc starts at the angle sa, extended to aext
func (gfx *GFXServer) Arc(x, y, w, h, sa, aext VGfloat, style ...string) {
	C.Arc(C.VGfloat(x), C.VGfloat(y), C.VGfloat(w), C.VGfloat(h), C.VGfloat(sa), C.VGfloat(aext))
}

// poly converts coordinate slices
func (gfx *GFXServer) poly(x, y []VGfloat) (*C.VGfloat, *C.VGfloat, C.VGint) {
	size := len(x)
	if size != len(y) {
		return nil, nil, 0
	}
	px := make([]C.VGfloat, size)
	py := make([]C.VGfloat, size)
	for i := 0; i < size; i++ {
		px[i] = C.VGfloat(x[i])
		py[i] = C.VGfloat(y[i])
	}
	return &px[0], &py[0], C.VGint(size)
}

// Polygon draws a polygon with coordinate in x,y
func (gfx *GFXServer) Polygon(x, y []VGfloat, style ...string) {
	px, py, np := gfx.poly(x, y)
	if np > 0 {
		C.Polygon(px, py, np)
	}
}

// Polyline draws a polyline with coordinates in x, y
func (gfx *GFXServer) Polyline(x, y []VGfloat, style ...string) {
	px, py, np := gfx.poly(x, y)
	if np > 0 {
		C.Polyline(px, py, np)
	}
}

// selectfont specifies the font by generic name
func (gfx *GFXServer) selectfont(s string) C.Fontinfo {
	switch s {
	case "sans":
		return C.SansTypeface
	case "serif":
		return C.SerifTypeface
	case "mono":
		return C.MonoTypeface
	}
	return C.SerifTypeface
}

// Text draws text whose aligment begins (x,y)
func (gfx *GFXServer) Text(x, y VGfloat, s string, font string, size int, style ...string) {
	t := C.CString(s)
	C.Text(C.VGfloat(x), C.VGfloat(y), t, gfx.selectfont(font), C.int(size))
	C.free(unsafe.Pointer(t))
}

// TextMid draws text centered at (x,y)
func (gfx *GFXServer) TextMid(x, y VGfloat, s string, font string, size int, style ...string) {
	t := C.CString(s)
	C.TextMid(C.VGfloat(x), C.VGfloat(y), t, gfx.selectfont(font), C.int(size))
	C.free(unsafe.Pointer(t))
}

// TextEnd draws text end-aligned at (x,y)
func (gfx *GFXServer) TextEnd(x, y VGfloat, s string, font string, size int, style ...string) {
	t := C.CString(s)
	C.TextEnd(C.VGfloat(x), C.VGfloat(y), t, gfx.selectfont(font), C.int(size))
	C.free(unsafe.Pointer(t))
}

// TextWidth returns the length of text at a specified font and size
func (gfx *GFXServer) TextWidth(s string, font string, size int) VGfloat {
	t := C.CString(s)
	defer C.free(unsafe.Pointer(t))
	return VGfloat(C.TextWidth(t, gfx.selectfont(font), C.int(size)))
}

// Translate translates the coordinate system to (x,y)
func (gfx *GFXServer) Translate(x, y VGfloat) {
	C.Translate(C.VGfloat(x), C.VGfloat(y))
}

// Rotate rotates the coordinate system around the specifed angle
func (gfx *GFXServer) Rotate(r VGfloat) {
	C.Rotate(C.VGfloat(r))
}

// Shear warps the coordinate system by (x,y)
func (gfx *GFXServer) Shear(x, y VGfloat) {
	C.Shear(C.VGfloat(x), C.VGfloat(y))
}

// Scale scales the coordinate system by (x,y)
func (gfx *GFXServer) Scale(x, y VGfloat) {
	C.Scale(C.VGfloat(x), C.VGfloat(y))
}

// SaveTerm saves terminal settings
func (gfx *GFXServer) SaveTerm() {
	C.saveterm()
}

// RestoreTerm retores terminal settings
func (gfx *GFXServer) RestoreTerm() {
	C.restoreterm()
}

// func RawTerm() sets the terminal to raw mode
func (gfx *GFXServer) RawTerm() {
	C.rawterm()
}
