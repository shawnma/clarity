package logging

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"shawnma.com/clarity/config"
)

type consoleLogger struct{}

func (c *consoleLogger) Log(l *HttpLog) {
	log.Printf("ACCESS: %s", l)
}

type MysqlLogger struct {
	db *sql.DB
}

func newMysqlLogger(c *config.Config) (*MysqlLogger, error) {
	if c.Logs.Config["url"] == "" {
		return nil, errors.New("no URL provided for DB Logger")
	}
	log.Print("Creating DB logger")
	db, err := sql.Open("mysql", c.Logs.Config["url"])
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &MysqlLogger{db}, nil
}

func (logger *MysqlLogger) Log(l *HttpLog) {
	stmt := `INSERT INTO LOG(RemoteAddr, Method,RequestContentType,RequestLength,RequestBody,
		ResponseCode,ResponseContentType,ResponseLength,ResponseBody,Title,URL,LogTime)
		 VALUES (?, ?, ?, ?,?,?,?,?,?,?,?,?)`
	_, e := logger.db.Exec(stmt, l.RemoteAddr, l.Method,
		l.RequestContentType, l.RequestLength, l.RequestBody,
		l.ResponseCode, l.ResponseContentType, l.ResponseLength, l.ResponseBody, l.Title,
		l.Url, time.Now())
	if e != nil {
		log.Printf("Unable to log to DB: %s, DATA: %s", e, l)
	}
}

func NewAccessLogger(c *config.Config) (AccessLogger, error) {
	switch c.Logs.Provider {
	case "db":
		return newMysqlLogger(c)
	case "console":
		return &consoleLogger{}, nil
	}
	return nil, fmt.Errorf("unsupported log provider: %s", c.Logs.Provider)
}
