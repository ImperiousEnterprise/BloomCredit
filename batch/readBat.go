package main

import (
	"BloomCredit/db"
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type column struct {
	start int
	end   int
}
type worker struct {
	sentence string
	columns  []column
}
type syncedArray struct {
	parseArray [][]interface{}
	syc        sync.Mutex
	counter    int
}

var LinesRead = 1

func main() {
	fmt.Println(os.Args)
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		"localhost", 5432, "postgres", "pass", "postgres")

	db, err := db.NewDB(psqlInfo)
	if err != nil {
		log.Panic(err)
	}
	start := time.Now()
	LinebyLineScan(os.Args[1], db)
	elapsed := time.Since(start)
	log.Printf("Parsing took %s", elapsed)
}
func LinebyLineScan(filepath string, db db.Store) {

	file, err := os.Open(filepath)
	if err != nil {
		log.Fatal(err)
	}
	jobs := make(chan worker, 100)
	records := make(chan []interface{})
	done := make(chan bool, 1)

	wg := new(sync.WaitGroup)
	batch := new(sync.WaitGroup)
	finish := new(sync.WaitGroup)

	for w := 1; w <= 10; w++ {
		wg.Add(1)
		go CallParseRecords(jobs, wg, records)
	}

	var array syncedArray

	batch.Add(1)
	go BatchInsertRecordsToDB(records, &array, &db, batch)

	finish.Add(1)
	go ProcessLastRecord(done, &array, &db, finish)

	scanner := bufio.NewScanner(file)
	defer file.Close()
	var columns []column
	counter := 0
	for scanner.Scan() {
		ken := scanner.Text()
		if counter == 0 {
			columns = generateColumns(ken)
		} else {
			jobs <- worker{ken, columns}
		}
		counter++
	}

	if err := scanner.Err(); err != nil {
		fmt.Println(err)
		log.Fatal(err)
		return
	}
	close(jobs)

	//Wait until jobs done and then get final record
	wg.Wait()
	done <- true
	finish.Wait()
}
func ProcessLastRecord(isFinished <-chan bool, ts *syncedArray, db db.Store, finish *sync.WaitGroup) {
	defer finish.Done()
	done := <-isFinished

	if done {
		fmt.Printf(" last number of parsedArray is %d\n", len(ts.parseArray))
		ts.insertIntoDB(db)
	}

}
func (cs *syncedArray) Append(item []interface{}) {
	cs.syc.Lock()
	defer cs.syc.Unlock()

	cs.counter++
	cs.parseArray = append(cs.parseArray, item)
}
func (cs *syncedArray) insertIntoDB(db db.Store) {
	cs.syc.Lock()
	defer cs.syc.Unlock()
	fmt.Printf(" parseArray is %d at line %d\n", cs.counter, LinesRead)
	err := db.AddCustomerAndTags(cs.parseArray)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	cs.parseArray = nil
}
func BatchInsertRecordsToDB(records <-chan []interface{}, ts *syncedArray, db *db.Store, batch *sync.WaitGroup) {
	for record := range records {
		ts.Append(record)
		if len(ts.parseArray) == 10000 {
			ts.insertIntoDB(db)
		}
		LinesRead++
	}

	batch.Done()
}

func CallParseRecords(jobs <-chan worker, wg *sync.WaitGroup, records chan<- []interface{}) {
	for j := range jobs {
		parse, err := parseRecords(j.sentence, j.columns)
		if err != nil {
			log.Fatal(err)
			break
		}
		records <- parse
	}
	wg.Done()
}

func parseRecords(ken string, columns []column) ([]interface{}, error) {
	var column []interface{}
	fullName := strings.TrimLeft(ken[columns[0].start:columns[0].end], " ")
	splitName := strings.Split(fullName, " ")
	ssn, _ := strconv.Atoi(strings.TrimLeft(ken[columns[1].start:columns[1].end], " "))

	length := len(splitName)
	if length > 2 {
		column = append(column, strings.ToLower(splitName[1]), strings.ToLower(splitName[2]), fullName, ssn)
	} else {
		column = append(column, strings.ToLower(splitName[0]), strings.ToLower(splitName[1]), fullName, ssn)
	}

	for pos := 2; pos < len(columns); pos++ {
		num, err := strconv.Atoi(strings.TrimLeft(ken[columns[pos].start:columns[pos].end], " "))
		if err != nil {
			log.Fatal(err)
		}

		//ensuring credit tags are 9 digits in width
		if num < -1000000000 && num > 100000000 {
			return nil, errors.New("credit tag out of range")
		}
		column = append(column, num)
	}
	return column, nil
}

func generateColumns(text string) []column {
	var columns []column

	start := 0
	length := len(text)
	for pos, char := range text {
		startPos := func() int {
			if pos-1 > 0 {
				return pos - 1
			}
			return 0
		}()
		if startPos != pos && char == ' ' && text[startPos:pos] != " " {
			columns = append(columns, column{start, pos})
			start = pos
		}
	}

	columns = append(columns, column{start, length})

	return columns
}
