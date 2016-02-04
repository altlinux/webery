package jsontype

import (
	"encoding/json"
	"strings"

	"gopkg.in/mgo.v2/bson"
)

type BaseString struct {
	value string
	ok    bool
}

func NewBaseString(v string) *BaseString {
	return &BaseString{
		value: v,
		ok:    true,
	}
}

func (o BaseString) GetBSON() (interface{}, error) {
	if !o.ok {
		return nil, nil
	}
	return o.value, nil
}

func (o *BaseString) SetBSON(raw bson.Raw) (error) {
	var v *string
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o *BaseString) MarshalJSON() (out []byte, err error) {
	if o.ok {
		out, err = json.Marshal(o.value)
	} else {
		var null *int
		out, err = json.Marshal(null)
	}
	return
}

func (o *BaseString) UnmarshalJSON(data []byte) error {
	var v *string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v != nil {
		o.value = *v
		o.ok = true
	}
	return nil
}

func (o BaseString) String() string {
	if o.ok {
		return o.value
	}
	return "<nil>"
}

func (o BaseString) Get() (string, bool) {
	return o.value, o.ok
}

func (o *BaseString) Set(v string) {
	o.value = v
	o.ok = true
}

type LowerString struct {
	BaseString
}

func NewLowerString(v string) *LowerString {
	o := &LowerString{}
	o.Set(strings.ToLower(v))
	return o
}

func (o *LowerString) UnmarshalJSON(data []byte) error {
	var v *string
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v != nil {
		o.value = strings.ToLower(*v)
		o.ok = true
	}
	return nil
}

func (o *LowerString) Set(v string) {
	o.value = strings.ToLower(v)
	o.ok = true
}
