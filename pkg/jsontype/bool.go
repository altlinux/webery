package jsontype

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
)

type Bool struct {
	value    bool
	ok       bool
	readonly bool
}

func NewBool(v bool) *Bool {
	return &Bool{
		value:    v,
		ok:       true,
		readonly: false,
	}
}

func (o Bool) GetBSON() (interface{}, error) {
	if !o.ok {
		return nil, nil
	}
	return o.value, nil
}

func (o *Bool) SetBSON(raw bson.Raw) error {
	if o.readonly {
		return nil
	}
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
	if o.readonly {
		return nil
	}
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

func (o Bool) IsDefined() bool {
	return o.ok
}

func (o Bool) Get() (value bool, ok bool) {
	return o.value, o.ok
}

func (o *Bool) Set(v bool) {
	if o.readonly {
		return
	}
	o.value = v
	o.ok = true
}

func (o *Bool) Readonly(v bool) {
	o.readonly = v
}
