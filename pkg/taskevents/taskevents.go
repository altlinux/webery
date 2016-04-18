package taskevents

import (
	"fmt"
	"encoding/json"
	"sort"

	"gopkg.in/mgo.v2/bson"
)

type EventList struct {
	value    sort.StringSlice
	ok       bool
	readonly bool
}

func NewEventList() *EventList {
	return &EventList{}
}

func (o EventList) GetBSON() (interface{}, error) {
	if !o.ok {
		return nil, nil
	}
	return o.value, nil
}

func (o *EventList) SetBSON(raw bson.Raw) error {
	if o.readonly {
		return nil
	}
	var v sort.StringSlice
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	o.value = v
	o.ok = true
	return nil
}

func (o EventList) MarshalJSON() (out []byte, err error) {
	if o.ok {
		out, err = json.Marshal(o.value)
	} else {
		var null *int
		out, err = json.Marshal(null)
	}
	return
}

func (o *EventList) UnmarshalJSON(data []byte) error {
	if o.readonly {
		return nil
	}
	var v sort.StringSlice
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.value = v
	o.ok = true
	return nil
}

func (o EventList) String() string {
	if o.ok {
		return fmt.Sprintf("%q", o.value)
	}
	return "<nil>"
}

func (o EventList) IsDefined() bool {
	return o.ok
}

func (o EventList) Get() (sort.StringSlice, bool) {
	return o.value, o.ok
}

func (o *EventList) Set(v []string) {
	if o.readonly {
		return
	}
	o.value = make(sort.StringSlice, len(v))
    copy(o.value, v)
	o.ok = true
}

func (o *EventList) Append(v string) {
	if o.readonly {
		return
	}
	for _, e := range o.value {
		if e == v {
			return
		}
	}
	o.value = append(o.value, v)
	o.value.Sort()
	o.ok = true
}

func (o *EventList) Readonly(v bool) {
	o.readonly = v
}
