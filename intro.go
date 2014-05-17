package main

import (
	"github.com/flothe/pinboard/grafic2d"
)

type Intro struct {
	filename string
	sprite   *grafic2d.Sprite
	gfx      *grafic2d.GFXServer
	waitForEndMillis int
	timerMessageShown grafic2d.Timer
	isReady bool
	
}

func NewIntroFromGif(filename string, waitForEndMillis int) *Intro {
	intro := Intro{}
	intro.filename = filename
	intro.waitForEndMillis = waitForEndMillis
	intro.load()
	intro.isReady = false
	return &intro
}

func (intro *Intro) IsReadyToEnd() bool {
	if intro.sprite == nil {
		return false
	}
	return intro.sprite.IsAnimationEnded()
}

func (intro *Intro) IsReady() bool {
	return intro.isReady
}

func (intro *Intro) load() error {
	var err error
	if intro.sprite == nil {
		intro.sprite, err = grafic2d.NewSpriteFromGif(intro.filename)
		intro.sprite.AnimDuration = intro.sprite.AnimDuration + intro.waitForEndMillis				
		intro.sprite.DoNotLoop = true
	}
	return err
}

func (intro *Intro) Begin(gfx *grafic2d.GFXServer) error {
	var err error
	intro.isReady = true
	intro.gfx = gfx
	intro.timerMessageShown.Start()
	intro.load()
	if intro.sprite != nil {
		intro.sprite.DoNotLoop = true
		intro.sprite.CenterAndFitToScreen(gfx.DisplayWidth, gfx.DisplayHeight)
		//intro.sprite.ScaleToScreen(gfx.DisplayWidth, gfx.DisplayHeight)
	}
	return err
}

func (intro *Intro) End() error {
	intro.isReady = false
	intro.sprite.Reset()
	intro.timerMessageShown.Reset()
	return nil
}


func (intro *Intro) Destroy() error {
	if intro.sprite != nil {
		intro.sprite.Destroy()		
	}
	intro.isReady = false
	return nil
}


func (intro *Intro) Update(ms int) error {
	if intro.sprite != nil {
		intro.sprite.Update(ms)		
	}
	return nil
}

func (intro *Intro) Draw() error {
	if intro.sprite != nil {
		intro.sprite.Draw()		
	}
	return nil
}

func (intro *Intro) GetMsgShowTime() int {
	return intro.timerMessageShown.TimeSinceStart()
}

