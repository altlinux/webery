package main

import (
	"fmt"
	"flag"

	log "github.com/Sirupsen/logrus"

	"github.com/altlinux/webery/pkg/config"
	"github.com/altlinux/webery/pkg/db"
	"github.com/altlinux/webery/pkg/jsontype"
	kwd "github.com/altlinux/webery/pkg/keywords"
	storage "github.com/altlinux/webery/pkg/db/mongo"
)

var (
	configFile = flag.String("config", "", "Path to configuration file")
)

type oldTask struct {
	Search     kwd.Keywords         `json:"-"`
	Try        jsontype.Int64       `json:"try,omitempty"`
	Iter       jsontype.Int64       `json:"iter,omitempty"`
	ObjType    jsontype.BaseString  `json:"objtype,omitempty"`
	TimeCreate jsontype.Int64       `json:"timecreate,omitempty"`
	TaskID     jsontype.Int64       `json:"taskid,omitempty"`
	Owner      jsontype.LowerString `json:"owner,omitempty"`
	State      jsontype.LowerString `json:"state,omitempty"`
	Repo       jsontype.LowerString `json:"repo,omitempty"`
	Aborted    jsontype.LowerString `json:"aborted,omitempty"`
	Shared     jsontype.Bool        `json:"shared,omitempty"`
	Swift      jsontype.Bool        `json:"swift,omitempty"`
	TestOnly   jsontype.Bool        `json:"testonly,omitempty"`
}

func oldNew() *oldTask {
	t := &oldTask{}

	t.ObjType.Set("task")
	t.Search = kwd.NewKeywords()

	return t
}

func main() {
	flag.Parse()

	cfg, err := config.NewConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	dbi := storage.NewSession(cfg.Mongo)
	defer dbi.Close()

	col, err := dbi.Coll("tasks")
	if err != nil {
		log.Fatal(err)
	}

	iter := col.Find(db.QueryDoc{"events": db.QueryDoc{"$exists": false}}).Iter()
	for {
		o := make(db.QueryDoc)
		if !iter.Next(o) {
			break
		}

		//fmt.Printf("%d\n", o["taskid"])

		taskid, ok := o["taskid"].(int64)
		if !ok {
			fmt.Printf("%d: error taskid not found\n", o["taskid"])
			continue
		}

		try, ok := o["try"].(int64)
		if !ok {
			fmt.Printf("%d: error try not found\n", o["taskid"])
			continue
		}

		var events []string
		for i := int64(1); i <= try; i += 1 {
			events = append(events, fmt.Sprintf("%d.1", i))
		}
		o["events"] = events

		s := db.QueryDoc{
			"taskid": taskid,
		}
		d := db.QueryDoc{
			"$set": db.QueryDoc{"events": events},
		}

		if err := col.Update(s, d); err != nil {
			fmt.Printf("%d: mongo error: %v\n", o["taskid"], err)
			break
		}
	}
	fmt.Printf("OK\n")
}
