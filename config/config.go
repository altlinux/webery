/*
* Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
*
* This file is covered by the GNU General Public License,
* which should be included with webery as the file COPYING.
*/

package config

import (
	"github.com/BurntSushi/toml"
	log "github.com/Sirupsen/logrus"

	"strings"
)

type CfgLogLevel struct {
	log.Level
}

func (d *CfgLogLevel) UnmarshalText(data []byte) (err error) {
	d.Level, err = log.ParseLevel(strings.ToLower(string(data)))
	return
}

type ConfigGlobal struct {
	Address    string
	Logfile    string
	Pidfile    string
	GoMaxProcs int
	MaxConns   int64
}

type ConfigContent struct {
	Path string
}

type ConfigLogging struct {
	Level            CfgLogLevel
	DisableColors    bool
	DisableTimestamp bool
	FullTimestamp    bool
	DisableSorting   bool
}

type ConfigMongo struct {
	Hosts     []string
	Direct    bool
	Database  string
	User      string
	Password  string
	PoolLimit int
}

type ConfigBuilder struct {
	TaskStates    []string
	SubTaskStates []string
	SubTaskTypes  []string
	Repos         []string
	Arches        []string
}

type Config struct {
	Global  ConfigGlobal
	Content ConfigContent
	Logging ConfigLogging
	Mongo   ConfigMongo
	Builder ConfigBuilder
}

// SetDefaults applies default values to config structure.
func (c *Config) SetDefaults() {
	c.Global.GoMaxProcs = 0
	c.Global.Logfile = "/var/log/webery.log"
	c.Global.Pidfile = "/run/webery.pid"

	c.Logging.Level.Level = log.InfoLevel
	c.Logging.DisableColors = true
	c.Logging.DisableTimestamp = false
	c.Logging.FullTimestamp = true
	c.Logging.DisableSorting = true

	c.Mongo.Hosts = []string{"localhost:27017"}
	c.Mongo.Database = "girar"
	c.Mongo.Direct = false
	c.Mongo.PoolLimit = 128
}

var cfgGlobal *Config

func GetConfig() *Config {
	if cfgGlobal == nil {
		panic("Config not initialized")
	}
	return cfgGlobal
}

func NewConfig(filename string) (*Config, error) {
	cfg := &Config{}
	cfg.SetDefaults()

	_, err := toml.DecodeFile(filename, cfg)

	if err == nil {
		cfgGlobal = cfg
	}

	return cfg, err
}
