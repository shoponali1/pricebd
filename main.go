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
	K22         int    `json:"k22"`
	K21         int    `json:"k21"`
	K18         int    `json:"k18"`
	Traditional int    `json:"traditional"`
}

func writeRow(row *[]string, priceData *Price) {
	(*row)[0] = priceData.Date
	(*row)[1] = strconv.Itoa(priceData.K18)
	(*row)[2] = strconv.Itoa(priceData.K21)
	(*row)[3] = strconv.Itoa(priceData.K22)
	(*row)[4] = strconv.Itoa(priceData.Traditional)
}

func main() {
	c := colly.NewCollector()

	todayPrice := Price{}
	todayPrice.Date = time.Now().Format("2006-01-02")

	todaySilverPrice := Price{}
	todaySilverPrice.Date = time.Now().Format("2006-01-02")

	getPrice := func(e *colly.HTMLElement) int {
		priceStr := strings.NewReplacer(",", "", " BDT/GRAM", "").Replace(e.Text)
		price, err := strconv.Atoi(priceStr)
		if err != nil {
			fmt.Println(err)
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
		fmt.Println("Gold:", todayPrice)
		fmt.Println("Silver:", todaySilverPrice)

		savePrice("./fe/src/prices.csv", &todayPrice)
		savePrice("./fe/src/silver-prices.csv", &todaySilverPrice)

		savePriceJSON("./fe/src/prices.json", &todayPrice)
		savePriceJSON("./fe/src/silver-prices.json", &todaySilverPrice)
	})

	c.Visit("https://www.bajus.org/gold-price")
}

func savePrice(filename string, priceData *Price) {
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)

	records, err := csvReader.ReadAll()
	if err != nil {
		panic(err)
	}
	exists := false
	for i := 0; i < len(records); i++ {
		if records[i][0] == priceData.Date {
			writeRow(&records[i], priceData)
			exists = true
			break
		}
	}
	if !exists {
		fmt.Printf("Adding new record for %s to %s\n", priceData.Date, filename)
		records = append(records, make([]string, 5))
		writeRow(&records[len(records)-1], priceData)
	} else {
		fmt.Printf("Updated existing record for %s in %s\n", priceData.Date, filename)
	}
	f.Seek(0, 0)
	f.Truncate(0)
	csv.NewWriter(f).WriteAll(records)
}

func savePriceJSON(filename string, priceData *Price) {
	var prices []Price
	file, err := ioutil.ReadFile(filename)
	if err == nil {
		json.Unmarshal(file, &prices)
	}

	exists := false
	for i, p := range prices {
		if p.Date == priceData.Date {
			prices[i] = *priceData
			exists = true
			break
		}
	}
	if !exists {
		fmt.Printf("Adding new JSON entry for %s to %s\n", priceData.Date, filename)
		prices = append(prices, *priceData)
	} else {
		fmt.Printf("Updated existing JSON entry for %s in %s\n", priceData.Date, filename)
	}

	data, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		panic(err)
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		panic(err)
	}
}
