package modules

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"log"
	"os"
	"score_query_server/utils"
	"time"
)

type Model struct {
	ID        uint `json:"id" gorm:"primary_key"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" sql:"index"`
}

var (
	db *gorm.DB
	dbName = os.Getenv("score_query_db_name")
	dbPassword = os.Getenv("score_query_db_password")
	dbHost = os.Getenv("score_query_db_host")
)

func Init() {

	var err error

	db, err = gorm.Open("mysql",
		fmt.Sprintf("%s:%s@tcp(%s:3306)/score_query?charset=utf8&parseTime=True&loc=Local", dbName, dbPassword, dbHost))
	if err != nil {
		log.Panic(err)
	}
	// 退出时运行
	utils.RunOnExit(db.Close)

	if !db.HasTable(new(Id)) {
		db.CreateTable(new(Id))
	}
	if !db.HasTable(new(Class)) {
		db.CreateTable(new(Class))
	}
	if !db.HasTable(new(Score)) {
		db.CreateTable(new(Score))
	}
	if !db.HasTable(new(Course)) {
		db.CreateTable(new(Course))
	}

	db.AutoMigrate(new(Id))
	db.AutoMigrate(new(Class))
	db.AutoMigrate(new(Score))
	db.AutoMigrate(new(Course))
}