package main

import (
	"github.com/jinzhu/gorm"
	"github.com/jmoiron/sqlx"
)

// Services to embed in internal handler
type Services struct {
	Config *Config
	db     *sqlx.DB
	gormDB *gorm.DB
}

// NewServices initialises services
func NewServices(conf *Config) (s *Services) {
	s = &Services{}
	s.Config = conf
	return s
}

// DB initialises a new client or returns existing
func (s *Services) DB() (*sqlx.DB, error) {
	var err error

	if s.db != nil {
		return s.db, nil
	}

	s.db, err = GetConnection(&ConnectionConfig{
		Host: s.Config.DbHost(),
		User: s.Config.DbUser(),
		Pass: s.Config.DbPass(),
		Port: s.Config.DbPort(),
		// Assuming DB name:= User
		DB: s.Config.DbUser(),
	})
	if err != nil {
		return s.db, err
	}

	return s.db, nil
}

/// GormDB initialises a new client or returns existing
func (s *Services) GormDB() (*gorm.DB, error) {
	var err error

	if s.gormDB != nil {
		return s.gormDB, nil
	}

	s.gormDB, err = GetGormConnection(&ConnectionConfig{
		Host: s.Config.DbHost(),
		User: s.Config.DbUser(),
		Pass: s.Config.DbPass(),
		Port: s.Config.DbPort(),
		// Assuming DB name:= User
		DB: s.Config.DbUser(),
	})
	if err != nil {
		return s.gormDB, err
	}

	return s.gormDB, nil
}

// Cleanup function must be called before the application exits
func (s *Services) Cleanup() {
	if s.db != nil {
		_ = s.db.Close()
	}
}
