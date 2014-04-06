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
	"github.com/gosexy/canvas"
	"log"
	"os"
	"bytes"
	"unsafe"
	"image"
	"encoding/binary"	
)

func NewVGImageFromPaletted(img *image.Paletted) (VGImage, error) {
	w := img.Rect.Dx()
	h := img.Rect.Dy()

	// create empty OpenVG image
	vgImage := C.vgCreateImage(C.VG_sABGR_8888, C.VGint(w), C.VGint(h), C.VG_IMAGE_QUALITY_FASTER)

	// convert the GO image to a openVG conform RGBA representation in memory
	data := make([]C.VGubyte, w*h*4)
	n := 0
	var r, g, b, a uint32
	for yp := 0; yp < h; yp++ {
		for xp := 0; xp < w; xp++ {
			index := img.Pix[yp*img.Stride + xp]			
			r, g, b, a = img.Palette[index].RGBA()
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
	C.vgImageSubData(vgImage, unsafe.Pointer(&data[0]), C.VGint(w*4), C.VG_sABGR_8888, 0, 0, C.VGint(w), C.VGint(h))

	return VGImage(vgImage), nil	
	
}


func SaveScaledVGImage(buf []byte, maxWidth, maxHeight uint, fn string) error {
	log.Printf("Enter SaveScaledVGImage: %v. ", fn)
	var t Timer
	t.Start()
	
	// Opening some image from disk.
	img := canvas.New()
	defer img.Destroy()
	
		
	err := img.OpenBlob(buf, uint(len(buf)))
	if err != nil {
		log.Printf("Failed to open image %v. Error: %v (%v)\n", fn, err, img.Error())
	}
	
	err = img.SetBackgroundColor("orange")
	if err != nil {
		log.Printf("Failed to set backround color for image %v. Error: %v (%v)\n", fn, err, img.Error())
	}
	// Photo auto orientation based on EXIF tags.
	img.AutoOrientate()
	// Flip because VGImages are upside down
	img.Flip()

	log.Printf("Loaded & Decoded image %v (%v px) in %v ms.\n", fn, img.Width()*img.Height(), t.TimeSinceLastCall())

	// calculate the new size, keeping aspect ratio
	imgH := float64(img.Height())
	imgW := float64(img.Width())
	maxH := float64(maxHeight)
	maxW := float64(maxWidth)
	// check if we need to shrink the image
	if imgH>maxH || imgW>maxW {
		// determine the factor of the shrinking
		factor := maxH/imgH
		factor2 := maxW/imgW
		if factor > factor2 {
			factor = factor2
		}
		// Shrink the image
		img.ResizeWithFilter(uint(factor*imgW), uint(factor*imgH), canvas.CATROM_FILTER, 1.0)
		log.Printf("Resized image %v %vx%v --> %vx%v in %v ms.\n", fn, uint(imgW), uint(imgH), img.Width(), img.Height(), t.TimeSinceLastCall())		
	}

	err = img.SetOption("-depth", "8")
	if err != nil {
		log.Printf("Failed to set depth to 8 for %v. Error: %v (%v)\n", fn, err, img.Error())
	}
	
	err = img.SetOption("define", "quantum:format=unsigned")
	if err != nil {
		log.Printf("Failed to set quantum:format=unsigned for %v. Error: %v (%v)\n", fn, err, img.Error())
	}

	err = img.SetFormat("RGBA")
	if err != nil {
		log.Printf("Failed to set RGBA format for %v. Error: %v (%v)\n", fn, err, img.Error())
	}

	log.Printf("Changed image format %v to %v in %v ms.\n", fn, img.GetFormat(), t.TimeSinceLastCall())

	b, err := img.GetImageBlob()
	if err != nil {
		log.Printf("Failed to get image blob for %v. Error: %v (%v)\n", fn, err, img.Error())
	}
	log.Printf("Got image blob %v (%v bytes) in %v ms.\n", fn, len(b), t.TimeSinceLastCall())
	
	// open output file
	fo, err := os.Create(fn)
	if err != nil {
		return err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			log.Println("fo.Close failed (we start to panic):", err)
			panic(err)
		}
	}()

	// write width and height to file
	dimBuf := new(bytes.Buffer)
	err = binary.Write(dimBuf, binary.LittleEndian, uint64(img.Width()))
	if err==nil {
		err = binary.Write(dimBuf, binary.LittleEndian, uint64(img.Height()))		
	}
	if err != nil {
		log.Println("binary.Write failed:", err)
		return err
	}
	if _, err = fo.Write(dimBuf.Bytes()); err != nil {
		log.Println("fo.Write of width x height failed:", err)
		return err
	}
	log.Printf("Writen image dimension (%vx%v) to %v\n", uint64(img.Width()), uint64(img.Height()), fn)
	

	if _, err = fo.Write(b); err != nil {
		log.Println("fo.Write of RGBA bytes failed:", err)
		return err
	}
	
	log.Printf("Quit SavedScaledImage: Finished writing image %v to disk in %v ms.\n", fn, t.TimeSinceLastCall())
	return nil
}


func LoadVGImage(fn string) (VGImage, error) {
	// open output file
	fo, err := os.Open(fn)
	if err != nil {
		return VGImage(0), err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	// read width and heigth from file
	var w, h uint64
	wh := make([]byte, 16)
	if _, err := fo.Read(wh[:len(wh)]); err != nil {
		log.Printf("Failed to read image width x height from file %v: %v", fn, err)
		return VGImage(0), err
	}
	buf := bytes.NewReader(wh)
	err = binary.Read(buf, binary.LittleEndian, &w)
	if err==nil {
		err = binary.Read(buf, binary.LittleEndian, &h)		
	}
	if err != nil {
		log.Println("binary.Read failed:", err)
	}
	
	log.Printf("Loaded dimension of image %v: %vx%v", fn, w, h)

	// convert the GO image to a openVG conform RGBA representation and write it to the file
	data := make([]byte, 4*w*h)
	if _, err := fo.Read(data[:len(data)]); err != nil {
		log.Printf("Failed to read image RGBA data from file %v: %v", fn, err)
		return VGImage(0), err
	}

	// initialize the OpenVG image with the data from memory
	// create empty OpenVG image
	vgImage := C.vgCreateImage(C.VG_sABGR_8888, C.VGint(w), C.VGint(h), C.VG_IMAGE_QUALITY_FASTER)
	C.vgImageSubData(vgImage, unsafe.Pointer(&data[0]), C.VGint(w*4), C.VG_sABGR_8888, 0, 0, C.VGint(w), C.VGint(h))

	return VGImage(vgImage), nil

}


func (img VGImage) Destroy() {
	C.vgDestroyImage(C.VGImage(img))
}

func (img VGImage) Width() VGint {
	return VGint(C.vgGetParameteri(C.VGHandle(img), C.VG_IMAGE_WIDTH))
}

func (img VGImage) Height() VGint {
	return VGint(C.vgGetParameteri(C.VGHandle(img), C.VG_IMAGE_HEIGHT))
}

func (img VGImage) Format() VGImageFormat {
	return VGImageFormat(C.vgGetParameteri(C.VGHandle(img), C.VG_IMAGE_FORMAT))
}


func (img VGImage) Draw(tx, ty, sx, sy, rx, ry, rdeg VGfloat) {
	//	w, h := C.VGint(img.Width()), C.VGint(img.Height())
	//	C.vgSetPixels(C.VGint(tx), C.VGint(ty), C.VGImage(*img), 0, 0, w, h)

	// store the current transformation
	var oldmatrix [9]C.VGfloat
	C.vgGetMatrix(&oldmatrix[0])
	oldImgMode := C.vgGeti(C.VG_IMAGE_MODE)
	oldMatMode := C.vgGeti(C.VG_MATRIX_MODE)

	// set the transformation for this image
	C.vgSeti(C.VG_IMAGE_MODE, C.VG_DRAW_IMAGE_NORMAL)
	C.vgSeti(C.VG_MATRIX_MODE, C.VG_MATRIX_IMAGE_USER_TO_SURFACE)

	C.vgLoadIdentity()
	if rdeg != 0 {
		C.vgTranslate(C.VGfloat(VGfloat(img.Width()/2)+rx), C.VGfloat(VGfloat(img.Height()/2)+ry))
		C.vgRotate(C.VGfloat(rdeg))
		C.vgTranslate(C.VGfloat(VGfloat(-img.Width()/2)-rx), C.VGfloat(VGfloat(-img.Height()/2)-ry))
	}
	if tx != 0 || ty != 0 {
		C.vgTranslate(C.VGfloat(tx), C.VGfloat(ty))
	}
	if sx != 1 || sy != 1 {
		C.vgScale(C.VGfloat(sx), C.VGfloat(sy))
	}

	// draw the image to the current drawing surface
	C.vgDrawImage(C.VGImage(img))

	// restore the old transormation
	C.vgSeti(C.VG_IMAGE_MODE, oldImgMode)
	C.vgSeti(C.VG_MATRIX_MODE, oldMatMode)
	C.vgLoadMatrix(&oldmatrix[0])
}
