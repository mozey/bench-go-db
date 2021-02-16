package main

import (
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"strings"
	"time"
)

var AppDir = "/Users/mozey/go/src/github.com/mozey/insert-bench"

// Setup configures logging,
// initializes services,
// re-creates the db tables,
// and returns the services handler
func Setup() (s *Services, err error) {
	SetupLogger()

	err = os.Setenv("APP_DIR", AppDir)
	if err != nil {
		return s, errors.WithStack(err)
	}

	conf, err := LoadConfig()
	if err != nil {
		return s, errors.WithStack(err)
	}

	s = NewServices(conf)

	db, err := s.DB()
	if err != nil {
		return s, errors.WithStack(err)
	}

	sql := `drop table if exists insertbench;`
	_, err = db.Exec(sql)
	if err != nil {
		return s, errors.WithStack(err)
	}

	sql = `
create table insertbench (
	product varchar(191) collate utf8mb4_unicode_ci not null,
	sku varchar(191) collate utf8mb4_unicode_ci not null,
	attr varchar(191) collate utf8mb4_unicode_ci not null,
	value text collate utf8mb4_unicode_ci not null,
	created datetime not null,
    modified datetime not null,
	primary key (sku),
	unique key (sku, attr)
) engine=innodb default charset=utf8mb4 collate=utf8mb4_unicode_ci;
`
	_, err = db.Exec(sql)
	if err != nil {
		return s, errors.WithStack(err)
	}

	return s, errors.WithStack(err)
}

type Product struct {
	DatabaseRow
	Product string `json:"product" db:"product" gorm:"column:product"`
	SKU     string `json:"sku" db:"sku" gorm:"column:sku"`
	Attr    string `json:"attr" db:"attr" gorm:"column:attr"`
	Value   string `json:"value" db:"value" gorm:"column:value"`
}

type Products []Product

// TableName overrides default gorm table name
func (p Product) TableName() string {
	return "insertbench"
}

var conf *Config

// LoadConfig loads dev config from file once and returns a copy.
// To override variables per test use the setter functions on conf
func LoadConfig() (*Config, error) {
	var err error
	// If config is not set
	if conf == nil {
		// Load config
		conf, err = LoadFile("dev")
		if err != nil {
			return conf, err
		}
	}
	// Copy the struct not the pointer!
	confCopy := *conf
	return &confCopy, nil
}

var productLabels = []string{"RunningShoe", "T-Shirt", "Backpack"}
var skuLabelMap = map[string]string{
	"RunningShoe": "shoe",
	"T-Shirt":     "shirt",
	"Backpack":    "pack",
}

// GenProducts generates product data.
// Use skuPrefix to avoid primary key conflicts
func GenProducts(groupCount int, skusPerGroup int, skuPrefix interface{}) (prod Products) {
	prods := make(Products, groupCount*skusPerGroup)
	i := 0
	for g := 0; g < groupCount; g++ {
		group := productLabels[g%3]
		groupFmt := fmt.Sprintf("%v-%v-%v", skuPrefix, productLabels[g%3], i)
		sku := skuLabelMap[group]
		for s := 0; s < skusPerGroup; s++ {
			skuFmt := fmt.Sprintf("%v-%v-%v", skuPrefix, sku, i)
			prods[i] = Product{
				Product: groupFmt,
				SKU:   skuFmt,
				Attr:   "x",
				Value: "y",
			}
			i++
		}
	}
	return prods
}

// SqlxSingle generates an insert statement for each product and executes it.
// It is very similar to running the raw query with db.Exec.
// Returns an error or the generated SQL
func SqlxSingle(db *sqlx.DB, prods Products) (sql string, err error) {
	sql = `
insert into insertbench (product, sku, attr, value, created, modified)
values (:product, :sku, :attr, :value, :created, :modified)
`
	for _, prod := range prods {
		prod.SetDates()
		_, err = db.NamedExec(sql, prod)
		if err != nil {
			return sql, errors.WithStack(err)
		}
	}
	return sql, nil
}

// SetValues repeats the values named param for rowCount
func SetValues(queryIn string, rowCount int) (query string, err error) {
	values := strings.Repeat(",(:values)\n", rowCount)
	values = strings.Replace(values, ",", "", 1)
	values = strings.TrimSpace(values)
	return strings.Replace(queryIn, "(:values)", values, 1), nil
}

