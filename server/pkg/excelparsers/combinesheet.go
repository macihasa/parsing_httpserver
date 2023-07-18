package excelparsers

import (
	"encoding/csv"
	"log"
	"mime/multipart"
	"os"
	"runtime"
	"sync"

	"github.com/xuri/excelize/v2"
)

const MAX_NUM_GOROUTINES = 10

func CombineSheet(form *multipart.Form, rowsch chan []string, sheetname string, finished chan<- bool) {

	readwg := new(sync.WaitGroup)
	writewg := new(sync.WaitGroup)

	limiter := make(chan int, MAX_NUM_GOROUTINES)

	writewg.Add(1)
	go writer("OutputXL", rowsch, writewg)

	filesprocessed := 0

	for _, fileheaders := range form.File {
		for _, fileheader := range fileheaders {
			if fileheader.Size == 0 {
				continue
			}
			log.Println("before limiter, goroutines active:", runtime.NumGoroutine())
			limiter <- 1
			log.Println(limiter, "After limiter, goroutines active:", runtime.NumGoroutine())
			readwg.Add(1)
			go reader(fileheader, sheetname, rowsch, limiter, readwg)
			filesprocessed++
			if filesprocessed%10 == 0 {
				log.Println("Files processed:", filesprocessed, "goroutines active:", runtime.NumGoroutine())
			}
		}
	}

	readwg.Wait()
	close(rowsch)
	writewg.Wait()

	close(limiter)

	finished <- true
}

func reader(fileheader *multipart.FileHeader, sheetname string, rowsch chan []string, limiter chan int, readwg *sync.WaitGroup) {
	defer readwg.Done()
	file, err := fileheader.Open()
	if err != nil {
		log.Println("Unable to open file ", err)
		file.Close()
		<-limiter
		return
	}
	defer file.Close()

	xlfile, err := excelize.OpenReader(file)
	if err != nil {
		log.Println("Unable to open excel file ", err)
		file.Close()
		<-limiter
		return
	}

	if sheetname == "1" {
		sheetname = xlfile.GetSheetName(0)
	}

	rows, err := xlfile.Rows(sheetname)
	if err != nil {
		log.Println("Unable to read rows from sheet:", sheetname, err)
		return
	}

	rowcount := 0

	for rows.Next() {
		row, err := rows.Columns()
		rowcount++
		if rowcount%500 == 0 {
			log.Println("Rows processed:", rowcount, "goroutines active:", runtime.NumGoroutine())
		}
		if err != nil {
			log.Println("Unable to read row from sheet:", sheetname, err)
			continue
		}
		if len(row) == 0 {
			continue
		}
		rowsch <- row
	}
	<-limiter
}

func writer(fileName string, rowsch <-chan []string, wg *sync.WaitGroup) {
	defer wg.Done()
	file, err := os.Create(fileName + ".csv")
	if err != nil {
		log.Fatal("Unable to create output file: ", err)
	}

	csvWriter := csv.NewWriter(file)
	csvWriter.Comma = rune(';')
	log.Println("Wrote header row to csv file")

	var rowindex int
	for row := range rowsch {
		err := csvWriter.Write(row)
		if err != nil {
			log.Println(err)
		}
		rowindex++
	}
	csvWriter.Flush()
	log.Println("Flushed csv writer")
	file.Close()
}
