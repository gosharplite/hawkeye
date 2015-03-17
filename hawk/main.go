package main

import (
	"flag"
	//	"fmt"
	"errors"
	"log"
	"net/http"
	"net/url"
	"runtime"
	//	"strings"
	"strconv"
	"time"
)

type flags struct {
	url url.URL
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
			}

			if err != nil {
				log.Printf("result: %v", err)
			}

		}(f.url)

		time.Sleep(time.Second * time.Duration(10))
	}
}

func getFlags() (flags, error) {

	u := flag.String("url", "http://localhost:8080", "snake url")

	flag.Parse()

	ur, err := url.Parse(*u)
	if err != nil {
		return flags{}, err
	}

	return flags{*ur}, nil
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
