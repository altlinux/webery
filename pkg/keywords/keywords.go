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

func (k Keywords) SetBSON(raw bson.Raw) (error) {
	var v []Keyword
	if err := raw.Unmarshal(&v); err != nil {
		return err
	}
	if v == nil {
		return nil
	}

	k = NewKeywords()
	for _, a := range v {
		k[a.Group] = a.Key
	}

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

func (k Keywords) UnmarshalJSON(data []byte) error {
	v :=  make([]Keyword, 0)
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if v == nil {
		return nil
	}

	k = NewKeywords()
	for _, a := range v {
		k[a.Group] = a.Key
	}
	return nil
}

func (k Keywords) Groups() []string {
	groups := make([]string, 0)

	for grp, _ := range k {
		groups = append(groups, grp)
	}

	return groups
}

func (k Keywords) Keywords() []Keyword {
	keywords := make([]Keyword, 0)

	for grp, key := range k {
		keywords = append(keywords, Keyword{
			Key:   key,
			Group: grp,
		})
	}
	return keywords
}
