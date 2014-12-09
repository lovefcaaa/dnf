package commitor

import (
	"database/sql"
	"sync"

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
	return &dbConf{
		ip:       "someip",
		port:     "someport",
		dbname:   "somedb",
		username: "someuser",
		passwd:   "somepwd",
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
