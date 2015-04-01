package main

import (
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/asaskevich/govalidator"
	"net/url"
	"path"
	"strings"
	"time"
)

var visited = make(map[string]bool)
var notVisited = make(map[string]bool)
var counter = 0

const NCPU = 5

func worker(id int, jobs chan string, queue chan string, quit chan bool) {
	for j := range jobs {
		Parse(j, jobs, queue)
		visited[j] = true

		if len(getNotVisitedUrls(1)) == 0 {
			counter = counter + 1
			if counter == NCPU {
				go func() { quit <- true }()
			}
		} else {
			counter = 0
		}
	}
}

func main() {

	started := time.Now().Unix()

	jobs := make(chan string)
	queue := make(chan string)
	quit := make(chan bool)

	//creating workers
	for w := 1; w <= NCPU; w++ {
		go worker(w, jobs, queue, quit)
	}

	//sending urls
	jobs <- "/"

	for {
		select {
		case q := <-queue:
			if q == "" && len(getNotVisitedUrls(0)) == 0 && counter == NCPU {
				fmt.Println("total - ", len(visited))
				fmt.Println("time - ", time.Now().Unix()-started)
				return
			}
			if !visited[q] {
				visited[q] = false
				jobs <- q
				visited[q] = true
			}
		case <-quit:
			fmt.Println("total - ", len(visited))
			fmt.Println("time - ", time.Now().Unix()-started)
			return
		}
	}
}

func Parse(notParsedPageUrl string, jobs chan string, queue chan string) {
	homeUrl, _ := url.Parse(getHomeUrl())

	doc, err := goquery.NewDocument(homeUrl.String() + notParsedPageUrl)
	if err == nil {
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			href, exists := s.Attr("href")
			allowedExtensions := map[string]int{".html": 1, ".htm": 2, "": 3, ".asp": 4}
			_, existsExension := allowedExtensions[path.Ext(href)]

			if exists && govalidator.IsURL(href) && existsExension {
				parsedUrl, err := url.Parse(href)
				checkErr(err, "parsing url")
				innerAbsUrl := getInnerAbsUrl(parsedUrl, homeUrl)

				go func() { queue <- innerAbsUrl }()
			}
		})
	}
	go func() { queue <- "" }()
}
func getHomeUrl() string {
	return "http://windowsten.ru"
}

func getNotVisitedUrls(length int) (output []string) {
	for url, isVisited := range visited {
		if !isVisited {
			if len(output) < length || length == 0 {
				output = append(output, url)
			}
		}
	}
	return output
}
func getInnerAbsUrl(parsedUrl *url.URL, homeUrl *url.URL) string {
	if parsedUrl.IsAbs() {
		if parsedUrl.Host == homeUrl.Host {
			return strings.Trim(parsedUrl.Path, " ")
		}
	} else if strings.Index(parsedUrl.String(), "#") == -1 {
		return strings.Trim(parsedUrl.String(), " ")
	}
	return ""
}

func checkErr(err error, msg string) {
	if err != nil {
		fmt.Println(msg, err)
	}
}
