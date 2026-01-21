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
	c := colly.NewCollector()

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
			fmt.Println("Error parsing price:", err)
			return 0
		}
		return price
	}

	// Scrape Gold Prices
	c.OnHTML(".gold-table tr:nth-child(1) .price", func(e *colly.HTMLElement) {
		todayPrice.K22 = getPrice(e)
	})
	c.OnHTML(".gold-table tr:nth-child(2) .price", func(e *colly.HTMLElement) {
		todayPrice.K21 = getPrice(e)
	})
	c.OnHTML(".gold-table tr:nth-child(3) .price", func(e *colly.HTMLElement) {
		todayPrice.K18 = getPrice(e)
	})
	c.OnHTML(".gold-table tr:nth-child(4) .price", func(e *colly.HTMLElement) {
		todayPrice.Traditional = getPrice(e)
	})

	// Scrape Silver Prices
	c.OnHTML(".silver-table tr:nth-child(1) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K22 = getPrice(e)
	})
	c.OnHTML(".silver-table tr:nth-child(2) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K21 = getPrice(e)
	})
	c.OnHTML(".silver-table tr:nth-child(3) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.K18 = getPrice(e)
	})
	c.OnHTML(".silver-table tr:nth-child(4) .price", func(e *colly.HTMLElement) {
		todaySilverPrice.Traditional = getPrice(e)
	})

	c.OnScraped(func(r *colly.Response) {
		fmt.Println("=== Scraping Completed ===")
		fmt.Printf("Gold: Date=%s Time=%s K22=%d K21=%d K18=%d Traditional=%d\n",
			todayPrice.Date, todayPrice.Time, todayPrice.K22, todayPrice.K21, todayPrice.K18, todayPrice.Traditional)
		fmt.Printf("Silver: Date=%s Time=%s K22=%d K21=%d K18=%d Traditional=%d\n",
			todaySilverPrice.Date, todaySilverPrice.Time, todaySilverPrice.K22, todaySilverPrice.K21, todaySilverPrice.K18, todaySilverPrice.Traditional)

		savePrice("./fe/src/prices.csv", &todayPrice)
		savePrice("./fe/src/silver-prices.csv", &todaySilverPrice)

		savePriceJSON("./fe/src/prices.json", &todayPrice)
		savePriceJSON("./fe/src/silver-prices.json", &todaySilverPrice)

		fmt.Println("=== Files Updated Successfully ===")
	})

	err := c.Visit("https://www.bajus.org/gold-price")
	if err != nil {
		fmt.Println("Error visiting website:", err)
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

	// Always append new entry
	fmt.Printf("✅ Adding new record for %s %s to %s\n", priceData.Date, priceData.Time, filename)
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

	// Always append new entry
	fmt.Printf("✅ Adding new JSON entry for %s %s to %s\n", priceData.Date, priceData.Time, filename)
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
