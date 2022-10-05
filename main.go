package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"mvdan.cc/xurls/v2"
)

type Summary struct {
	Domain      string         `json:"domain"`
	TotalResult int            `json:"total_results"`
	Urls        map[string]int `json:"urls"`
}

var visitedUrls = make(map[string]int)
var domainName string

func main() {
	domain := flag.String("domain", "example.com", "Domain name to scan")
	flag.Parse()
	domainName = *domain
	url := "https://" + domainName
	spiderWebsite(url)
	fmt.Println("Total urls fetched", len(visitedUrls))
	writeJSON()
	fmt.Println("Finished scraping")
}

func spiderWebsite(inputUrl string) {
	fmt.Println("Scrapping url", inputUrl)
	statusCode, urls := getUrls(inputUrl)

	visitedUrls[inputUrl] = statusCode
	validUrls := getValidUrls(urls, domainName)

	for _, url := range validUrls {
		if _, exist := visitedUrls[url]; !exist {
			spiderWebsite(url)
		}
	}
}

func getUrls(url string) (int, []string) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/106.0.0.0 Safari/537.36")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Got an error while http get")
		return 0, []string{}
	}
	body, _ := ioutil.ReadAll(resp.Body)
	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(string(body), -1)
	return resp.StatusCode, urls
}

func getValidUrls(urls []string, domain string) []string {
	validUrls := []string{}
	for _, url := range urls {
		if strings.Contains(url, domain+"/") {
			joinUrls := strings.Split(url, "url=")
			if len(joinUrls) > 1 {
				url = joinUrls[len(joinUrls)-1]
			}
			if !strings.Contains(url, domain+"/") {
				continue
			}
			url = strings.TrimSuffix(url, "/")
			if strings.HasSuffix(url, ".png") || strings.HasSuffix(url, ".jpg") || strings.HasSuffix(url, ".css") {
				chunks := strings.Split(url, "/")
				newUrl := strings.Join(chunks[:len(chunks)-1], "/")
				validUrls = append(validUrls, newUrl)
			} else {
				validUrls = append(validUrls, url)
			}
		}
	}
	return validUrls
}

func writeJSON() {
	results := Summary{}
	results.Domain = domainName
	results.TotalResult = len(visitedUrls)
	results.Urls = visitedUrls

	jsonData, _ := json.MarshalIndent(results, "", "	")
	fileName := fmt.Sprintf("outputs/%s.json", domainName)
	ioutil.WriteFile(fileName, jsonData, 0644)
}
