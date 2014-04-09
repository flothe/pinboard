package main

import (
	"github.com/flothe/pinboard/grafic2d"
	"github.com/flothe/pinboard/web"
	"os"
	"strings"
	"log"
	"strconv"
	"bytes"
	"math/rand"
	"fmt"
)

type PinMessage interface {
	IsReadyToEnd() bool 
	// Start the rendering of this message. After this call you may use Update() and Draw()
	Begin(gfx *grafic2d.GFXServer) error 
	// End the rendering of this message. After this call Update() and Draw() is not possible
	End() error 
	// Update the message 
	Update(ms int) error 
	// Draw the message
	Draw() error 
	// Ready to Update and Draw?
	IsReady() bool
	// How long since Begin() has been called
	GetMsgShowTime() int
	// is called when the PinMessage is discarded to clean up
	Destroy() error
}


type Pinboard struct {
	msgs []PinMessage
	msgIndex int
	gfx *grafic2d.GFXServer	
	s *grafic2d.Sprite	
	debugTimerFps grafic2d.Timer
}

func (pb *Pinboard) AddMessage(msg PinMessage) {
	// add intro and message
	filename := fmt.Sprintf("internal/m%v.gif", (rand.Int()%5)+1)
	pb.msgs = append(pb.msgs, PinMessage(NewIntroFromGif(filename, 1000)), msg)
}

func (pb *Pinboard) AddMessageData(data *web.MessageData) {
	m := PinMessage(NewMessage(data.ShortText, data.SenderName, data.Timestamp, data.ImageNames))
	pb.AddMessage(m)
}

func (pb *Pinboard) LoadMessages() error {
	// open directory
	 d, err := os.Open("." + string(os.PathSeparator))
     if err != nil {
		 log.Println("Failed to open base directory: ", err)
		 return err
     }
     defer d.Close()
	 // read all entries in the directory
     fi, err := d.Readdir(-1)
     if err != nil {
		 log.Println("Failed to read content of base directory: ", err)
		 return err
     }
	 // cylce over all directory entries and load the .cmsg files
     for _, fi := range fi {
         if fi.Mode().IsRegular() && strings.HasSuffix(fi.Name(), ".cmsg") {
		 	data := new(web.MessageData)
		 	err := data.Load(fi.Name())
			if err != nil {
				 log.Println("Failed to load message from file: ", fi)
				 return err
		     }
			 log.Println("Loaded message from file: ", fi)	
		 	m := PinMessage(NewMessage(data.ShortText, data.SenderName, data.Timestamp, data.ImageNames))
			pb.AddMessage(m)		 
         }
     }
	 return nil
}

func (pb *Pinboard) Begin(gfx *grafic2d.GFXServer) error {
	pb.gfx = gfx
	pb.debugTimerFps.Start()	
	return nil
}


func (pb *Pinboard) Update(ms int) error {
	// if there is no message to update do nothing and return
	if(pb.msgIndex>=len(pb.msgs)) {
		return nil
	}
	
	// if the current message is not ready: call begin
	if !pb.msgs[pb.msgIndex].IsReady() {
		pb.msgs[pb.msgIndex].Begin(pb.gfx)
	}
		
	// if the current message is finished
	if(pb.msgs[pb.msgIndex].IsReadyToEnd()) {
		// unload the resources of the old message
		log.Printf("Presentation of message %v is finished.", pb.msgIndex)
		pb.msgs[pb.msgIndex].End()
		// get the next message and begin it
		pb.msgIndex++
		pb.msgIndex = pb.msgIndex % len(pb.msgs)
		log.Printf("Switch to message %v.\n", pb.msgIndex)
		pb.msgs[pb.msgIndex].Begin(pb.gfx)
	}
	
	// update the current message
	pb.msgs[pb.msgIndex].Update(ms)
	
	return nil
}

func (pb *Pinboard) Draw() error {

	var w, h int
	pb.gfx.Start(w, h)
	pb.gfx.Background(0, 0, 0)


	// paint the current message
	if(pb.msgIndex<len(pb.msgs)) {
		pb.msgs[pb.msgIndex].Draw()
	}
	
	pb.drawDebugInfo()
	
	pb.gfx.End()
	
	return nil
}


func (pb *Pinboard) drawDebugInfo() {
	var buffer bytes.Buffer
	
	// draw fps
	buffer.WriteString("fps=")
	buffer.WriteString(strconv.Itoa(pb.debugTimerFps.CallsPerSec()))
	pb.gfx.FillColor("white")       // White text
	pb.gfx.Text(20, grafic2d.VGfloat(pb.gfx.DisplayHeight-30), buffer.String(), "serif", 20)
	buffer.Reset()	

	// draw message info
	if len(pb.msgs)>0 {
		buffer.WriteString("Message ")
		buffer.WriteString(strconv.Itoa(pb.msgIndex+1))
		buffer.WriteString(" of ")
		buffer.WriteString(strconv.Itoa(len(pb.msgs)))
		buffer.WriteString(": ")
		buffer.WriteString(strconv.Itoa(pb.msgs[pb.msgIndex].GetMsgShowTime()))
		buffer.WriteString(" ms")	
		pb.gfx.FillColor("white")       // White text
		pb.gfx.Text(20, grafic2d.VGfloat(pb.gfx.DisplayHeight-60), buffer.String(), "serif", 20)
		buffer.Reset()	
	}
}


func (pb *Pinboard) End() {
	// end the current message
	if(pb.msgIndex<len(pb.msgs)) {
		pb.msgs[pb.msgIndex].Destroy()
	}
	
}

