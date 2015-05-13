package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type flags struct {
	url     url.URL
	content []byte
}

var Dat *flags

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()*2 + 1)

	log.Printf("Snake is moving...")

	f, err := getFlags()
	if err != nil {
		log.Fatalf("flags parsing fail: %v", err)
	}

	Dat = &f

	http.HandleFunc("/", handler)

	err = http.ListenAndServe(getPort(f.url), nil)
	if err != nil {
		log.Fatalf("ListenAndServe: ", err)
	}
}

func getFlags() (flags, error) {

	u := flag.String("url", "http://localhost:8080", "snake url")

	fn := flag.String("file", "content.html", "file name")

	flag.Parse()

	ur, err := url.Parse(*u)
	if err != nil {
		log.Printf("url parse err: %v", err)
		return flags{}, err
	}

	body, err := ioutil.ReadFile(*fn)
	if err != nil {
		log.Printf("ioutil.ReadFile err: %v", err)
		body = nil
	}

	return flags{*ur, body}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Someone is looking!")

	if Dat.content == nil {
		fmt.Fprint(w, r.Host+","+strconv.FormatInt(time.Now().UnixNano(), 10))
	} else {
		fmt.Fprint(w, string(Dat.content))
	}
}

func getPort(u url.URL) string {

	r := u.Host

	if n := strings.Index(r, ":"); n != -1 {
		r = r[n:]
	} else {
		r = ":8080"
	}

	return r
}
