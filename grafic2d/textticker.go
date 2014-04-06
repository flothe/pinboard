package grafic2d

import (
//	"log"
)

type TextTicker struct {
	// ticker text stuff
	tickerText string
	tickerTextPos1, tickerTextPos2 VGfloat
	tickerTextY VGfloat
	tickerTextWidth VGfloat
	tickerTextSpeed int   // pixel per Second
	// font stuff
	font string
	fontSize int
	
	tickerPrefix string
	tickerPrefixWidth VGfloat
	
	dateText string
	sourceText string
	gfx *GFXServer
	time int	
}

func NewTextTicker(tickerText, tickerPrefix, dateText, sourceText string) *TextTicker {
	tt := TextTicker{tickerText:tickerText + "   +++   ", tickerPrefix:tickerPrefix, dateText:dateText, sourceText:sourceText}
	tt.font = "sans"
	tt.fontSize = 20
	tt.tickerTextY = 100
	tt.tickerTextSpeed = 150
	return &tt
}


func (tt *TextTicker) Begin(gfx *GFXServer) error {
	tt.gfx = gfx
	tt.time = 0
	// assure that the ticker text is longer than the display width	
	tt.tickerTextWidth = tt.gfx.TextWidth(tt.tickerText, tt.font, tt.fontSize)
	for tt.tickerTextWidth < VGfloat(tt.gfx.DisplayWidth) {
		tt.tickerText = tt.tickerText +  tt.tickerText
		tt.tickerTextWidth = tt.gfx.TextWidth(tt.tickerText, tt.font, tt.fontSize)
	}
	tt.tickerTextPos1 = VGfloat(tt.gfx.DisplayWidth)
	tt.tickerTextPos2 = VGfloat(tt.gfx.DisplayWidth) + tt.tickerTextWidth		

	tt.tickerPrefixWidth = tt.gfx.TextWidth(tt.tickerPrefix, tt.font, int(VGfloat(tt.fontSize) * 1.5))

	
	return nil
}

func (tt *TextTicker) End() error {
	return nil
}


func (tt *TextTicker) Update(ms int) error {
	tt.time = tt.time + ms
	
	// calculate ticker text postions
	diff := VGfloat(ms*tt.tickerTextSpeed)/1000.0
	tt.tickerTextPos1 = tt.tickerTextPos1 - diff
	if tt.tickerTextPos1 < -tt.tickerTextWidth {
		tt.tickerTextPos1 = tt.tickerTextPos2 + tt.tickerTextWidth
	}
	tt.tickerTextPos2 = tt.tickerTextPos2 - diff
	if tt.tickerTextPos2 < -tt.tickerTextWidth {
		tt.tickerTextPos2 = tt.tickerTextPos1 + tt.tickerTextWidth
	}
	
	return nil
}

func (tt *TextTicker) Draw() error {
	
	// draw ticker background
	tt.gfx.FillRGB(0,0,64,0.8)
	height := 3.0*VGfloat(tt.fontSize)
	bottom := tt.tickerTextY-VGfloat(tt.fontSize)
	tt.gfx.Rect(0.0, bottom, VGfloat(tt.gfx.DisplayWidth), height)

	tt.gfx.FillRGB(255,255,255,1.0)	
	// if visible render ticker text 1
	if(tt.tickerTextPos1 < VGfloat(tt.gfx.DisplayWidth)) {
		tt.gfx.Text(tt.tickerTextPos1, tt.tickerTextY ,tt.tickerText, tt.font, tt.fontSize)		
	}
	// if visible render ticker text 2
	if(tt.tickerTextPos2 < VGfloat(tt.gfx.DisplayWidth)) {
		tt.gfx.Text(tt.tickerTextPos2, tt.tickerTextY ,tt.tickerText, tt.font, tt.fontSize)		
	}

	// draw prefix
	tt.gfx.FillRGB(0,0,128,1.0)
	tt.gfx.Rect(0.0, bottom, tt.tickerPrefixWidth + 40, height)
	tt.gfx.FillRGB(255,255,255,1.0)	
	tt.gfx.Text(20, tt.tickerTextY - 0.25 * VGfloat(tt.fontSize) ,tt.tickerPrefix, tt.font, int(VGfloat(tt.fontSize) * 1.5))		
	
	// draw date and author background
	fs := int(VGfloat(tt.fontSize)*0.85)
	height = 3*VGfloat(fs)
	bottom = bottom - height
	textY := bottom + height/2.0 - VGfloat(fs)/2.0
	tt.gfx.FillRGB(0,0,64,0.95)
	tt.gfx.Rect(0.0, bottom, VGfloat(tt.gfx.DisplayWidth), height)
	// draw date and author
	tt.gfx.FillRGB(255,255,255,1.0)	
	tt.gfx.Text(20, textY ,tt.dateText, tt.font, fs)		
	tt.gfx.TextEnd(VGfloat(tt.gfx.DisplayWidth-20), textY ,tt.sourceText, tt.font, fs)		
	
	return nil
}

