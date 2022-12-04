package main

import (
	"log"
	"os"
	entries "tackytortoise/readart/entries"
	"tackytortoise/readart/visualization"
)

func getAllBooks() []entries.BookLog {
	var books []entries.BookLog
	dir := "./input/"
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range files {
		books = append(books, entries.NewBookLogFromFile(dir+f.Name()))
	}
	return books
}

func drawAllBookChart() {
	visualization.CreateLineChart(getAllBooks(), "allBooks", false)
}

func drawIndividualFiles() {
	books := getAllBooks()
	for _, b := range books {
		current := make([]entries.BookLog, 1)
		current[0] = b
		visualization.CreateLineChart(current, b.Name, true)
	}
}

func main() {
	drawAllBookChart()
	drawIndividualFiles()
	visualization.CreateAvgPageChart(getAllBooks(), "Running 14 day average", 14)
	visualization.CreateTotalPagesRead(getAllBooks(), "Total pages read")
}
