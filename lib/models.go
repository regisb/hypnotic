package hypnotic

import (
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
)
import "fmt"
import "time"

type User struct {
	ID     string `sql:"type:varchar(100);not null;unique_index"`
	Videos []Video
}

type Video struct {
	ID               string `sql:"type:varchar(100);not null;unique_index"`
	UserID           string `sql:"type:varchar(100);not null;index"`
	OriginalFilename string `sql:"type:varchar(255);not null"`
	Published        bool   `sql:"index;default:true"`
	// TODO store sha256 of file and don't transcode if it already exists for this user
}

type TranscodingJob struct {
	ID        int `sql:"AUTO_INCREMENT"`
	Video     Video
	VideoID   string
	Status    string `sql:"type:varchar(100);index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Stdout    string `sql:"type:varchar(1000)"`
	Stderr    string `sql:"type:varchar(1000)"`
}

func Db() *gorm.DB {
	// TODO is there a smarter way to handle DB singletons
	db, err := gorm.Open("sqlite3", "./hypnotic.sql") // TODO smarter directory
	if err != nil {
		// TODO shouldn't we stop here?
		fmt.Println("Error while opening database:", err)
	}
	db.DB()
	return &db
}

func MigrateDb() {
	Db().AutoMigrate(&User{}, &Video{}, &TranscodingJob{}) // Migrate only when necessary
}
