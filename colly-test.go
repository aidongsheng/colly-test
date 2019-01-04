package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/debug"
	"github.com/gocolly/colly/proxy"
	queue2 "github.com/gocolly/colly/queue"
	"log"
	"net/http"
	"time"
)

func parallel () {
	// Instantiate default collector
	c := colly.NewCollector(
		// MaxDepth is 2, so only the links on the scraped page
		// and links on those pages are visited
		colly.MaxDepth(2),
		colly.Async(true),
	)

	// Limit the maximum parallelism to 2
	// This is necessary if the goroutines are dynamically
	// created to control the limit of simultaneous requests.
	//
	// Parallelism can be controlled also by spawning fixed
	// number of go routines.
	c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 2})

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Attr("href")
		// Print link
		fmt.Println(link)
		// Visit link found on page on a new thread
		e.Request.Visit(link)
	})

	// Start scraping on https://en.wikipedia.org
	c.Visit("https://en.wikipedia.org/")
	// Wait until threads are finished
	c.Wait()
}


func proxySwitcher() {
	// Instantiate default collector
	c := colly.NewCollector(colly.AllowURLRevisit())

	// Rotate two socks5 proxies
	rp, err := proxy.RoundRobinProxySwitcher("socks5://127.0.0.1:1337", "socks5://127.0.0.1:1338")
	if err != nil {
		log.Fatal(err)
	}
	c.SetProxyFunc(rp)

	// Print the response
	c.OnResponse(func(r *colly.Response) {
		log.Printf("%s\n", bytes.Replace(r.Body, []byte("\n"), nil, -1))
	})

	c.OnHTML("div[id]", func(element *colly.HTMLElement) {
		if element.Attr("id") == "json" {
			fmt.Printf("%s",element.Text)
		}
	})
	// Fetch httpbin.org/ip five times
	for i := 0; i < 5; i++ {
		c.Visit("https://httpbin.org/ip")
	}
}

func queue() {
	url := "https://httpbin.org/delay/1"

	// Instantiate default collector
	c := colly.NewCollector()

	// create a request queue with 2 consumer threads
	q, _ := queue2.New(
		2, // Number of consumer threads
		&queue2.InMemoryQueueStorage{MaxSize:10000},
	)

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("visiting", r.URL)
	})

	for i := 0; i < 5; i++ {
		// Add URLs to the queue
		q.AddURL(fmt.Sprintf("%s?n=%d", url, i))
	}
	// Consume URLs
	q.Run(c)
}

func randomDelay()  {
	url := "https://httpbin.org/delay/2"

	// Instantiate default collector
	c := colly.NewCollector(
		// Attach a debugger to the collector
		colly.Debugger(&debug.LogDebugger{}),
		colly.Async(true),
	)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*httpbin.*",
		Parallelism: 2,
		RandomDelay: 5 * time.Second,
	})

	// Start scraping in four threads on https://httpbin.org/delay/2
	for i := 0; i < 4; i++ {
		c.Visit(fmt.Sprintf("%s?n=%d", url, i))
	}
	// Start scraping on https://httpbin.org/delay/2
	c.Visit(url)
	// Wait until threads are finished
	c.Wait()
}

func rateLimit() {
	url := "https://httpbin.org/delay/2"

	// Instantiate default collector
	c := colly.NewCollector(
		// Turn on asynchronous requests
		colly.Async(true),
		// Attach a debugger to the collector
		colly.Debugger(&debug.LogDebugger{}),
	)

	// Limit the number of threads started by colly to two
	// when visiting links which domains' matches "*httpbin.*" glob
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*httpbin.*",
		Parallelism: 2,
		//Delay:      5 * time.Second,
	})

	// Start scraping in five threads on https://httpbin.org/delay/2
	for i := 0; i < 5; i++ {
		c.Visit(fmt.Sprintf("%s?n=%d", url, i))
	}
	// Wait until threads are finished
	c.Wait()
}

func redisbackend() {
	//urls := []string{
	//	"http://httpbin.org/",
	//	"http://httpbin.org/ip",
	//	"http://httpbin.org/cookies/set?a=b&c=d",
	//	"http://httpbin.org/cookies",
	//}
	//
	//c := colly.NewCollector()
	//
	//// create the redis storage
	//storage := &redisstorage.Storage{
	//	Address:  "127.0.0.1:6379",
	//	Password: "",
	//	DB:       0,
	//	Prefix:   "httpbin_test",
	//}
	//
	//// add storage to the collector
	//err := c.SetStorage(storage)
	//if err != nil {
	//	panic(err)
	//}
	//
	//// delete previous data from storage
	//if err := storage.Clear(); err != nil {
	//	log.Fatal(err)
	//}
	//
	//// close redis client
	//defer storage.Client.Close()
	//
	//// create a new request queue with redis storage backend
	//q, _ := queue.New(2, storage)
	//
	//c.OnResponse(func(r *colly.Response) {
	//	log.Println("Cookies:", c.Cookies(r.Request.URL.String()))
	//})
	//
	//// add URLs to the queue
	//for _, u := range urls {
	//	q.AddURL(u)
	//}
	//// consume requests
	//q.Run(c)
}

func requestContext() {
	// Instantiate default collector
	c := colly.NewCollector()

	// Before making a request put the URL with
	// the key of "url" into the context of the request
	c.OnRequest(func(r *colly.Request) {
		r.Ctx.Put("url", r.URL.String())
	})

	// After making a request get "url" from
	// the context of the request
	c.OnResponse(func(r *colly.Response) {
		fmt.Println(r.Ctx.Get("url"))
	})

	// Start scraping on https://en.wikipedia.org
	c.Visit("https://en.wikipedia.org/")
}


type pageInfo struct {
	StatusCode int
	Links      map[string]int
}

func handler(w http.ResponseWriter, r *http.Request) {
	URL := r.URL.Query().Get("url")
	if URL == "" {
		log.Println("missing URL argument")
		return
	}
	log.Println("visiting", URL)

	c := colly.NewCollector()

	p := &pageInfo{Links: make(map[string]int)}

	// count links
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		link := e.Request.AbsoluteURL(e.Attr("href"))
		if link != "" {
			p.Links[link]++
		}
	})

	// extract status code
	c.OnResponse(func(r *colly.Response) {
		log.Println("response received", r.StatusCode)
		p.StatusCode = r.StatusCode
	})
	c.OnError(func(r *colly.Response, err error) {
		log.Println("error:", r.StatusCode, err)
		p.StatusCode = r.StatusCode
	})

	c.Visit(URL)

	// dump results
	b, err := json.Marshal(p)
	if err != nil {
		log.Println("failed to serialize response:", err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	w.Write(b)
}

func main() {
	// example usage: curl -s 'http://127.0.0.1:7171/?url=http://go-colly.org/'
	addr := ":7171"

	http.HandleFunc("/", handler)

	log.Println("listening on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}