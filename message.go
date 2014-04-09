package main

import (
	"github.com/flothe/pinboard/grafic2d"
	"log"
	"time"
	"fmt"
)

type Message struct {
	photos   *grafic2d.PhotoSlider
	text     *grafic2d.TextTicker
	timestamp time.Time
	time     int
	gfx      *grafic2d.GFXServer
	isReady bool
	timerMessageShown grafic2d.Timer
}

func NewMessage(text, from string, timestamp time.Time, fnPhotos []string) *Message {
	msg := Message{}
	
	msg.photos = grafic2d.NewPhotoSlider(fnPhotos, 10000, 9000)
	msg.text = grafic2d.NewTextTicker(text, "NEWS", timestamp.Format("2. Jan 06 - 15:04"), from)
	return &msg
}

func (msg *Message) IsReadyToEnd() bool {
	if msg.photos != nil {
		return msg.photos.IsReadyToEnd()			
	}
	return true
}


func (msg *Message) IsReady() bool {
	return msg.isReady
}

func (msg *Message) GetMsgShowTime() int {
	return msg.timerMessageShown.TimeSinceStart()
}


func (msg *Message) Begin(gfx *grafic2d.GFXServer) error {
	var err error
	msg.gfx = gfx
	if msg.photos != nil {
		err = msg.photos.Begin(gfx)			
	}
	if msg.text != nil {
		err = msg.text.Begin(gfx)			
	}
	msg.isReady = true
	msg.timerMessageShown.Start()
	return err
}

func (msg *Message) End() error {
	var err error
	if msg.photos != nil {
		err = msg.photos.End()			
	}
	if msg.text != nil {
		err = msg.text.End()			
	}
	msg.time = 0
	msg.isReady = false
	msg.timerMessageShown.Reset()
	return err
}

func (msg *Message) Destroy() error {
	return nil
}



func (msg *Message) Update(ms int) error {
	var err error
	if(!msg.isReady) {
		err = fmt.Errorf("Message is not ready for update: %v", msg)
		log.Println(err)
		return err
	}

	msg.time = msg.time + ms

	if msg.photos != nil {
		err = msg.photos.Update(ms)			
	}
	if msg.text != nil {
		err = msg.text.Update(ms)			
	}
	return err
}

func (msg *Message) Draw() error {
	var err error
	if(!msg.isReady) {
		err = fmt.Errorf("Message is not ready for draw: %v", msg)
		log.Println(err)
		return err
	}

	// draw the photos
	if msg.photos != nil {
		err = msg.photos.Draw()			
	}
	// draw the text
	if msg.text != nil {
		err = msg.text.Draw()			
	}
	return err
}
