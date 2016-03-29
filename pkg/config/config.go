/*
* Copyright (C) 2016 Alexey Gladkov <gladkov.alexey@gmail.com>
*
* This file is covered by the GNU General Public License,
* which should be included with webery as the file COPYING.
 */

package config

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/Sirupsen/logrus"
	"gopkg.in/gcfg.v1"
)

type cfgKeyConfig int

type CfgLogLevel struct {
	logrus.Level
}

func (d *CfgLogLevel) UnmarshalText(data []byte) (err error) {
	d.Level, err = logrus.ParseLevel(strings.ToLower(string(data)))
	return
}

type Global struct {
	Address    string
	Logfile    string
	Pidfile    string
	GoMaxProcs int
	MaxConns   int64
}

type Content struct {
	Path string
}

type Logging struct {
	Level            CfgLogLevel
	DisableColors    bool
	DisableTimestamp bool
	FullTimestamp    bool
	DisableSorting   bool
}

type Mongo struct {
	Hosts     []string
	Direct    bool
	Database  string
	User      string
	Password  string
	PoolLimit int
}

type Builder struct {
	TaskStates    []string
	SubTaskStates []string
	SubTaskTypes  []string
	Repos         []string
	Arches        []string
}

type Config struct {
	Global  Global
	Content Content
	Logging Logging
	Mongo   Mongo
	Builder Builder
}

// SetDefaults applies default values to config structure.
func (c *Config) SetDefaults() *Config {
	c.Global.GoMaxProcs = 0
	c.Global.Logfile = "/var/log/webery.log"
	c.Global.Pidfile = "/run/webery.pid"

	c.Logging.Level.Level = logrus.InfoLevel
	c.Logging.DisableColors = true
	c.Logging.DisableTimestamp = false
	c.Logging.FullTimestamp = true
	c.Logging.DisableSorting = false

	c.Mongo.Hosts = []string{"localhost:27017"}
	c.Mongo.Database = "girar"
	c.Mongo.Direct = false
	c.Mongo.PoolLimit = 128

	return c
}

func (c *Config) ParseString(str string) error {
	if err := gcfg.ReadStringInto(c, str); err != nil {
		return err
	}
	return nil
}

func (c *Config) LoadString(str string) error {
	return c.SetDefaults().ParseString(str)
}

func NewConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := cfg.LoadString(string(buf)); err != nil {
		return nil, err
	}

	return cfg, err
}
