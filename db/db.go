package db

import "database/sql"

type Store interface {
	AddCustomerAndTags([][]interface{}) error
	GetConsumerId(string, string) (string, error)
	GetCreditTags(string) (map[string]string, error)
	GetStats(string) (map[string]interface{}, error)
}

type DB struct {
	*sql.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return &DB{db}, nil
}
