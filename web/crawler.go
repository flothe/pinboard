package web

import (
	"bytes"
	"encoding/gob"
	"github.com/flothe/pinboard/grafic2d"
	"fmt"
	"github.com/jhillyerd/go.enmime"
	"io/ioutil"
	"log"
	"net/mail"
	"os"
	"strings"
	"time"
)

type MessageDataType int

const (
	UNDEF MessageDataType = iota
	TWEET
	EMAIL
)

type Crawler interface {
	Crawl(chan<- MessageData, <-chan bool, time.Duration)
}

type MessageData struct {
	Type       MessageDataType
	Timestamp  time.Time
	SenderName string
	ShortText  string
	LongText   string
	ImageNames []string
	VideoNames []string
	AudioNames []string
}

type MailCrawler struct {
	client          Client
	url, user, pass string
}

func (e *MessageData) Save(filename string) error {

	m := new(bytes.Buffer)
	enc := gob.NewEncoder(m)

	err := enc.Encode(e)
	if err != nil {
		log.Printf("Failed to create gob encoder for CrawlEntry: %v, %v, %v, %v\n", e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
		log.Println(err)
		return err
	}

	err = ioutil.WriteFile(filename, m.Bytes(), 0600)
	if err != nil {
		log.Printf("Failed to write CrawlEntry to %s: %v, %v, %v, %v\n", filename, e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
		log.Println(err)
		return err
	}
	log.Printf("Saved CrawlEntry to %s: %v, %v, %v, %v\n", filename, e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
	return nil
}

func (e *MessageData) Load(filename string) error {
	n, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Failed to read file %s: %v, %v, %v, %v\n", filename, e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
		log.Println(err)
		return err
	}

	p := bytes.NewBuffer(n)
	dec := gob.NewDecoder(p)

	err = dec.Decode(e)
	if err != nil {
		log.Printf("Failed to create gob decoder for CrawlEntry %s: %v, %v, %v, %v\n", filename, e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
		log.Println(err)
		return err
	}
	log.Printf("Loaded CrawlEntry from %s: %v, %v, %v, %v\n", filename, e.ShortText, e.Timestamp, e.SenderName, e.ImageNames)
	return nil
}

func (e *MessageData) CreateFilename() string {
	t := fmt.Sprintf("%d%02d%02d-%02d%02d%02d",
		e.Timestamp.Year(),
		e.Timestamp.Month(),
		e.Timestamp.Day(),
		e.Timestamp.Hour(),
		e.Timestamp.Minute(),
		e.Timestamp.Second())
	fn := fmt.Sprintf("Pin-%s-%s", t, e.SenderName)
	r := strings.NewReplacer(" ", "-", "@", "-", "<", "", ">", "", ".", "-", "--", "-")
	fn = r.Replace(fn) + ".cmsg"

	return fn
}

// saves the attachment if it is a image, audio or video.
// Returns the filename, a type indicator ("image", "audio" or "video")
// if the attachment is not saved nil is returned as  filename and type indicator
func (data *MessageData) saveMultimediaAttachment(part enmime.MIMEPart) error {

	t := part.ContentType()

	log.Printf("Start saving of a %v attachment.", t)

	if t == "image/jpeg" || t == "image/png" {

		var buf []byte
		buf = part.Content()
		
		// remove file extemsion
		fn := part.FileName()
		n := strings.LastIndex(fn, ".")
		if n>0 {
			fn = fn[:n]
		}

		fn = createUniqueFilename(fn, "rgba")
		err := grafic2d.SaveScaledVGImage(buf, 1920, 1080, fn) 
		if err != nil {
			log.Printf("Failed to save attachment %v as %v: %v", t, fn, err)
			return err
		}
		
		data.ImageNames = append(data.ImageNames, fn)

		return nil
	}

	log.Printf("Failed to save content type: %v", t)
	return nil
}

func NewMailCrawler(url, username, password string) Crawler {
	crawler := &MailCrawler{
		url:  url,
		user: username,
		pass: password,
	}

	return crawler
}

func (crawler *MailCrawler) Crawl(entries chan<- MessageData, quit <-chan bool, repeatDuration time.Duration) {

	log.Printf("Start crawling mails from: %s", crawler.url)
	loop := true

	//wait in a separate goroutine for the quit message. if received stop the crawling loop
	go func() {
		<-quit
		loop = false
	}()

	// relogin every hour
	reloginTime := time.Now().Add(1 * time.Hour)

	// login to the server
	client, er := crawler.login()
	if er != nil {
		reloginTime = time.Now()
		log.Println("Going to sleep for 5 minutes. Afterwards trying again.")
		time.Sleep(time.Minute * 5)
	}

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
}

func (crawler *MailCrawler) login() (*Client, error) {
	// create client
	client, er := DialTLS(crawler.url)
	if er == nil {
		// login to server
		er = client.Auth(crawler.user, crawler.pass)
	}

	if er != nil {
		log.Printf("Login to server (%v) failed: %v\n", crawler.url, er)
	} else {
		log.Printf("Login to server (%v) successful.\n", crawler.url)
	}

	return client, er
}

func processMailMessage(msg *mail.Message) (*MessageData, error) {

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

	return data, nil
}

func createUniqueFilename(filename, extension string) string {

	// remove file extemsion
	fn := filename
	n := strings.LastIndex(filename, ".")
	if n>0 {
		fn = filename[:n]
	}
	// remove all whitespaces
	fn = strings.Replace(fn, " ", "", -1)

	fnRes := fn+"."+extension
	_, err := os.Stat(fnRes)
	i := 0
	for (err == nil) && i < 10000 {
		fnRes = fmt.Sprintf("%v-%04v.%v", fn, i, extension)
		_, err = os.Stat(fnRes)
		i++
	}
	return fnRes

}
