package grafic2d

import (
	"log"
	"math"
	"fmt"
)

type PhotoSlider struct {
	filenames []string
	images []*Sprite
	gfx *GFXServer
	time int	
	slideTimeMs int // how long is a image shown in milliseconds
	minShowTimeMs int // how long should this slider be shown at least in milliseconds
	imageIndex int
}

func NewPhotoSlider(fn []string, slideTimeMs, minShowTimeMs int) *PhotoSlider {
	ps := PhotoSlider{filenames:fn,slideTimeMs:slideTimeMs,minShowTimeMs:minShowTimeMs}
	return &ps
}

func (ps *PhotoSlider) IsReadyToEnd() bool {
	// if the minShowTime has not exceeded return false
	if ps.time < ps.minShowTimeMs {
		return false
	}
	// return true if all images have been shown for ps.slideTimeMs
	return int(ps.time / ps.slideTimeMs) >= len(ps.images)
}



func (ps *PhotoSlider) Begin(gfx *GFXServer) error {
	ps.gfx = gfx
	ps.time = 0
	ps.imageIndex = 0
	
	// default value for slide time is 8 secs
	if ps.slideTimeMs<=0 {
		ps.slideTimeMs = 8000
	}

	var t Timer
	t.Start()
	// load the images
	for _, fn := range ps.filenames {
		vgImg, err := LoadVGImage(fn)
		if err != nil {
			return err
		}
		s := NewSprite()
		s.AddImg(vgImg, 0)	
		s.AnimDuration = ps.slideTimeMs
		s.DoNotLoop = true
		ps.images = append(ps.images, s)	
	}
	log.Printf("Loaded %v images in %v ms.\n", len(ps.images), t.TimeSinceStart())
	
	ps.centerAllImages()
	//ps.animateHorizontalAllImages()
	
	return nil
}

func (ps *PhotoSlider) End() error {
	// free all space used by the images
	var t Timer
	t.Start()
	for _, img := range ps.images {
		img.Destroy()
	}
	log.Printf("Destroyed %v images in %v ms.\n", len(ps.images), t.TimeSinceStart())
	ps.images = nil
	
	return nil
}


func (ps *PhotoSlider) Update(ms int) error {
	ps.time = ps.time + ms
	
	ps.updateImageIndex()
	if ps.imageIndex >= len(ps.images) {
		return fmt.Errorf("Failed to update image.")
	}
	ps.images[ps.imageIndex].Update(ms)
	
	return nil
	
}

func (ps *PhotoSlider) Draw() error {
	if ps.imageIndex >= len(ps.images) {
		return fmt.Errorf("Failed to draw image.")
	}
	ps.images[ps.imageIndex].Draw()
	
	return nil
}


func (ps *PhotoSlider) updateImageIndex() {
	
	if(len(ps.images)==0) {
		return		
	}

	if(ps.slideTimeMs==0) {
		return		
	}
	
	ix := int(ps.time / ps.slideTimeMs) % len(ps.images)
	if ix != ps.imageIndex {
		ps.images[ps.imageIndex].Reset()
		log.Printf("Switch images %v --> %v \n",ps.imageIndex, ix)		
		ps.imageIndex = ix		
	}
}

func (ps *PhotoSlider) centerAllImages() error {
	for _, img := range ps.images {
		img.CenterAndFitToScreen(ps.gfx.DisplayWidth, ps.gfx.DisplayHeight)
	}
	return nil	
}

func (ps *PhotoSlider) animateHorizontalAllImages() error {
	for _, img := range ps.images {
		// scale to fit
		factor := math.Min(float64(ps.gfx.DisplayWidth)/float64(img.Width), float64(ps.gfx.DisplayHeight)/float64(img.Height))
		s := AnimLinear{}
		s.AddFrame(0, []float32{float32(factor*0.4), float32(factor*0.4)})
		s.AddFrame(2000, []float32{float32(factor), float32(factor)})
		s.AddFrame(ps.slideTimeMs-2000, []float32{float32(factor), float32(factor)})
		s.AddFrame(ps.slideTimeMs, []float32{float32(factor*0.4), float32(factor*0.4)})
		log.Printf("Scale image (%vx%v px) for display (%vx%v px): %v", img.Width, img.Height, ps.gfx.DisplayWidth, ps.gfx.DisplayHeight, factor)		
		img.SetAnimScale(&s)
	
		// center vertical, animate horizontal
		t := AnimLinear{}
		w := float64(img.Width) * factor
		h := float64(img.Height) * factor
		dx := (float64(ps.gfx.DisplayWidth) - w) / 2.0
		dy := (float64(ps.gfx.DisplayHeight) - h) / 2.0 
		log.Printf("Center image (%vx%v px) for display (%vx%v px): %v, %v", w, h, ps.gfx.DisplayWidth, ps.gfx.DisplayHeight, dx, dy)		
		t.AddFrame(0, []float32{float32(-w), float32(ps.gfx.DisplayHeight/2)})
		t.AddFrame(2000, []float32{float32(dx), float32(dy)})
		t.AddFrame(ps.slideTimeMs-2000, []float32{float32(dx), float32(dy)})
		t.AddFrame(ps.slideTimeMs, []float32{float32(ps.gfx.DisplayWidth), float32(ps.gfx.DisplayHeight/2)})
		img.SetAnimTrans(&t)
	}
	return nil	
}

