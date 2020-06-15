package main

import (
	"astuart.co/goq"
	"fmt"
	"github.com/gocolly/colly"
	"io/ioutil"
	"log"
	"regexp"
	"strconv"
	"strings"
)

type Book struct {
	Title      string `goquery:"a.bookTitle span"`
	URL        string `goquery:"a,[href]"`
	Ratings    string `goquery:"span.minirating"`
	AvgRating  uint16 // 0 - 500
	NumRatings uint16
}

func parseRatings(book *Book) {
	r, err := regexp.Compile(`(\d\.\d{2}) avg rating — (\d+) ratings?`)
	if err != nil {
		log.Fatal(err)
	}
	groups := r.FindStringSubmatch(book.Ratings)

	avgRating, err := strconv.ParseUint(strings.Replace(groups[1], ".", "", 1), 10, 9)
	if err != nil {
		log.Fatal(err)
	}
	book.AvgRating = uint16(avgRating)

	numRatings, err := strconv.ParseUint(groups[2], 10, 16)
	if err != nil {
		log.Fatal(err)
	}
	book.NumRatings = uint16(numRatings)
}

func handleBookElement(bookElement *colly.HTMLElement) {
	if bookElement.Attr("itemtype") != "http://schema.org/Book" {
		return
	}

	var book *Book = &Book{}
	selector := goq.NodeSelector(bookElement.DOM.Nodes)
	err := goq.UnmarshalSelection(selector, book)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(book.Ratings)
	parseRatings(book)
	fmt.Printf("%d|%d|%s\n", book.NumRatings, book.AvgRating, book.Title)
}

func main() {
	log.SetOutput(ioutil.Discard)

	// Instantiate default collector
	c := colly.NewCollector(
		colly.AllowedDomains("www.goodreads.com"),
		colly.CacheDir("./cache"),
		colly.MaxDepth(2),
	)

	c.OnHTML("tr[itemtype]", handleBookElement)

	// On every a element which has href attribute call callback
	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		if e.Attr("class") != "next_page" {
			return
		}

		link := e.Attr("href")
		// Print link
		log.Printf("Link found: %q -> %s\n", e.Text, link)
		// Visit link found on page
		// Only those links are visited which are in AllowedDomains
		c.Visit(e.Request.AbsoluteURL(link))
	})

	// Before making a request print "Visiting ..."
	c.OnRequest(func(r *colly.Request) {
		log.Println("Visiting", r.URL.String())
	})

	c.Visit("https://www.goodreads.com/search?utf8=%E2%9C%93&q=baby+sign+language&search_type=books")
}
