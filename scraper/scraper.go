package scraper

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"example.com/go-fiis-scraper/repository"

	"github.com/gocolly/colly"
)

// FundData is ...
type FundData struct {
	Ticker               string
	Price                int
	DailyLiquidity       int
	Dividend             int
	DividendYield        int
	EquityValue          int
	MonthlyProfitability int
	PricePerBookValue    int
	BondsAmount          int
	Date                 time.Time
}

func escapeMonetaryToken(str string) string {
	return strings.Replace(str, "R$ ", "", -1)
}

func escapeDots(str string) string {
	return strings.Replace(str, ".", "", -1)
}

func escapeAndExtractSuffix(str string) string {
	return strings.Replace(str, ".", "", -1)
}

func replaceCommasWithDots(str string) string {
	return strings.Replace(str, ",", ".", -1)
}

func stringToInt(str string) int {
	scapedValue := escapeDots(str)

	numericValue := 0

	if str != "N/A" {
		parsedValue, err := strconv.Atoi(scapedValue)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		numericValue = parsedValue
	}

	return numericValue * 100
}

func stringToFloat(str string) int {
	percentize := false
	scapedValue := replaceCommasWithDots(escapeMonetaryToken(escapeDots(str)))

	if strings.Contains(scapedValue, "%") {
		percentize = true
		scapedValue = strings.Replace(scapedValue, "%", "", -1)
	}

	numericValue := 0.00

	if str != "N/A" {
		parsedValue, err := strconv.ParseFloat(scapedValue, 12)

		if err != nil {
			fmt.Println(err)
			os.Exit(2)
		}

		numericValue = parsedValue
	}

	if percentize == true {
		return int(numericValue * 100)
	}

	return int(numericValue * 100)
}

func (f *FundData) parseAndInsertValue(title string, value string) {
	switch title {
	case "Liquidez Diária":
		f.DailyLiquidity = stringToInt(value)
	case "Último Rendimento":
		f.Dividend = stringToFloat(value)
	case "Dividend Yield":
		f.DividendYield = stringToFloat(value)
	case "Cotas emitidas":
		f.BondsAmount = stringToInt(value)
	case "Valor Patrimonial":
		f.EquityValue = stringToFloat(value)
	case "Rentab. no mês":
		f.MonthlyProfitability = stringToFloat(value)
	case "P/VP":
		f.PricePerBookValue = stringToFloat(value)
	case "Preço Atual":
		f.Price = stringToFloat(value)
	default:
	}
}

func scrapeURL(ticker string, url string, c *colly.Collector, wg *sync.WaitGroup, collection *repository.MongoDBCollection) {
	var fundData FundData

	c.OnHTML("#main-indicators-carousel", func(e *colly.HTMLElement) {
		e.ForEach(".carousel-cell", func(i int, e *colly.HTMLElement) {
			title := e.ChildText(".indicator-title")
			value := e.ChildText(".indicator-value")

			fundData.parseAndInsertValue(title, value)
		})

		fundData.Ticker = ticker
		fundData.Date = time.Now()
	})

	c.OnHTML("#basic-infos", func(e *colly.HTMLElement) {
		title := e.ChildText(".col-md-6:first-child > ul > li:nth-child(3) > .text-wrapper > .title")
		value := e.ChildText(".col-md-6:first-child > ul > li:nth-child(3) > .text-wrapper > .description")

		if title != "Cotas emitidas" {
			log.Fatal("Requested title is not 'Cotas emitidas")
		}

		fundData.parseAndInsertValue(title, value)
	})

	c.OnHTML("#stock-price-wrapper", func(e *colly.HTMLElement) {
		value := e.ChildText(".price")

		fundData.parseAndInsertValue("Preço Atual", value)
	})

	c.OnScraped(func(r *colly.Response) {
		log.Println("Finished scraping URL:", r.Request.URL)

		collection.Add(fundData)

		wg.Done()
	})

	c.Visit(url)
}

func scrapeTicker(ticker string, wg *sync.WaitGroup, collection *repository.MongoDBCollection) {
	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		log.Println("Scraping URL:", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
	})

	url := os.Getenv("SERVICE_BASE_URL") + strings.ToLower(ticker)

	scrapeURL(ticker, url, c, wg, collection)
}

// Scrape returns nil
func Scrape(tickers []string, collection *repository.MongoDBCollection) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(tickers))

	for _, ticker := range tickers {
		go scrapeTicker(ticker, &waitGroup, collection)
	}

	waitGroup.Wait()

	collection.ConsumeAddQueue()
}
