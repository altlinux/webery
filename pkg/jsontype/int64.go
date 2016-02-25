package jsontype

import (
	"encoding/json"
	"fmt"

	"gopkg.in/mgo.v2/bson"
)

type Int64 struct {
	value int64
	ok    bool
}

func NewInt64(v int64) *Int64 {
	return &Int64{
		value: v,
		ok:    true,
	}
}

func (o Int64) GetBSON() (interface{}, error) {
	if !o.ok {
		return nil, nil
	}
	return o.value, nil
}

func (o *Int64) SetBSON(raw bson.Raw) error {
	var v *int64
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o Int64) MarshalJSON() (out []byte, err error) {
	if o.ok {
		out, err = json.Marshal(o.value)
	} else {
		var null *int
		out, err = json.Marshal(null)
	}
	return
}

func (o *Int64) UnmarshalJSON(data []byte) error {
	var v *int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o Int64) String() string {
	if o.ok {
		return fmt.Sprintf("%d", o.value)
	}
	return "<nil>"
}

func (o Int64) IsDefined() bool {
	return o.ok
}

func (o Int64) Get() (int64, bool) {
	return o.value, o.ok
}

func (o *Int64) Set(v int64) {
	o.value = v
	o.ok = true
}
