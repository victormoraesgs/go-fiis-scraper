package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"example.com/go-fiis-scraper/csv"
	"example.com/go-fiis-scraper/repository"
	"example.com/go-fiis-scraper/scraper"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func loadEnv() {
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Error loading .env file")
	}

	log.Println("environment variables loaded.")
}

func loadFundsList() []string {
	content, err := ioutil.ReadFile("funds.txt")

	if err != nil {
		log.Println(err)
	}

	textContent := string(content)
	return strings.Split(textContent, "\n")
}

func aggregateCollection(collection *repository.MongoDBCollection) []bson.M {
	sortStage := bson.D{
		{"$sort", bson.D{
			{"ticker", 1},
			{"date", 1},
		}},
	}

	groupStage := bson.D{
		{"$group", bson.D{
			{"_id", "$ticker"},
			{"price", bson.D{{"$last", "$price"}}},
			{"priceperbookvalue", bson.D{{"$last", "$priceperbookvalue"}}},
			{"date", bson.D{{"$last", "$date"}}},
			{"dividend", bson.D{{"$last", "$dividend"}}},
			{"dividendyield", bson.D{{"$last", "$dividendyield"}}},
		}},
	}

	cursor, err := collection.Collection.Aggregate(context.TODO(), mongo.Pipeline{sortStage, groupStage})
	if err != nil {
		log.Fatal(err)
	}

	var results []bson.M
	if err = cursor.All(context.TODO(), &results); err != nil {
		log.Fatal(err)
	}

	return results
}

func getDateStringBR(date time.Time) string {
	return fmt.Sprintf("%02d/%02d/%d",
		date.Day(), date.Month(), date.Year())
}

func generateReport(collection *repository.MongoDBCollection) {
	reportFile := csv.GetCSV("./output/", os.Getenv("REPORT_FILENAME"))
	reportFile.Add([]string{"Ticker", "Link", "Preço", "Dividendo", "Dividend Yield", "P/VP", "Data"})

	results := aggregateCollection(collection)

	log.Printf("Generating report with %v lines.", len(results))

	for _, result := range results {
		reportFile.Add([]string{
			result["_id"].(string),
			os.Getenv("SERVICE_BASE_URL") + result["_id"].(string),
			fmt.Sprintf("%.2f", float64(result["price"].(int32))/100),
			fmt.Sprintf("%.2f", float64(result["dividend"].(int32))/100),
			fmt.Sprintf("%.2f", float64(result["dividendyield"].(int32))/100),
			fmt.Sprintf("%.2f", float64(result["priceperbookvalue"].(int32))/100),
			getDateStringBR(result["date"].(primitive.DateTime).Time()),
		})
	}

	log.Println("Report generated.")
}

func main() {
	startTime := time.Now()
	log.Println("Starting process.")

	loadEnv()

	collection := repository.GetCollection("fiis")
	tickers := loadFundsList()

	log.Println(tickers)

	scraper.Scrape(tickers, &collection)
	generateReport(&collection)

	log.Println("Process finished in", time.Now().Sub(startTime))
}
