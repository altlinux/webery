package keywords

import (
	"encoding/json"

	"gopkg.in/mgo.v2/bson"
)

type Keyword struct {
	Key   string `json:"key"`
	Group string `json:"group"`
}

type Keywords map[string]string

func NewKeywords(arr ...Keyword) Keywords {
	kwds := make(Keywords)

	for _, a := range arr {
		kwds[a.Group] = a.Key
	}

	return kwds
}

func (k Keywords) GetBSON() (interface{}, error) {
	return k.Keywords(), nil
}

func (k *Keywords) SetBSON(raw bson.Raw) error {
	var v []Keyword
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	if v == nil || len(v) == 0 {
		return nil
	}

	*k = NewKeywords(v...)
	return nil
}

func (k Keywords) MarshalJSON() (out []byte, err error) {
	if len(k) > 0 {
		out, err = json.Marshal(k.Keywords())
	} else {
		var null *int
		out, err = json.Marshal(null)
	}
	return
}

func (k *Keywords) UnmarshalJSON(data []byte) error {
	var v []Keyword
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil || len(v) == 0 {
		return nil
	}

	*k = NewKeywords(v...)
	return nil
}

func (k Keywords) Groups() []string {
	var groups []string

	for grp := range k {
		groups = append(groups, grp)
	}

	return groups
}

func (k Keywords) Keywords() []Keyword {
	var keywords []Keyword

	for grp, key := range k {
		keywords = append(keywords, Keyword{
			Key:   key,
			Group: grp,
		})
	}
	return keywords
}
