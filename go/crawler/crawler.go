package main

import (
	"fmt"
	"net/http"
	"time"
	"github.com/PuerkitoBio/goquery"
	"regexp"
)

func get_a_article(url string)string {
	resp, err := http.Get(url)
	if err != nil{
		panic(err)
	}
	defer resp.Body.Close()

	/*
	byteArray, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	*/

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil{
		panic(err)
	}

	retrunText := ""

	doc.Find("p").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		rawText := s.Text()
		if !regexp.MustCompile("[0-9]+年").MatchString(rawText) {
			re := regexp.MustCompile("[[0-9]+]")
			context := re.ReplaceAllString(rawText, "")
			re = regexp.MustCompile("\n")
			context = re.ReplaceAllString(context, "")
			retrunText += context
		}
	})
	return retrunText
}

func scrape() {
	urlList := [2]string{
		"https://ja.wikipedia.org/wiki/%E6%9C%88%E3%83%8E%E7%BE%8E%E5%85%8E", //月ノ美兎
		"https://ja.wikipedia.org/wiki/%E5%90%8D%E5%8F%96%E3%81%95%E3%81%AA", //名取さな
	}

	for i := range urlList {
		fmt.Printf("%s : %s\n", urlList[i], get_a_article(urlList[i]))
	}
}

func main() {
	fmt.Printf("get ready!")
	for {
		scrape()
		fmt.Printf("the end!")
		time.Sleep(1*time.Hour)
	}
}