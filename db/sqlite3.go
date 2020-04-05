package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
	"os"
	"path"
)

func GetOrmDb() *gorm.DB {
	cwd, err := os.Getwd()
	if err != nil {
		log.Error(err)
		return nil
	}
	//todo: create the directory if needed.
	dbPath := path.Join(cwd, "db", "main.db")
	db, err := gorm.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Error with db %v: %v", dbPath, err)
	}
	return db
}