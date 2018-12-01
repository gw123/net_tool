package main

import (
	"fmt"
	"log"
	"net/http"
	"github.com/PuerkitoBio/goquery"
	"time"
	"io/ioutil"
)

func getFileName() (filename, filepath string) {
	now := time.Now().UnixNano()
	filename = fmt.Sprintf("%d.png", now)
	filepath = "/home/gw/data/tmp/"
	return
}

func downloadImg(url string) (filename, filepath string) {
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	filename, filepath = getFileName()
	data, err := ioutil.ReadAll(res.Body)
	ioutil.WriteFile(filepath+filename, data, 0660)
	return
}

func ExampleScrape() {
	// Request the HTML page.
	res, err := http.Get("https://blog.csdn.net/RA681t58CJxsgCkJ31/article/details/80504482")
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	// Find the review items
	doc.Find("#article_content").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the band and title
		//fmt.Println(s.Html())
		//return
		s.Find("img").Each(func(index int, selection *goquery.Selection) {
			src, isExist := selection.Attr("src")
			if !isExist {
				return
			}
			filename, path := downloadImg(src)
			fmt.Println(index, src, filename, path)
		})
		//title := s.Find("i").Text()
		//fmt.Printf("Review %d: %s - %s\n", i, band, title)
	})
}

func main() {
	ExampleScrape()
}
