package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
	"gopkg.in/yaml.v3"
)

var kUserAgent = []string{
	"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:47.0) Gecko/20100101 Firefox/47.0",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36",
	"Mozilla/5.0 (Macintosh; Intel Mac OS X x.y; rv:42.0) Gecko/20100101 Firefox/42.0",
	"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36 Edg/91.0.864.59",
	"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.106 Safari/537.36 OPR/38.0.2220.41",
	"Mozilla/5.0 (iPhone; CPU iPhone OS 13_5_1 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.1.1 Mobile/15E148 Safari/604.1",
}

type ConfigData struct {
	SearchTerms  string `yaml:"search_terms"`
	GeolocTerms  string `yaml:"geoloc_terms"`
	MaxPages     int    `yaml:"max_pages"`
	RequestTimes int    `yaml:"request_times"`
}

func main() {
	configData := &ConfigData{}

	yamlFile, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal("Failed to read config file", err)
	}

	err = yaml.Unmarshal(yamlFile, configData)
	if err != nil {
		log.Fatal("Failed to map config data:", err)
	}

	fmt.Printf("SearchTerms: %s\n", configData.SearchTerms)
	fmt.Printf("GeolocTerms: %s\n", configData.GeolocTerms)

	kSearchTerms := url.QueryEscape(configData.SearchTerms)
	kGeolocTerms := url.QueryEscape(configData.GeolocTerms)

	file, err := os.Create("result.csv")
	if err != nil {
		log.Fatal("Failed to create csv file:", err)
	}
	defer file.Close()

	csvFile := csv.NewWriter(file)
	csvFile.Comma = ';'
	csvFile.Write([]string{
		"Business Name",
		"Website",
		"Telephone",
		"Address",
	})
	defer csvFile.Flush()

	for i := 0; i < configData.MaxPages; i++ {
		userAgentID := rand.Intn(len(kUserAgent))

		c := colly.NewCollector(colly.UserAgent(kUserAgent[userAgentID]))
		c.SetRequestTimeout(time.Minute)

		c.OnHTML(".info", func(h *colly.HTMLElement) {
			businessName := h.ChildText("div.info-section.info-primary > h2 > a > span")
			website := h.ChildAttr("div.info-section.info-primary > div.links > a.track-visit-website", "href")
			telephone := h.ChildText("div.info-section.info-secondary > div.phones.phone.primary")
			address := h.ChildText("div.info-section.info-secondary > div.adr")

			if len(website) > 50 {
				website = ""
			}

			if businessName != "" {
				csvFile.Write([]string{
					businessName,
					website,
					telephone,
					address,
				})
			}
		})

		fmt.Printf("\rScraping page number: %.2d", i+1)

		URL := fmt.Sprintf("https://www.yellowpages.com/search?search_terms=%s&geo_location_terms=%s&page=%d", kSearchTerms, kGeolocTerms, i+1)
		c.Visit(URL)

		time.Sleep(time.Duration(configData.RequestTimes) * time.Second)
	}

	fmt.Println("\n🎉 DONE!")
}
