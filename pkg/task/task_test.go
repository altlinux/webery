package task

import (
	"testing"

//	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/config"
	storage "github.com/altlinux/webery/pkg/db/dbtest"
)

func TestParse(t *testing.T) {
	cfg := &config.Config{}
	dbi := storage.NewSession(cfg.Mongo)

	_, err := GetTask(dbi, 149239)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
