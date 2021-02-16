package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// APP_DB_HOST
var dbHost string

// APP_DB_MAX_IDLE_CONNECTIONS
var dbMaxIdleConnections string

// APP_DB_MAX_OPEN_CONNECTIONS
var dbMaxOpenConnections string

// APP_DB_PASS
var dbPass string

// APP_DB_PORT
var dbPort string

// APP_DB_USER
var dbUser string

// Config fields correspond to config file keys less the prefix
type Config struct {
	dbHost                  string // APP_DB_HOST
	dbMaxIdleConnections    string // APP_DB_MAX_IDLE_CONNECTIONS
	dbMaxOpenConnections    string // APP_DB_MAX_OPEN_CONNECTIONS
	dbPass                  string // APP_DB_PASS
	dbPort                  string // APP_DB_PORT
	dbUser                  string // APP_DB_USER
}

// DbHost is APP_DB_HOST
func (c *Config) DbHost() string {
	return c.dbHost
}

// DbMaxIdleConnections is APP_DB_MAX_IDLE_CONNECTIONS
func (c *Config) DbMaxIdleConnections() string {
	return c.dbMaxIdleConnections
}

// DbMaxOpenConnections is APP_DB_MAX_OPEN_CONNECTIONS
func (c *Config) DbMaxOpenConnections() string {
	return c.dbMaxOpenConnections
}

// DbPass is APP_DB_PASS
func (c *Config) DbPass() string {
	return c.dbPass
}

// DbPort is APP_DB_PORT
func (c *Config) DbPort() string {
	return c.dbPort
}

// DbUser is APP_DB_USER
func (c *Config) DbUser() string {
	return c.dbUser
}

// SetDbHost overrides the value of dbHost
func (c *Config) SetDbHost(v string) {
	c.dbHost = v
}

// SetDbMaxIdleConnections overrides the value of dbMaxIdleConnections
func (c *Config) SetDbMaxIdleConnections(v string) {
	c.dbMaxIdleConnections = v
}

// SetDbMaxOpenConnections overrides the value of dbMaxOpenConnections
func (c *Config) SetDbMaxOpenConnections(v string) {
	c.dbMaxOpenConnections = v
}

// SetDbPass overrides the value of dbPass
func (c *Config) SetDbPass(v string) {
	c.dbPass = v
}

// SetDbPort overrides the value of dbPort
func (c *Config) SetDbPort(v string) {
	c.dbPort = v
}

// SetDbUser overrides the value of dbUser
func (c *Config) SetDbUser(v string) {
	c.dbUser = v
}

// New creates an instance of Config.
// Build with ldflags to set the package vars.
// Env overrides package vars.
// Fields correspond to the config file keys less the prefix.
// The config file must have a flat structure
func New() *Config {
	conf := &Config{}
	SetVars(conf)
	SetEnv(conf)
	return conf
}

// SetVars sets non-empty package vars on Config
func SetVars(conf *Config) {

	if dbHost != "" {
		conf.dbHost = dbHost
	}

	if dbMaxIdleConnections != "" {
		conf.dbMaxIdleConnections = dbMaxIdleConnections
	}

	if dbMaxOpenConnections != "" {
		conf.dbMaxOpenConnections = dbMaxOpenConnections
	}

	if dbPass != "" {
		conf.dbPass = dbPass
	}

	if dbPort != "" {
		conf.dbPort = dbPort
	}

	if dbUser != "" {
		conf.dbUser = dbUser
	}
}

// SetEnv sets non-empty env vars on Config
func SetEnv(conf *Config) {
	var v string

	v = os.Getenv("APP_DB_HOST")
	if v != "" {
		conf.dbHost = v
	}

	v = os.Getenv("APP_DB_MAX_IDLE_CONNECTIONS")
	if v != "" {
		conf.dbMaxIdleConnections = v
	}

	v = os.Getenv("APP_DB_MAX_OPEN_CONNECTIONS")
	if v != "" {
		conf.dbMaxOpenConnections = v
	}

	v = os.Getenv("APP_DB_PASS")
	if v != "" {
		conf.dbPass = v
	}

	v = os.Getenv("APP_DB_PORT")
	if v != "" {
		conf.dbPort = v
	}

	v = os.Getenv("APP_DB_USER")
	if v != "" {
		conf.dbUser = v
	}

}

// LoadFile sets the env from file and returns a new instance of Config
func LoadFile(mode string) (conf *Config, err error) {
	appDir := os.Getenv("APP_DIR")
	p := fmt.Sprintf("%v/config.%v.json", appDir, mode)
	b, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	configMap := make(map[string]string)
	err = json.Unmarshal(b, &configMap)
	if err != nil {
		return nil, err
	}
	for key, val := range configMap {
		_ = os.Setenv(key, val)
	}
	return New(), nil
}
