package db

import (
	"fmt"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"log"
	"strings"
)

func (d *DB) GetConsumerId(first string, last string) (string, error) {
	var id string
	sqlStatment := "SELECT id FROM credit_tags WHERE first_name=$1 AND last_name=$2"
	err := d.QueryRow(sqlStatment, first, last).Scan(&id)

	if err != nil {
		return id, err
	}

	return id, nil
}

func (d *DB) AddCustomerAndTags(parsed [][]interface{}) error {

	txn, err := d.Begin()
	if err != nil {
		log.Fatal(err)
	}

	var columnNames []string
	titles := []string{"first_name", "last_name", "full_name", "social_security_number"}
	for _, title := range titles {
		columnNames = append(columnNames, title)
	}
	for i := 1; i < 201; i++ {
		columnNames = append(columnNames, fmt.Sprintf("x%04d", i))
	}

	stmt, err := txn.Prepare(pq.CopyIn("credit_tags", columnNames...))
	if err != nil {
		log.Fatal(err)
	}

	for _, prep := range parsed {
		_, err = stmt.Exec(prep...)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}

	if err != nil {
		log.Fatal(err)
		return err
	}

	_, err = stmt.Exec()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = stmt.Close()
	if err != nil {
		log.Fatal(err)
		return err
	}

	err = txn.Commit()
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}
func (d *DB) GetCreditTags(id string) (map[string]string, error) {
	var sb strings.Builder
	sb.WriteString("SELECT ")
	for i := 1; i < 201; i++ {
		sb.WriteString(fmt.Sprintf("X%04d ", i))
		if i != 200 {
			sb.WriteRune(',')
		}
	}
	sb.WriteString("FROM credit_tags ")
	sb.WriteString("WHERE id = $1")

	fmt.Println(sb.String())
	row, err := d.Query(sb.String(), id)

	resmap := make(map[string]string)

	if err != nil {
		return resmap, err
	}

	cols, _ := row.Columns()

	defer row.Close()
	for row.Next() {

		columns := make([]int, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		row.Scan(columnPointers...)

		//Convert integers to 9 digit width
		for i, colName := range cols {
			resmap[colName] = fmt.Sprintf("%09d ", columns[i])
		}

	}
	return resmap, nil
}
func (d *DB) GetStats(stat string) (map[string]interface{}, error) {

	sqlStatment := `SELECT to_char(median(` + stat + `),'FM999999999.00') as median, to_char(avg(` + stat + `),'FM999999999.00') as mean, to_char(stddev(` + stat + `),'FM999999999.00') as standard_deviation FROM credit_tags WHERE ` + stat + ` > 0`
	rows, err := d.Query(sqlStatment)
	resmap := make(map[string]interface{})
	if err != nil {
		log.Fatal(err)
		return resmap, err
	}

	cols, _ := rows.Columns()

	defer rows.Close()
	for rows.Next() {

		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}
		rows.Scan(columnPointers...)

		for i, colName := range cols {
			resmap[colName] = columns[i]
		}

	}
	return resmap, nil
}
