package sssql

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	// SQLite
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	_ "github.com/ziutek/mymysql/godrv"
)

type USQL struct {
	D      *sql.DB
	Engine string
}

func (_s *USQL) Init(mode, connect string, initDB string) int {
	var err error

	switch strings.ToLower(mode) {
	case "my":
		mode = "mymysql"
		_s.Engine = "MySQL"
	case "pg":
		mode = "postgres"
		_s.Engine = "PostgreSQL"
	case "sqlite":
		mode = "sqlite3"
		_s.Engine = "SQLite"
	default:
		log.Printf("ERROR in SQL Engine selector: \"%s\"! Allowed only \"PG\", \"MY\" and \"SQLite\"!\n", mode)
		os.Exit(1)
	}

	_s.D, err = sql.Open(mode, connect)
	if err != nil {
		log.Printf("%s::Open() error: %v\n", _s.Engine, err)
		return -1
	}

	if _s.Engine == "SQLite" {
		err = _s.D.Ping()
		if err != nil {
			log.Printf("SQLite::Ping() error: %v\n", err)
			log.Printf("%s::Ping() error: %v\n", _s.Engine, err)
			return -1
		}
	}

	if len(initDB) > 10 {
		_, err = _s.D.Exec(initDB)
		if err != nil {
			log.Printf("%s::Exec() InitDB error: %v\n", _s.Engine, err)
			return -1
		}
	}

	return 0
}

func (_s *USQL) Close() {
	_s.D.Close()
}

func (_s *USQL) QSimple(query ...interface{}) int {
	var (
		i       int
		queryGo = string("")
	)
	for _, x := range query {
		queryGo = fmt.Sprintf("%s%v", queryGo, x)
	}
	res, err := _s.D.Query(queryGo)
	if err != nil {
		log.Printf("SQL::Query() QSimple error: %v\n", err)
		log.Printf("Query: %s\n", queryGo)
		return -1
	}

	res.Next()
	res.Scan(&i)

	return i
}
