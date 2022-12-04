package entries

import (
	"bufio"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type LogEntry struct {
	Date time.Time
	Page int
}

type BookLog struct {
	Entries []LogEntry
	Name    string
}

func (b BookLog) GetFirstDate() time.Time {
	return b.Entries[0].Date
}

func (b BookLog) GetLastDate() time.Time {
	return b.Entries[len(b.Entries)-1].Date
}

// Parses book entry from a file
func NewBookLogFromFile(path string) BookLog {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	book := BookLog{Name: filepath.Base(file.Name())}
	// Go over each line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		splits := strings.Split(line, " ")

		// If there is only 1 item on the line, this should be the last line, so final page
		if len(splits) == 1 {
			page, err := strconv.Atoi(splits[0])
			if err != nil {
				log.Fatal(err)
			}
			// Take previous entry date + 1
			day := book.Entries[len(book.Entries)-1].Date.Add(time.Hour * 24)
			entry := LogEntry{Date: day, Page: page}
			book.Entries = append(book.Entries, entry)
			break
		}

		// Parse day and month
		dateText := splits[0]
		dateSplit := strings.Split(dateText, "/")
		day, err := strconv.Atoi(dateSplit[0])
		if err != nil {
			log.Fatal(err)
		}
		month, err := strconv.Atoi(dateSplit[1])
		if err != nil {
			log.Fatal(err)
		}

		// Parse year if provided
		year := time.Now().Year()
		if len(dateSplit) > 2 {
			year, err = strconv.Atoi(dateSplit[2])
			if err != nil {
				log.Fatal(err)
			}
		}

		// Construct time
		entryDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)

		page, err := strconv.Atoi(splits[1])
		if err != nil {
			log.Fatal(err)
		}
		if page == 0 {
			page = 1
		}

		// If first entry is not first page, add one day earlier at page one for nice graphs
		if len(book.Entries) == 0 && page > 1 {
			prevDate := entryDate.Add(-time.Hour * 24)
			fakeEntry := LogEntry{Date: prevDate, Page: 1}
			book.Entries = append(book.Entries, fakeEntry)
		}

		// Register entry
		entry := LogEntry{Date: entryDate, Page: page}
		book.Entries = append(book.Entries, entry)
	}
	// // Make charts start at 1 for prettier lines
	// book.Entries[0].Page = 1
	return book
}
