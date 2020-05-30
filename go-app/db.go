package main

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/viper"
	"strings"
)

type ExecStats struct {
	Affected     int64 `json:"affected"`
	LastInsertId int64 `json:"last_insert_id,omitempty"`
}

func runMigrations(dir string) error {
	db, err := dbConn()
	if err != nil {
		return err
	}
	defer db.Close()
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithDatabaseInstance("file://./migrations", "mysql", driver)
	if err != nil {
		return err
	}
	defer m.Close()
	if strings.ToLower(dir) == "up" {
		err = m.Up()
		if err != nil {
			return err
		}
	}
	if strings.ToLower(dir) == "down" {
		err = m.Down()
		if err != nil {
			return err
		}
	}
	return nil
}

func dbConn() (db *sql.DB, err error) {
	db, err = sql.Open("mysql", viper.GetString("database"))
	if err != nil {
		return nil, err
	}
	return db, nil
}
