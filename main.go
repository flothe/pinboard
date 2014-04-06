package pinboard

import (
	"bytes"
	"git.maibornwolff.de/pinboard/grafic2d"
	"git.maibornwolff.de/pinboard/web"
	"fmt"
	"strconv"
	"time"
	"bufio"
	"os"
	"log"
)

func main() {
	
	log.SetFlags(log.Ldate|log.Ltime|log.Lshortfile)
	os.Chdir("./data")
	
	
	gfx := new(grafic2d.GFXServer)
	width, height := gfx.Init() // OpenGL, etc initialization
	log.Printf("Screen dimension = %vx%v\n", width, height)
	//spritetest(gfx)
	showPinboard(gfx)
	
	defer gfx.Finish() // Graphics cleanup
}

func spritetest(gfx *grafic2d.GFXServer) {
	fmt.Printf("Start Spritetest\n")
	twitty := grafic2d.NewSprite()
	twitty.AddImg(gfx.CreateImage("twitter-bird-sprite1.png"), 0)
	twitty.AddImg(gfx.CreateImage("twitter-bird-sprite2.png"), 250)
	twitty.AddImg(gfx.CreateImage("twitter-bird-sprite3.png"), 500)
	twitty.AddImg(gfx.CreateImage("twitter-bird-sprite4.png"), 750)
	t := grafic2d.AnimLinear{}
	t.EnableLooping = true
	t.AddFrame(0, []float32{0.0, 0.0})
	t.AddFrame(4000, []float32{1.0, 1.0})
	t.AddFrame(8000, []float32{0.0, 0.0})
	t.ValueMult = []float32{float32(gfx.DisplayWidth - twitty.Width), float32(gfx.DisplayHeight - twitty.Height)}
	anim := grafic2d.Animation(&t)
	twitty.SetAnimTrans(anim)

	imagetest(1024, 768, twitty, gfx)
	defer twitty.Destroy()
}

// imgtest draws images at the corners and center
func imagetest(w int, h int, sprite *grafic2d.Sprite, gfx *grafic2d.GFXServer) {
	var t grafic2d.Timer
	t.Start()
	var c uint8
	var buffer bytes.Buffer

	// Start the key checker
	keyPressed := make(chan bool)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		keyPressed <- true
	}()

	// Start the crawler
	entries := make(chan web.MessageData)
	quit := make(chan bool)
	crawler := web.NewMailCrawler("pop-mail.outlook.com:995", "mw-memories@outlook.de", "fitcup2014")
	go crawler.Crawl(entries, quit, 15*time.Second)

	// init values for the render loop
	infotext := "Last crawl result: none"
	loop := true
	
	// the render loop
	for loop {

		select {
		case _ = <-keyPressed:
			// send quit message to the crawler
			log.Println("Send quit message to crawler")
			quit <- true			
			// wait for the channel to be closed by the crawler
			<-entries			
			log.Println("Crawler has closed the channel, now we quit")
			loop = false				
		case e := <-entries:
			log.Printf("New entry received: %v, %v, %v\n", e.SenderName, e.ShortText, e.ImageNames)
			infotext = fmt.Sprintf("Last crawl result:\n %v\n %v\n %v", e.SenderName, e.ShortText, e.ImageNames)
			e.Save(e.CreateFilename())
			e.Load(e.CreateFilename())
		default:
			// no entry in the channel
		}

		gfx.Start(w, h)
		c = uint8((t.TimeSinceStart() / 50) % 256)
		gfx.Background(c%200, c, c%100)

		gfx.FillRGB(44, 100, 232, 1) // Big blue marble
		gfx.Circle(400, 0, 400)      // The "world"
		gfx.FillColor("white")       // White text

		buffer.WriteString("fps=")
		buffer.WriteString(strconv.Itoa(t.CallsPerSec()))
		gfx.Text(20, grafic2d.VGfloat(h-50), buffer.String(), "serif", 20)
		buffer.Reset()
		
		gfx.FillColor("black")       // Black text		
		gfx.Text(20, grafic2d.VGfloat(h-100), infotext, "serif", 10)

		sprite.Update(t.TimeSinceLastCall())
		sprite.Draw()
		gfx.End()
	}
}

func showPinboard(gfx *grafic2d.GFXServer) {
	// create the pinboard
	pb := new(Pinboard)
	// load messages from disk
	pb.LoadMessages()
	
	
	var t grafic2d.Timer
	t.Start()

	// Start the key checker
	keyPressed := make(chan bool)
	go func() {
		reader := bufio.NewReader(os.Stdin)
		reader.ReadString('\n')
		keyPressed <- true
	}()

	// Start the crawler
	entries := make(chan web.MessageData)
	quit := make(chan bool)
	crawler := web.NewMailCrawler("pop-mail.outlook.com:995", "mw-memories@outlook.de", "fitcup2014")
	go crawler.Crawl(entries, quit, time.Minute)

	// begin the pinboard
	pb.Begin(gfx)


	// init values for the render loop
	loop := true
	// the render loop
	for loop {

		select {
		case _ = <-keyPressed:
			// send quit message to the crawler
			log.Println("Send quit message to crawler")
			quit <- true			
			// wait for the channel to be closed by the crawler
			<-entries			
			log.Println("Crawler has closed the channel, now we quit")
			loop = false				
		case e := <-entries:
			log.Printf("New entry received: %v, %v, %v\n", e.SenderName, e.ShortText, e.ImageNames)
			e.Save(e.CreateFilename())
			pb.AddMessageData(&e)			
		default:
			// no entry in the channel
		}


		pb.Update(t.TimeSinceLastCall())
		pb.Draw()
	}
	
	// end the pinboard
	pb.End()
	
}
