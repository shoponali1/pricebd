package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type Price struct {
	Date        string `json:"date"`
	Time        string `json:"time"`
	K22         int    `json:"k22"`
	K21         int    `json:"k21"`
	K18         int    `json:"k18"`
	Traditional int    `json:"traditional"`
}

func writeRow(row *[]string, priceData *Price) {
	(*row)[0] = priceData.Date
	(*row)[1] = priceData.Time
	(*row)[2] = strconv.Itoa(priceData.K18)
	(*row)[3] = strconv.Itoa(priceData.K21)
	(*row)[4] = strconv.Itoa(priceData.K22)
	(*row)[5] = strconv.Itoa(priceData.Traditional)
}

func main() {
	// Create collector with browser-like settings
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36"),
	)

	// Add headers to mimic a real browser
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		r.Headers.Set("Accept-Language", "en-US,en;q=0.9")
		r.Headers.Set("Accept-Encoding", "gzip, deflate, br")
		r.Headers.Set("Connection", "keep-alive")
		r.Headers.Set("Upgrade-Insecure-Requests", "1")
		r.Headers.Set("Sec-Fetch-Dest", "document")
		r.Headers.Set("Sec-Fetch-Mode", "navigate")
		r.Headers.Set("Sec-Fetch-Site", "none")
		r.Headers.Set("Cache-Control", "max-age=0")
		fmt.Println("🌐 Visiting:", r.URL)
	})

	// Add delay to avoid rate limiting
	c.Limit(&colly.LimitRule{
		DomainGlob:  "*bajus.org*",
		Delay:       2 * time.Second,
		RandomDelay: 1 * time.Second,
	})

	now := time.Now()
	todayPrice := Price{}
	todayPrice.Date = now.Format("2006-01-02")
	todayPrice.Time = now.Format("15:04:05")

	todaySilverPrice := Price{}
	todaySilverPrice.Date = now.Format("2006-01-02")
	todaySilverPrice.Time = now.Format("15:04:05")

	getPrice := func(e *colly.HTMLElement) int {
		priceStr := strings.NewReplacer(",", "", " BDT/GRAM", "").Replace(e.Text)
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			fmt.Println("⚠️ Error parsing price:", err)
			return 0
		}
		return price
	}

	// Scrape Gold Prices
	c.OnHTML(".gold-table tr:nth-child(1) .price", func(e *colly.HTMLElement) {
		todayPrice.K22 = getPrice(e)
		fmt.Println("✅ Gold K22:", todayPrice.K22)
	})
	c.OnHTML(".gold-table tr:nth-child(2) .price", func(e *colly.HTMLElement) {
		todayPrice.K21 = getPrice(e)
		fmt.Println("✅ Gold K21:", todayPrice.K21)
	})
	c.OnHTML(".gold-table tr:nth-child(3) .price", func(e *colly.HTMLElement) {
		todayPrice.K18 = getPrice(e)
		fmt.Println("✅ Gold K18:", todayPrice.K18)
	})
	c.OnHTML(".gold-table tr:nth-child(4) .price", func(e *colly.HTMLElement) {
		todayPrice.Traditional = getPrice(e)
		fmt.Println("✅ Gold Traditional:", todayPrice.Traditional)
	})

	// Scrape Silver Prices
	c.OnHTML(".silver-table tr:nth-child(1) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K22 = getPrice(e)
		fmt.Println("✅ Silver K22:", todaySilverPrice.K22)
	})
	c.OnHTML(".silver-table tr:nth-child(2) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K21 = getPrice(e)
		fmt.Println("✅ Silver K21:", todaySilverPrice.K21)
	})
	c.OnHTML(".silver-table tr:nth-child(3) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K18 = getPrice(e)
		fmt.Println("✅ Silver K18:", todaySilverPrice.K18)
	})
	c.OnHTML(".silver-table tr:nth-child(4) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.Traditional = getPrice(e)
		fmt.Println("✅ Silver Traditional:", todaySilverPrice.Traditional)
	})

	// Error handler
	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("❌ Request failed with response: %d %s\n", r.StatusCode, err)
		fmt.Println("Response headers:")
		fmt.Println(r.Headers)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("\n=== Scraping Completed ===")
		fmt.Printf("Gold: Date=%s Time=%s K22=%d K21=%d K18=%d Traditional=%d\n",
			todayPrice.Date, todayPrice.Time, todayPrice.K22, todayPrice.K21, todayPrice.K18, todayPrice.Traditional)
		fmt.Printf("Silver: Date=%s Time=%s K22=%d K21=%d K18=%d Traditional=%d\n",
			todaySilverPrice.Date, todaySilverPrice.Time, todaySilverPrice.K22, todaySilverPrice.K21, todaySilverPrice.K18, todaySilverPrice.Traditional)

		if todayPrice.K22 == 0 {
			fmt.Println("⚠️ WARNING: Gold prices were not scraped! Check selectors.")
		}
		if todaySilverPrice.K22 == 0 {
			fmt.Println("⚠️ WARNING: Silver prices were not scraped! Check selectors.")
		}

		savePrice("./fe/src/prices.csv", &todayPrice)
		savePrice("./fe/src/silver-prices.csv", &todaySilverPrice)

		savePriceJSON("./fe/src/prices.json", &todayPrice)
		savePriceJSON("./fe/src/silver-prices.json", &todaySilverPrice)

		fmt.Println("=== Files Updated Successfully ===")
	})

	// Visit the website
	fmt.Println("🚀 Starting scraper...")
	err := c.Visit("https://www.bajus.org/gold-price")
	if err != nil {
		fmt.Println("❌ Error visiting website:", err)
		fmt.Println("💡 Tip: The website might be blocking automated requests")
		os.Exit(1)
	}
}

func savePrice(filename string, priceData *Price) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil || len(records) == 0 {
		records = [][]string{{"Date", "Time", "K18", "K21", "K22", "Traditional"}}
	}

	fmt.Printf("📝 Adding new record for %s %s to %s\n", priceData.Date, priceData.Time, filename)
	newRecord := make([]string, 6)
	writeRow(&newRecord, priceData)
	records = append(records, newRecord)

	f.Seek(0, 0)
	f.Truncate(0)
	writer := csv.NewWriter(f)
	writer.WriteAll(records)
	writer.Flush()

	if err := writer.Error(); err != nil {
		panic(err)
	}
}

func savePriceJSON(filename string, priceData *Price) {
	var prices []Price
	file, err := ioutil.ReadFile(filename)
	if err == nil {
		json.Unmarshal(file, &prices)
	}

	fmt.Printf("📝 Adding new JSON entry for %s %s to %s\n", priceData.Date, priceData.Time, filename)
	prices = append(prices, *priceData)

	data, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		panic(err)
	}
}
