package commitor

import (
	"database/sql"
	"sync"
	"time"

	"dnf"
	_ "github.com/ziutek/mymysql/godrv"
)

var once sync.Once
var db *sql.DB

type dbConf struct {
	ip       string
	port     string
	dbname   string
	username string
	passwd   string
}

func loadDbConf() *dbConf {
	// TODO: load conf from file

	/* hehe: I won't commit those info to my github */
	// return &dbConf{
	// 	ip:       "someip",
	// 	port:     "someport",
	// 	dbname:   "somedb",
	// 	username: "someuser",
	// 	passwd:   "somepwd",
	// }

	return &dbConf{
		ip:       "qtmysql",
		port:     "3306",
		dbname:   "adserver-stable",
		username: "root",
		passwd:   "qazxs913",
	}
}

func Init() {
	once.Do(func() {
		var err error
		cf := loadDbConf()
		db, err = sql.Open("mymysql",
			"tcp:"+cf.ip+":"+cf.port+"*"+cf.dbname+"/"+cf.username+"/"+cf.passwd)
		if err != nil {
			panic(err)
		}
	})
}

func CommitLoop() {
	adCommit()
	dnf.DisplayDocs()
	for {
		time.Sleep(1 * time.Minute)
		//now := time.Now()
		//if now.Hour() == 3 && now.Minute() == 2 {
		adCommit()
		dnf.DisplayDocs()
		//}
	}
}
