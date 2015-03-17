package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"
)

type flags struct {
	url url.URL
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU()*2 + 1)

	log.Printf("Snake is moving...")

	f, err := getFlags()
	if err != nil {
		log.Fatalf("flags parsing fail: %v", err)
	}

	http.HandleFunc("/", handler)

	err = http.ListenAndServe(getPort(f.url), nil)
	if err != nil {
		log.Fatalf("ListenAndServe: ", err)
	}
}

func getFlags() (flags, error) {

	u := flag.String("url", "http://localhost:8080", "snake url")

	flag.Parse()

	ur, err := url.Parse(*u)
	if err != nil {
		log.Printf("url parse err: %v", err)
		return flags{}, err
	}

	return flags{*ur}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {

	log.Printf("Someone is looking!")

	fmt.Fprint(w, r.Host+","+strconv.FormatInt(time.Now().UnixNano(), 10))
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
