package web

import (
	"bytes"
	"encoding/gob"
	"github.com/flothe/pinboard/grafic2d"
	"fmt"
	"github.com/ChimeraCoder/anaconda"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"
)


type TwitterCrawler struct {
	api  anaconda.TwitterApi
	consumerKey, consumerSecret, accessToken, accessTokenSecret string	
}


func NewTwitterCrawler(consumerKey, consumerSecret, accessToken, accessTokenSecret string) Crawler {
	crawler := &TwitterCrawler{
		consumerKey:  consumerKey,
		consumerSecret: consumerSecret,
		accessToken: accessToken,
		accessTokenSecret: accessTokenSecret,
	}
	
	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)

	return crawler
}

func (crawler *TwitterCrawler) Crawl(entries chan<- MessageData, quit <-chan bool, repeatDuration time.Duration) {

	log.Println("Start crawling tweets.")
	loop := true

	//wait in a separate goroutine for the quit message. if received stop the crawling loop
	go func() {
		<-quit
		loop = false
	}()

	// relogin every hour
	reloginTime := time.Now().Add(1 * time.Hour)

	// login to the server
	crawler.api := anaconda.NewTwitterApi(crawler.accessToken, crawler.accessTokenSecret)
	tweets, err := crawler.api.GetHomeTimeline()
	if err != nil {
		fmt.Print(err.Error())
		return
	}
	
	for _, tweet := range tweets {
	    fmt.Print(tweet.Text)
	}
	
/*	
	
	// now we start crawling until an error occures or the channel is closed
	for loop {
		t0 := time.Now()
		log.Printf("Start new crawling attemp\n")

		// get number of messages
		msgNrs, _, er := client.ListAll()
		if er != nil {
			fmt.Print(er.Error())
			close(entries)
			return
		}
		log.Printf("Number of messages: %v\n", msgNrs)

		client.Stat()

		// if it exists, get the first message ...
		if len(msgNrs) > 0 {
			msg, er := client.RetrMsg(msgNrs[0])
			if er == nil {
				log.Printf("Retrieve messages: %v\n", msgNrs[0])
				// create a CrawlEntry out of the message
				entry, er := processMailMessage(msg)
				if er == nil {
					log.Printf("Created entry from mail: %v, %v, %v\n", entry.SenderName, entry.ShortText, entry.ImageNames)
					// ... and delete the message on the server
					er = client.Dele(msgNrs[0])
					if er == nil {
						log.Printf("Delete message: %v\n", msgNrs[0])
						// send message using the channel
						entries <- *entry
					}
				}
			}
			if er != nil {
				log.Printf("Crawling attemp failed: %v\n", er)
				//try to recover
				reloginTime = time.Now()
				time.Sleep(time.Minute * 1)
			}

		}

		// wait some time
		t1 := time.Now()
		duration := repeatDuration - (t1.Sub(t0))
		for duration > 0 && loop {
			// wait a second
			time.Sleep(time.Second)
			// recalculate the duration of this crawling attemp
			t1 = time.Now()
			duration = repeatDuration - (t1.Sub(t0))
		}

		// relogin to the server
		if reloginTime.Before(time.Now()) {
			log.Println("Relogin to the server ...")
			client.Quit()
			time.Sleep(time.Second * 3)
			client, er = crawler.login()
			if er != nil {
				reloginTime = time.Now()
				log.Println("Going to sleep for 5 minutes. Afterwards trying again to login.")
				time.Sleep(time.Minute * 5)
			} else {
				// relogin every hour
				reloginTime = time.Now().Add(1 * time.Hour)
				log.Printf("Next relogin at: %v\n", reloginTime)
			}
		}

	}
	client.Quit()
	log.Printf("Stop crawling mails from: %s", crawler.url)
	close(entries)
	
	*/
}


func processTweet(tweet *anaconda.Tweet) (*MessageData, error) {


	/*
	data := new(MessageData)
	data.Type = EMAIL

	mime, err := enmime.ParseMIMEBody(msg) // Parse message body with enmime
	if err != nil {
		return nil, fmt.Errorf("Failed to parse MIME body: ", err.Error())
	}

	data.Timestamp, err = msg.Header.Date()
	if err != nil {
		return nil, fmt.Errorf("Failed to get date from message: ", err.Error())
	}
	data.ShortText = mime.GetHeader("Subject")
	data.SenderName = mime.GetHeader("From")
	data.LongText = mime.Text
	log.Printf("MIME Text:%v\n", mime.Text)

	// handle attachments
	log.Printf("MIME attachments:%v\n", len(mime.Attachments))
	for _, a := range mime.Attachments {
		err := data.saveMultimediaAttachment(a)
		if err != nil {
			return nil, fmt.Errorf("Failed to checkAndSaveAttachment: ", err.Error())
		}
	}

	// handle inlines
	log.Printf("MIME inlines:%v\n", len(mime.Inlines))
	for _, a := range mime.Inlines {
		err := data.saveMultimediaAttachment(a)
		if err != nil {
			return nil, fmt.Errorf("Failed to checkAndSaveAttachment: ", err.Error())
		}
	}
	*/

	/*
		log.Printf("mime.Root = %v\n", mime.Root)
		if mime.Root != nil {
			pics := enmime.BreadthMatchAll(mime.Root, func(pt enmime.MIMEPart) bool {
				log.Printf("BreadthMatchAll: %s\n‚", pt.ContentType())
				return pt.ContentType() == "image/jpeg" || pt.ContentType() == "image/png" || pt.ContentType() == "image/jpg"
			})
			log.Printf("MIME BreadthMatchAll found %v images.\n", len(pics))
			for _, a := range pics {
				err := data.saveMultimediaAttachment(a)
				if err != nil {
					return nil, fmt.Errorf("Failed to checkAndSaveAttachment: ", err.Error())
				}
			}
		}

	*/
	return nil, nil

}

func createUniqueFilename(filename, extension string) string {

	// remove file extemsion
	fn := filename
	n := strings.LastIndex(filename, ".")
	if n>0 {
		fn = filename[:n]
	}

	fnRes := fn+"."+extension
	_, err := os.Stat(fnRes)
	i := 0
	for (err == nil) && i < 10000 {
		fnRes = fmt.Sprintf("%v%04v.%v", fn, i, extension)
		_, err = os.Stat(fnRes)
		i++
	}
	return fnRes

}
