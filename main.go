package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"mvdan.cc/xurls/v2"
)


type Summary struct {
	Domain string `json:"domain"`
	TotalResult int `json:"total_results"`
	Urls []string `json:"urls"`
}

var visitedUrls = make(map[string]bool)
var domainName string

func main() {
	domain := flag.String("domain", "example.com", "Domain name to scan")
	flag.Parse()
	domainName = *domain
	url := "https://" + domainName
	spiderWebsite(url)
	fmt.Println("Total urls fetched",len(visitedUrls))
	writeJSON()
	fmt.Println("Finished scraping")
}


func spiderWebsite(inputUrl string){
	fmt.Println("Scrapping url", inputUrl)
	visitedUrls[inputUrl] = true
	urls := getUrls(inputUrl)
	validUrls := getValidUrls(urls,domainName)

	for _,url := range validUrls {
		if _,exist := visitedUrls[url]; !exist {
			spiderWebsite(url)
		}
	}
}

func getUrls(url string) []string{
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Got an error while http get")
		return []string{}
	}
	body,_ := ioutil.ReadAll(resp.Body)
	rxStrict := xurls.Strict()
	urls := rxStrict.FindAllString(string(body), -1)
	return urls
}

func getValidUrls(urls []string,domain string) []string{
	validUrls := []string{}
	for _,url := range urls{
		if strings.Contains(url,domain+"/"){
			joinUrls := strings.Split(url, "url=")
			if len(joinUrls) > 1 {
				url = joinUrls[len(joinUrls)-1]
			}
			if !strings.Contains(url,domain+"/"){
				continue
			}
			url = strings.TrimSuffix(url,"/")
			if strings.HasSuffix(url,".png") || strings.HasSuffix(url,".jpg") || strings.HasSuffix(url,".css"){
				chunks := strings.Split(url, "/")
				newUrl := strings.Join(chunks[:len(chunks)-1],"/")
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

	for url := range visitedUrls{
		results.Urls = append(results.Urls, url)
	}

	jsonData,_ := json.MarshalIndent(results,"", "	")
	fileName := fmt.Sprintf("outputs/%s.json",domainName)
	ioutil.WriteFile(fileName, jsonData, 0644)
}