package visualization

import (
	"fmt"
	"log"
	"math"
	"os"
	"tackytortoise/readart/entries"
	"time"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
)

func dateMatches(one, two time.Time) bool {
	return one.Day() == two.Day() && one.Month() == two.Month() && one.Year() == two.Year()
}

func generateLineItems(book entries.BookLog, startDate, endDate time.Time) []opts.LineData {
	totalDays := int(math.Ceil(endDate.Sub(startDate).Hours()/24)) + 1
	currentDate := startDate.Add(-time.Hour * 24) // Start 1 day earlier cause we increment at start of loop
	indexToCheck := 0
	lastPageEntry := 0
	items := make([]opts.LineData, totalDays)

	// Create line point for every day
	for i := 0; i < totalDays; i++ {
		// Jump to next day
		currentDate = currentDate.Add(time.Hour * 24)

		// Check if we hit a read date
		if indexToCheck < len(book.Entries) {
			current := book.Entries[indexToCheck]
			if dateMatches(currentDate, current.Date) {
				items[i] = opts.LineData{Value: current.Page, Symbol: "none"}
				lastPageEntry = current.Page
				indexToCheck++
				continue
			}
		}

		// Add empty entry before start or after end of book
		if lastPageEntry == 0 || indexToCheck >= len(book.Entries) {
			items[i] = opts.LineData{Value: "=", Symbol: "none"}
		} else {
			// If not re-add last entry
			items[i] = opts.LineData{Value: lastPageEntry, Symbol: "none"}
		}
	}
	return items
}

// Get slice of every single day between first book start and last book finish
func getAllBookDates(books []entries.BookLog) []time.Time {
	startDate := time.Now()
	endDate := time.Time{}

	// Find start and end date of all books combined
	for _, book := range books {
		bookEnd := book.GetLastDate()
		if bookEnd.After(endDate) {
			endDate = bookEnd
		}
		bookStart := book.GetFirstDate()
		if bookStart.Before(startDate) {
			startDate = bookStart
		}
	}

	// Create array of all dates between start date and end date
	totalDays := int(math.Ceil(endDate.Sub(startDate).Hours()/24)) + 1
	allDates := make([]time.Time, totalDays)
	currentDate := startDate
	for i := 0; i < len(allDates); i++ {
		allDates[i] = currentDate
		currentDate = currentDate.Add(time.Hour * 24)
	}

	return allDates
}

// Get only months from slice of dates
func datesToAxis(dates []time.Time, isSingle bool) []string {
	result := make([]string, len(dates))
	for i, d := range dates {
		if isSingle {
			result[i] = fmt.Sprintf("%v/%v", d.Day(), int(d.Month()))
		} else {
			result[i] = fmt.Sprintf("%v", d.Month())
		}
	}
	return result
}

func setupLine(name string, isSingle bool, yAxisName string) *charts.Line {
	// create a new line instance
	line := charts.NewLine()

	chartName := "Books read 2022"
	chartSize := "1440px"
	if isSingle {
		chartName = name
		chartSize = "720px"
	}

	// set some global options like Title/Legend/ToolTip or anything else
	line.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: "Books read 2022",
			Theme:     types.ThemeInfographic,
			Width:     chartSize,
		}),
		charts.WithTitleOpts(opts.Title{
			Title: chartName,
		}),
		charts.WithXAxisOpts(opts.XAxis{
			Name: "Time",
			Type: "category",
			AxisLabel: &opts.AxisLabel{
				ShowMinLabel: true,
				ShowMaxLabel: true,
			},
		}),
		charts.WithYAxisOpts(opts.YAxis{
			Name: yAxisName,
			Type: "value",
		}),
		charts.WithLegendOpts(opts.Legend{
			Show:   !isSingle,
			Orient: "vertical",
			Align:  "right",
			X:      "right",
		}))
	return line
}

func CreateLineChart(books []entries.BookLog, name string, isSingle bool) {
	line := setupLine(name, isSingle, "Page Number")

	bookDates := getAllBookDates(books)
	startDate := bookDates[0]
	endDate := bookDates[len(bookDates)-1]
	allDates := datesToAxis(bookDates, isSingle)
	line.SetXAxis(allDates)

	// Put data into instance
	for _, book := range books {
		data := generateLineItems(book, startDate, endDate)
		line.AddSeries(book.Name, data).SetSeriesOptions(
			charts.WithLineChartOpts(opts.LineChart{Smooth: true}))
	}

	renderLine(name, line)
}

func renderLine(name string, line *charts.Line) {
	os.Mkdir("./output", os.ModePerm)
	f, err := os.Create(fmt.Sprintf("./output/%v.html", name))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	err = line.Render(f)
	if err != nil {
		log.Fatal(err)
	}
}

// Get amount of pages read per day
func getPagesReadPerDay(books []entries.BookLog) map[time.Time]int {
	totalPagesPerDay := make(map[time.Time]int)
	for _, book := range books {
		for i, entry := range book.Entries {
			read := entry.Page
			if i > 0 {
				read -= book.Entries[i-1].Page
			}
			totalPagesPerDay[entry.Date] += read
		}
	}
	return totalPagesPerDay
}

// Create chart displaying average pages per {running_days}
func CreateAvgPageChart(books []entries.BookLog, name string, running_days int) {
	line := setupLine(name, true, "Page Count")

	// Create map of pages read per day
	totalPagesPerDay := getPagesReadPerDay(books)

	bookDates := getAllBookDates(books)
	items := make([]opts.LineData, len(bookDates))
	runningTotal := make([]int, 0)

	// Loop over all dates in reading range
	for i, date := range bookDates {
		ok := false
		pageCount := 0
		// Find matching day by date, ignore hours...
		for d, pc := range totalPagesPerDay {
			if d.Day() == date.Day() && d.Month() == date.Month() && d.Year() == date.Year() {
				ok = true
				pageCount = pc
			}
		}

		// If we found entry add pages read, otherwise add 0 pages read
		if ok {
			runningTotal = append(runningTotal, pageCount)
		} else {
			runningTotal = append(runningTotal, 0)
		}

		// Clamp list to running days length
		if len(runningTotal) > running_days {
			runningTotal = runningTotal[1:]
		}

		// Calculate average
		avg := float64(0)
		for _, e := range runningTotal {
			avg += float64(e)
		}
		avg = avg / float64(len(runningTotal))

		// Add average entry
		items[i] = opts.LineData{Value: avg, Symbol: "none"}
	}

	allDates := datesToAxis(bookDates, false)
	line.SetXAxis(allDates)

	line.AddSeries(fmt.Sprintf("Avg pages per %v days", running_days), items).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	renderLine(name, line)
}

func CreateTotalPagesRead(books []entries.BookLog, name string) {
	line := setupLine(name, true, "Page Total")

	// Create map of pages read per day
	totalPagesPerDay := getPagesReadPerDay(books)

	bookDates := getAllBookDates(books)
	items := make([]opts.LineData, len(bookDates))

	totalCount := 0
	for i, date := range bookDates {
		// Find matching day by date, ignore hours...
		for d, pc := range totalPagesPerDay {
			if d.Day() == date.Day() && d.Month() == date.Month() && d.Year() == date.Year() {
				totalCount = totalCount + pc
			}
		}

		// Add entry for total pages
		items[i] = opts.LineData{Value: totalCount, Symbol: "none"}
	}

	allDates := datesToAxis(bookDates, false)
	line.SetXAxis(allDates)

	line.AddSeries("Total pages read", items).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	renderLine(name, line)
}
