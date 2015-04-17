package main

import (
	"bytes"
	"errors"
	"flag"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"runtime"
	"strconv"
	"text/template"
	"time"
)

type flags struct {
	url        url.URL
	caption    string
	from_Gmail string
	to_mail    string
	password   string
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()*2 + 1)

	log.Printf("Hawk is flying...")

	f, err := getFlags()
	if err != nil {
		log.Fatalf("flags parsing fail: %v", err)
	}

	for {

		go func(url url.URL) {

			c := make(chan error, 1)
			go looking(url, c)

			t := time.Now()

			err := error(nil)

			select {

			case <-time.After(1 * time.Second):
				err = errors.New("timeout")

			case err = <-c:
				log.Printf("looking: " + time.Since(t).String())
				close(c)
			}

			if err != nil {

				go sendGMail(f, err)
				log.Printf("result: %v", err)

				err, ok := <-c
				if ok {

					if err == nil {
						err = errors.New("timeout without error")
					}

					go sendGMail(f, err)
					log.Printf("result: %v", err)
				}
			}

		}(f.url)

		time.Sleep(time.Second * time.Duration(10))
	}
}

func getFlags() (flags, error) {

	// parse
	u := flag.String("u", "http://localhost:8080", "hawk url")
	c := flag.String("c", "cobra", "caption")
	f := flag.String("f", "sender@gmail.com", "gmail sender")
	t := flag.String("t", "receiver@example.com", "email receiver")
	p := flag.String("p", "123456", "gmail password")

	flag.Parse()

	// url
	ur, err := url.Parse(*u)
	if err != nil {
		return flags{}, err
	}

	// caption
	ca := *c

	// from_Gmail
	fr := *f

	// to_mail
	to := *t

	//password
	pw := *p

	return flags{*ur, ca, fr, to, pw}, nil
}

func looking(url url.URL, c chan error) {

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		c <- err
		return
	}

	req.Header.Set("Connection", "close")
	req.Header.Set("User-Agent", "HawkEye")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		c <- err
		return
	}

	if err := resp.Body.Close(); err != nil {
		c <- err
		return
	}

	if resp.StatusCode != 200 {
		c <- errors.New("http resp: " + strconv.Itoa(resp.StatusCode) + " " + resp.Status)
		return
	}

	c <- nil
}

func sendGMail(f flags, e error) {

	auth := smtp.PlainAuth(
		"",
		f.from_Gmail,
		f.password,
		"smtp.gmail.com",
	)

	type SmtpTemplateData struct {
		From    string
		To      string
		Subject string
		Body    string
	}

	const emailTemplate = `From: {{.From}}
To: {{.To}}
Subject: {{.Subject}}

{{.Body}}
`

	var err error
	var doc bytes.Buffer

	context := &SmtpTemplateData{
		f.from_Gmail,
		f.to_mail,
		f.caption + " " + time.Now().Format("01/02 15:04:05"),
		e.Error(),
	}

	t := template.New("emailTemplate")
	t, err = t.Parse(emailTemplate)
	if err != nil {
		log.Printf("error trying to parse mail template")
		return
	}
	err = t.Execute(&doc, context)
	if err != nil {
		log.Printf("error trying to execute mail template")
		return
	}

	err = smtp.SendMail(
		"smtp.gmail.com:587",
		auth,
		f.from_Gmail,
		[]string{f.to_mail},
		doc.Bytes(),
	)
	if err != nil {
		log.Printf("smtp.SendMail err: " + err.Error())
		return
	}
}
