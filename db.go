package main

import (
	"bytes"
	"fmt"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	// mysql driver must be imported
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"text/template"
	"time"
)

// MysqlDataSourceName connection string template
const MysqlDataSourceName = "{{.User}}:{{.Pass}}@({{.Host}}:{{.Port}})/{{.DB}}"

// ConnectionConfig for db
type ConnectionConfig struct {
	Host         string
	User         string
	Pass         string
	Port         string
	DB           string
	Timeout      time.Duration
	MaxOpenConns int
	MaxIdleConns int
}

// GetConnection returns a database connection, remember to close it
//
// "A DB instance is not a connection, but an abstraction representing a
// Database... It maintains a connection pool internally"
// Open vs Connect
// "open a DB and connect at the same time;
// for instance, in order to catch configuration issues during your
// initialization phase"
// http://jmoiron.github.io/sqlx/#connecting
//
// "...use DB.SetMaxOpenConns to set the maximum size of the pool
// ...set the maximum idle size with DB.SetMaxIdleConns"
// http://jmoiron.github.io/sqlx/#connectionPool
//
func GetConnection(c *ConnectionConfig) (db *sqlx.DB, err error) {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 2
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 2
	}

	tFn := template.Must(template.New("dsTemplate").Parse(MysqlDataSourceName))
	var dataSourceName bytes.Buffer
	err = tFn.Execute(&dataSourceName, c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Timeout if connection takes too long
	// https://gobyexample.com/timeouts
	if c.Timeout == 0 {
		c.Timeout = time.Second * 3
	}
	type dbMsg struct {
		DB  *sqlx.DB
		Err error
	}
	dbChan := make(chan dbMsg, 1)
	go func() {
		db, err = sqlx.Connect("mysql", dataSourceName.String())
		dbChan <- dbMsg{
			DB:  db,
			Err: errors.WithStack(err),
		}
	}()

	select {
	case res := <-dbChan:
		if res.Err != nil {
			return nil, res.Err
		}
		db = res.DB
		db.SetMaxOpenConns(c.MaxOpenConns)
		db.SetMaxIdleConns(c.MaxIdleConns)
		return db, nil
	case <-time.After(c.Timeout):
		return nil, errors.WithStack(fmt.Errorf("timeout connecting to db"))
	}
}

// GetGormConnection returns a database connection, remember to close it
//
// "A DB instance is not a connection, but an abstraction representing a
// Database... It maintains a connection pool internally"
// Open vs Connect
// "open a DB and connect at the same time;
// for instance, in order to catch configuration issues during your
// initialization phase"
// http://jmoiron.github.io/sqlx/#connecting
//
// "...use DB.SetMaxOpenConns to set the maximum size of the pool
// ...set the maximum idle size with DB.SetMaxIdleConns"
// http://jmoiron.github.io/sqlx/#connectionPool
//
func GetGormConnection(c *ConnectionConfig) (db *gorm.DB, err error) {
	if c.MaxOpenConns == 0 {
		c.MaxOpenConns = 2
	}
	if c.MaxIdleConns == 0 {
		c.MaxIdleConns = 2
	}

	tFn := template.Must(template.New("dsTemplate").Parse(MysqlDataSourceName))
	var dataSourceName bytes.Buffer
	err = tFn.Execute(&dataSourceName, c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// Timeout if connection takes too long
	// https://gobyexample.com/timeouts
	if c.Timeout == 0 {
		c.Timeout = time.Second * 3
	}
	type dbMsg struct {
		DB  *gorm.DB
		Err error
	}
	dbChan := make(chan dbMsg, 1)
	go func() {
		db, err = gorm.Open("mysql", dataSourceName.String())
		dbChan <- dbMsg{
			DB:  db,
			Err: errors.WithStack(err),
		}
	}()

	select {
	case res := <-dbChan:
		if res.Err != nil {
			return nil, res.Err
		}
		db = res.DB
		db.DB().SetMaxOpenConns(c.MaxOpenConns)
		db.DB().SetMaxIdleConns(c.MaxIdleConns)
		return db, nil
	case <-time.After(c.Timeout):
		return nil, errors.WithStack(fmt.Errorf("timeout connecting to db"))
	}
}

// DateFormat can be used to get MySQL formatted date times
// https://golang.org/pkg/time/#Time.Format
const DateFormat = "2006-01-02 15:04:05"

// DatabaseRow can be embedded in all row types
type DatabaseRow struct {
	Created  string `json:"created" db:"created" gorm:"column:created"`
	Modified string `json:"modified" db:"modified" gorm:"column:modified"`
}

// SetCreated date on database row.
// Pass in empty string to use current timestamp
func (r *DatabaseRow) SetCreated(now string) {
	if now == "" {
		now = time.Now().UTC().Format(DateFormat)
	}
	r.Created = now
}

// SetModified date on database row.
// Pass in empty string to use current timestamp
func (r *DatabaseRow) SetModified(now string) {
	if now == "" {
		now = time.Now().UTC().Format(DateFormat)
	}
	r.Modified = now
}

// SetDates set all dates on the database row
func (r *DatabaseRow) SetDates() {
	now := time.Now().UTC().Format(DateFormat)
	r.SetCreated(now)
	r.SetModified(now)
}
