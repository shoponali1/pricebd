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

	"github.com/playwright-community/playwright-go"
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
	fmt.Println("🚀 Starting browser automation...")

	// Install playwright browsers if needed
	err := playwright.Install(&playwright.RunOptions{
		Verbose: true,
	})
	if err != nil {
		fmt.Printf("⚠️ Could not install playwright: %v\n", err)
	}

	pw, err := playwright.Run()
	if err != nil {
		fmt.Printf("❌ Could not start playwright: %v\n", err)
		os.Exit(1)
	}
	defer pw.Stop()

	// Advanced Stealth Launch Options
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-infobars",
			"--no-sandbox",
			"--disable-setuid-sandbox",
			"--disable-dev-shm-usage",
			"--disable-accelerated-2d-canvas",
			"--disable-gpu",
			"--window-size=1920,1080",
		},
	})
	if err != nil {
		fmt.Printf("❌ Could not launch browser: %v\n", err)
		os.Exit(1)
	}
	defer browser.Close()

	// Robust Browser Context with Headers
	context, err := browser.NewContext(playwright.BrowserNewContextOptions{
		UserAgent: playwright.String("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		Viewport: &playwright.Size{
			Width:  1920,
			Height: 1080,
		},
		HasTouch: playwright.Bool(false),
		IsMobile: playwright.Bool(false),
		ExtraHttpHeaders: map[string]string{
			"Accept-Language":           "en-US,en;q=0.9",
			"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8",
			"Upgrade-Insecure-Requests": "1",
			"Sec-Ch-Ua":                 `"Not_A Brand";v="8", "Chromium";v="120", "Google Chrome";v="120"`,
			"Sec-Ch-Ua-Mobile":          "?0",
			"Sec-Ch-Ua-Platform":        `"Windows"`,
		},
	})
	if err != nil {
		fmt.Printf("❌ Could not create browser context: %v\n", err)
		os.Exit(1)
	}

	page, err := context.NewPage()
	if err != nil {
		fmt.Printf("❌ Could not create page: %v\n", err)
		os.Exit(1)
	}

	// Advanced Stealth Script Injection
	err = page.AddInitScript(playwright.Script{
		Content: playwright.String(`
			// Mask WebDriver
			Object.defineProperty(navigator, 'webdriver', { get: () => undefined });
			
			// Mask Chrome Runtime
			window.chrome = { runtime: {} };
			
			// Mock Plugins
			Object.defineProperty(navigator, 'plugins', { get: () => [1, 2, 3, 4, 5] });
			Object.defineProperty(navigator, 'languages', { get: () => ['en-US', 'en'] });
			
			// Mock Permissions
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: 'denied' }) :
					originalQuery(parameters)
			);
		`),
	})
	if err != nil {
		fmt.Printf("❌ Could not add init script: %v\n", err)
	}

	fmt.Println("🌐 Navigating to bajus.org...")
	if _, err = page.Goto("https://www.bajus.org/gold-price", playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateNetworkidle,
		Timeout:   playwright.Float(60000),
	}); err != nil {
		fmt.Printf("❌ Could not goto page: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("⏳ Waiting for content to load...")

	// Wait for table elements to be visible
	_, err = page.WaitForSelector("table", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(20000),
		State:   playwright.WaitForSelectorStateVisible,
	})
	if err != nil {
		fmt.Printf("⚠️ Could not find table elements: %v\n", err)
	}

	time.Sleep(5 * time.Second)

	now := time.Now()
	todayPrice := Price{
		Date: now.Format("2006-01-02"),
		Time: now.Format("15:04:05"),
	}

	todaySilverPrice := Price{
		Date: now.Format("2006-01-02"),
		Time: now.Format("15:04:05"),
	}

	// Scrape Gold Prices
	fmt.Println("📊 Scraping gold prices...")

	// Retry loop for scraping
	maxRetries := 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		fmt.Printf("🔄 Attempt %d/%d to scrape gold prices...\n", attempt, maxRetries)

		// Debug: Check if tables exist
		tableCount, _ := page.Locator("table").Count()
		fmt.Printf("🔍 DEBUG: Found %d tables on page\n", tableCount)

		goldTableCount, _ := page.Locator(".gold-table").Count()
		fmt.Printf("🔍 DEBUG: Found %d .gold-table elements\n", goldTableCount)

		goldK22, _ := page.Locator(".gold-table tbody tr:nth-child(1) .price").TextContent()
		fmt.Printf("🔍 DEBUG: Raw goldK22 text: '%s'\n", goldK22)
		todayPrice.K22 = parsePrice(goldK22)

		if todayPrice.K22 > 0 {
			fmt.Printf("✅ Successfully scraped gold price: %d\n", todayPrice.K22)
			break
		}

		fmt.Println("⚠️ Failed to scrape gold price. Retrying...")
		time.Sleep(5 * time.Second)

		if attempt == maxRetries {
			fmt.Println("❌ All retry attempts failed!")

			// Save error state
			page.Screenshot(playwright.PageScreenshotOptions{
				Path: playwright.String("error_screenshot.png"),
			})
			html, _ := page.Content()
			os.WriteFile("error_page.html", []byte(html), 0644)

			fmt.Println("📸 Saved error_screenshot.png and error_page.html")
		}
	}

	fmt.Printf("  K22: %d\n", todayPrice.K22)

	goldK21, _ := page.Locator(".gold-table tbody tr:nth-child(2) .price").TextContent()
	todayPrice.K21 = parsePrice(goldK21)
	fmt.Printf("  K21: %d\n", todayPrice.K21)

	goldK18, _ := page.Locator(".gold-table tbody tr:nth-child(3) .price").TextContent()
	todayPrice.K18 = parsePrice(goldK18)
	fmt.Printf("  K18: %d\n", todayPrice.K18)

	goldTraditional, _ := page.Locator(".gold-table tbody tr:nth-child(4) .price").TextContent()
	todayPrice.Traditional = parsePrice(goldTraditional)
	fmt.Printf("  Traditional: %d\n", todayPrice.Traditional)

	// Scrape Silver Prices
	fmt.Println("📊 Scraping silver prices...")

	silverK22, _ := page.Locator(".silver-table tbody tr:nth-child(1) .price").TextContent()
	todaySilverPrice.K22 = parsePrice(silverK22)
	fmt.Printf("  K22: %d\n", todaySilverPrice.K22)

	silverK21, _ := page.Locator(".silver-table tbody tr:nth-child(2) .price").TextContent()
	todaySilverPrice.K21 = parsePrice(silverK21)
	fmt.Printf("  K21: %d\n", todaySilverPrice.K21)

	silverK18, _ := page.Locator(".silver-table tbody tr:nth-child(3) .price").TextContent()
	todaySilverPrice.K18 = parsePrice(silverK18)
	fmt.Printf("  K18: %d\n", todaySilverPrice.K18)

	silverTraditional, _ := page.Locator(".silver-table tbody tr:nth-child(4) .price").TextContent()
	todaySilverPrice.Traditional = parsePrice(silverTraditional)
	fmt.Printf("  Traditional: %d\n", todaySilverPrice.Traditional)

	fmt.Println("\n=== Scraping Completed ===")
	fmt.Printf("Gold: %+v\n", todayPrice)
	fmt.Printf("Silver: %+v\n", todaySilverPrice)

	if todayPrice.K22 == 0 {
		fmt.Println("⚠️ WARNING: Gold prices were not scraped!")
	}
	if todaySilverPrice.K22 == 0 {
		fmt.Println("⚠️ WARNING: Silver prices were not scraped!")
	}

	savePrice("./fe/src/prices.csv", &todayPrice)
	savePrice("./fe/src/silver-prices.csv", &todaySilverPrice)

	savePriceJSON("./fe/src/prices.json", &todayPrice)
	savePriceJSON("./fe/src/silver-prices.json", &todaySilverPrice)

	fmt.Println("✅ Files updated successfully!")
}

func parsePrice(priceStr string) int {
	cleaned := strings.NewReplacer(",", "", " BDT/GRAM", "", " ", "").Replace(priceStr)
	price, err := strconv.Atoi(cleaned)
	if err != nil {
		fmt.Printf("⚠️ Error parsing price '%s': %v\n", priceStr, err)
		return 0
	}
	return price
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

	fmt.Printf("📝 Adding record to %s\n", filename)
	newRecord := make([]string, 6)
	writeRow(&newRecord, priceData)
	records = append(records, newRecord)

	f.Seek(0, 0)
	f.Truncate(0)
	writer := csv.NewWriter(f)
	writer.WriteAll(records)
	writer.Flush()
}

func savePriceJSON(filename string, priceData *Price) {
	var prices []Price
	file, err := ioutil.ReadFile(filename)
	if err == nil {
		json.Unmarshal(file, &prices)
	}

	fmt.Printf("📝 Adding JSON entry to %s\n", filename)
	prices = append(prices, *priceData)

	data, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(filename, data, 0644)
}