// SqlxValues uses helper functions to generate one insert
// for all the products and executes it.
// Returns an error or the generated SQL
func SqlxValues(db *sqlx.DB, prods Products) (sql string, err error) {
	sql = `
insert into insertbench (product, sku, attr, value, created, modified)
values (:values)
on duplicate key update
product=values(product), value=values(value), modified=values(modified)
`
	sql, err = SetValues(sql, len(prods))
	if err != nil {
		return sql, errors.WithStack(err)
	}

	// Create values
	colsCount := 6
	now := time.Now().UTC().Format(DateFormat)
	values := make([]interface{}, len(prods))
	for i, prod := range prods {
		row := make([]interface{}, colsCount)
		row[0] = prod.Product
		row[1] = prod.SKU
		row[2] = prod.Attr
		row[3] = prod.Value
		row[4] = now
		row[5] = now
		values[i] = row
	}

	sql, _, err = sqlx.Named(sql, map[string]interface{}{
		"values": values,
	})
	if err != nil {
		return sql, errors.WithStack(err)
	}

	sql, args, err := sqlx.In(sql, values...)
	if err != nil {
		return sql, errors.WithStack(err)
	}

	_, err = db.Exec(sql, args...)
	if err != nil {
		return sql, errors.WithStack(err)
	}

	return sql, nil
}

// GormSingle is the same as SqlxSingle, but using gorm
func GormSingle(db *gorm.DB, prods Products) (sql string, err error) {
	for _, prod := range prods {
		prod.SetDates()
		// "Error handling in GORM is different than idiomatic Go
		// code because of its chainable API"
		// http://gorm.io/docs/error_handling.html
		err = db.Create(prod).Error
		if err != nil {
			return sql, errors.WithStack(err)
		}
	}
	return sql, nil
}

// GormValues is the same as SqlxValues, but using gorm
func GormValues(db *gorm.DB, prods Products) (sql string, err error) {
	// TODO
	return sql, nil
}

// JSONDumpIndent can be used to pretty print a struct representing json data
func JSONDumpIndent(i interface{}) string {
	indentString := "    "
	s, _ := json.MarshalIndent(i, "", indentString)
	return string(s)
}

// SetupLogger sets up logging using zerolog.
//
// Wrap new errors with WithStack
//		errors.WithStack(fmt.Errorf("foo"))
//
func SetupLogger() {
	zerolog.TimeFieldFormat = time.RFC3339
	zerolog.TimestampFieldName = "created"
	zerolog.ErrorFieldName = "message"
	zerolog.ErrorStackMarshaler = MarshalStack
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	log.Logger = log.Output(ConsoleWriter{
		Out:           os.Stderr,
		NoColor:       false,
		TimeFormat:    "2006-01-02 15:04:05",
		MarshalIndent: true,
	})
}


func main() {
	s, err := Setup()
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Msg("")
		os.Exit(1)
	}

	//db, err := s.DB()
	//if err != nil {
	//	log.Error().Stack().Err(errors.WithStack(err)).Msg("")
	//	os.Exit(1)
	//}

	db, err := s.GormDB()
	if err != nil {
		log.Error().Stack().Err(errors.WithStack(err)).Msg("")
		os.Exit(1)
	}

	var sql string

	//prods := GenProducts(10, 2, 0)
	////fmt.Println(JSONDumpIndent(prods))
	//sql, err = SqlxSingle(db, prods)
	//if err != nil {
	//	log.Error().Stack().Err(err).Msg("")
	//	os.Exit(1)
	//}
	//fmt.Println(sql)
	//
	//prods = GenProducts(10, 2, 0)
	//sql, err = SqlxValues(db, prods)
	//if err != nil {
	//	log.Error().Stack().Err(err).Msg("")
	//	os.Exit(1)
	//}
	//fmt.Println(sql)

	prods := GenProducts(10, 2, 0)
	//fmt.Println(JSONDumpIndent(prods))
	sql, err = GormSingle(db, prods)
	if err != nil {
		log.Error().Stack().Err(err).Msg("")
		os.Exit(1)
	}
	fmt.Println(sql)

	//prods = GenProducts(10, 2, 0)
	//sql, err = GormValues(db, prods)
	//if err != nil {
	//	log.Error().Stack().Err(err).Msg("")
	//	os.Exit(1)
	//}
	//fmt.Println(sql)

}
