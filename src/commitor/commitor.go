package commitor

import (
	"database/sql"
	"encoding/json"
	"os"
	"sync"
	"time"

	"dnf"
	_ "github.com/ziutek/mymysql/godrv"
)

var once sync.Once
var db *sql.DB

type dbConf struct {
	Ip       string
	Port     string
	Dbname   string
	Username string
	Passwd   string
}

func loadDbConf() *dbConf {
	var conf dbConf
	f, err := os.Open("conf/db.json")
	if err != nil {
		panic(err)
	}
	decoder := json.NewDecoder(f)
	if err = decoder.Decode(&conf); err != nil {
		panic(err)
	}
	return &conf
}

func Init() {
	once.Do(func() {
		var err error
		cf := loadDbConf()
		db, err = sql.Open("mymysql",
			"tcp:"+cf.Ip+":"+cf.Port+"*"+cf.Dbname+"/"+cf.Username+"/"+cf.Passwd)
		if err != nil {
			panic(err)
		}
	})
}

func CommitLoop() {
	adCommit()
	zoneInfoCommit()
	dnf.DisplayDocs()
	for {
		time.Sleep(1 * time.Minute)
		adCommit()
		zoneInfoCommit()
		dnf.DisplayDocs()
	}
}
