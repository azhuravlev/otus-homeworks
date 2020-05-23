package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/viper"
)

type ExecStats struct {
	Affected     int64 `json:"affected"`
	LastInsertId int64 `json:"last_insert_id,omitempty"`
}

func dbConn() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", viper.GetString("database"))
	if err != nil {
		return nil, err
	}
	return db, nil
}
