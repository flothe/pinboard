package grafic2d

import (
	"image/gif"
	"os"
	"log"
	"fmt"
	"math"
)


type SpriteImg struct {
	img VGImage
	ms  int
}

type Sprite struct {
	images                           []SpriteImg
	time                             int
	AnimDuration                     int
	animTrans, animScale, animRotate Animation
	Width, Height                    int
	DoNotLoop						bool
	ignoreFirstUpdateMillis	bool
}

func NewSprite() *Sprite {
	return new(Sprite)
}


func NewSpriteFromGif(filename string) (*Sprite, error) {
	// open output file
	fo, err := os.Open(filename)
	if err != nil {
		log.Printf("Failed to open %v (%v).\n", filename, err)
		wd, _ := os.Getwd()
		log.Printf("Current working dir: %v\n", wd)
		return nil, err
	}
	// close fo on exit and check for its returned error
	defer func() {
		if err := fo.Close(); err != nil {
			panic(err)
		}
	}()

	gifImg, err := gif.DecodeAll(fo)
	if err != nil {
		log.Printf("Failed to decode GIF: %v (%v).\n", filename, err)
		return nil, err
	}
	
	if len(gifImg.Image) != len(gifImg.Delay) {
		return nil, fmt.Errorf("%v contains more images (%v) than delay times (%v).", filename, len(gifImg.Image), len(gifImg.Delay))
	}
	
	s := NewSprite()
	
	ms:=0
	for i := 0; i < len(gifImg.Image); i++ {
		vgImg, err := NewVGImageFromPaletted(gifImg.Image[i])
		if err != nil {
			log.Printf("Failed to create VGImage from Paletted image: %v (%v).\n", filename, err)
			return nil, err
		}
		s.AddImg(vgImg, ms)
		ms = ms + gifImg.Delay[i] * 10
	}
	s.AnimDuration = ms
	s.ignoreFirstUpdateMillis = true
	
	return s, nil
}




// initialize the sprite
func (s *Sprite) Init() {
	// set default values
	s.AnimDuration = 0
	s.time = 0
	s.ignoreFirstUpdateMillis = true
}

// initialize the sprite
func (s *Sprite) Reset() {
	s.time = 0
	s.ignoreFirstUpdateMillis = true
}

// clean up the sprite and free all resources
func (s *Sprite) Destroy() {
	for i := 0; i < len(s.images); i++ {
		s.images[i].img.Destroy()
	}
	s.Width = 0
	s.Height = 0
	s.ignoreFirstUpdateMillis = true
}

// add image to the sprite animation
func (s *Sprite) AddImg(image VGImage, ms int) {
	sImg := SpriteImg{image, ms}
	s.images = append(s.images, sImg)
	w, h := int(image.Width()), int(image.Height())

	if w > s.Width {
		s.Width = w
	}
	if h > s.Height {
		s.Height = h
	}

	//fmt.Printf("The sprite holds %v images\n", len(s.images))
}

// add a translation animation
func (s *Sprite) SetAnimTrans(anim Animation) {
	s.animTrans = anim
}

// add a scale animation
func (s *Sprite) SetAnimScale(anim Animation) {
	s.animScale = anim
}

// add a rotation animation
func (s *Sprite) SetAnimRotate(anim Animation) {
	s.animRotate = anim
}

func (s *Sprite) CenterAndFitToScreen(displayWidth, displayHeight int) {
	// scale to fit
	factor := math.Min(float64(displayWidth)/float64(s.Width), float64(displayHeight)/float64(s.Height))
	sanim := AnimLinear{}
	sanim.AddFrame(0, []float32{float32(factor), float32(factor)})
	log.Printf("Scale image (%vx%v px) for display (%vx%v px): %v", s.Width, s.Height, displayWidth, displayHeight, factor)		
	s.SetAnimScale(&sanim)

	// center
	tanim := AnimLinear{}
	w := float64(s.Width) * factor
	h := float64(s.Height) * factor
	dx := (float64(displayWidth) - w) / 2.0
	dy := (float64(displayHeight) - h) / 2.0 
	log.Printf("Center image (%vx%v px) for display (%vx%v px): %v, %v", w, h, displayWidth, displayHeight, dx, dy)		
	tanim.AddFrame(0, []float32{float32(dx), float32(dy)})
	s.SetAnimTrans(&tanim)	
}

func (s *Sprite) ScaleToScreen(displayWidth, displayHeight int) {
	// scale to fit
	sanim := AnimLinear{}
	sanim.AddFrame(0, []float32{float32(displayWidth)/float32(s.Width), float32(displayHeight)/float32(s.Height)})
	s.SetAnimScale(&sanim)
}


func (s *Sprite) Update(ms int) {
	if s.ignoreFirstUpdateMillis {
		s.ignoreFirstUpdateMillis = false
		return
	}
	

	s.time = s.time + ms
}

func (s *Sprite) Draw() {
	img := s.calcImage(s.time)
	x, y, sx, sy := float32(0), float32(0), float32(1), float32(1)
	rx, ry, rdegree := float32(0), float32(0), float32(0)

	if s.animTrans != nil {
		t := s.animTrans.Value(s.time)
		x, y = t[0], t[1]
	}
	if s.animScale != nil {
		s := s.animScale.Value(s.time)
		sx, sy = s[0], s[1]
	}
	if s.animRotate != nil {
		r := s.animRotate.Value(s.time)
		rdegree, rx, ry = r[0], r[1], r[2]
	}
	img.Draw(VGfloat(x), VGfloat(y), VGfloat(sx), VGfloat(sy), VGfloat(rx), VGfloat(ry), VGfloat(rdegree))
}

func (s *Sprite) IsAnimationEnded() bool {
	if s.DoNotLoop {
		return s.time > s.AnimDuration		
	}
	return false
}


func (s *Sprite) calcImage(ms int) *VGImage {

	if s.AnimDuration == 0 {
		return &s.images[0].img
	}
	
	millis := ms
	if !s.DoNotLoop {
		// loop the milliseconds
		millis = ms % s.AnimDuration		
	}

	numImg := 0
	for (numImg + 1) < len(s.images) {
		if s.images[numImg+1].ms < millis {
			numImg++
		} else {
			break
		}
	}

	return &s.images[numImg].img
}
