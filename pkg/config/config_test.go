package config

import (
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestParse(t *testing.T) {
	cfgStr := `
[global]
address = "0.0.0.0:8080"
logfile = "/dev/stdout"
pidfile = "/tmp/webery.pid"
MaxConns = 1000000
GoMaxProcs = 0

[content]
path = "/scm/webery/data"

[builder]
TaskStates = "new"
TaskStates = "awaiting"

SubTaskStates = "active"
SubTaskStates = "cancelled"

SubTaskTypes = "srpm"
SubTaskTypes = "delete"

Repos = "4.0"
Repos = "4.1"

Arches = "i586"
Arches = "x86_64"

[logging]
level = "debug"
`

	cfg := &Config{}
	if err := cfg.LoadString(cfgStr); err != nil {
		t.Errorf("unexpected parse error: %v", err)
	}

	if cfg.Global.Address != "0.0.0.0:8080" {
		t.Errorf("field global.address has wrong value: %+v", cfg.Global.Address)
	}

	if cfg.Global.MaxConns != 1000000 {
		t.Errorf("field global.maxconns has wrong value: %+v", cfg.Global.MaxConns)
	}

	if cfg.Logging.Level.Level != logrus.DebugLevel {
		t.Errorf("field logging.level has wrong value: %+v", cfg.Logging.Level)
	}
}
