package logging

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	"shawnma.com/clarity/config"
)

type consoleLogger struct{}

func (c *consoleLogger) Log(l *HttpLog) {
	log.Printf("ACCESS: [%s | %s][%s | %d | %s][%d | %s | %d | %s] %s",
		l.RemoteAddr, l.Method,
		l.RequestContentType, l.RequestLength, l.RequestBody,
		l.ResponseCode, l.ResponseContentType, l.ResponseLength, l.Title,
		l.Url)
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
