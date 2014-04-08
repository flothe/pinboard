package pinboard

import (
	"github.com/flothe/pinboard/grafic2d"
	"github.com/flothe/pinboard/web"
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
	var url, user, pw string 
	if len(os.Args) > 3 {
		// os.Args[0] holds the programm name
		url = os.Args[1]
		user = os.Args[2]
		pw = os.Args[3]
	}
	
	showPinboard(gfx, url, user, pw)
	
	defer gfx.Finish() // Graphics cleanup
}


func showPinboard(gfx *grafic2d.GFXServer, url, user, pw string) {
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
	crawler := web.NewMailCrawler(url, user, pw)
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
