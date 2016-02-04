package jsontype

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
)

type Bool struct {
	value bool
	ok    bool
}

func NewBool(v bool) *Bool {
	return &Bool{
		value: v,
		ok:    true,
	}
}

func (o Bool) GetBSON() (interface{}, error) {
	if !o.ok {
		return nil, nil
	}
	return o.value, nil
}

func (o *Bool) SetBSON(raw bson.Raw) (error) {
	var v *bool
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o Bool) MarshalJSON() (out []byte, err error) {
	if o.ok {
		out, err = json.Marshal(o.value)
	} else {
		var null *int
		out, err = json.Marshal(null)
	}
	return
}

func (o *Bool) UnmarshalJSON(data []byte) error {
	var v *bool
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o Bool) String() string {
	if o.ok {
		if o.value {
			return "true"
		}
		return "false"
	}
	return "<nil>"
}

func (o Bool) Get() (value bool, ok bool) {
	return o.value, o.ok
}

func (o *Bool) Set(v bool) {
	o.value = v
	o.ok = true
}
