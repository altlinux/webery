package keywords

type Keyword struct {
	Key   string `json:"key"`
	Group string `json:"group"`
}

type Keywords struct {
	keywords []Keyword
	groups   map[string]*Keyword
}

func NewKeywords() *Keywords {
	return &Keywords{
		keywords: make([]Keyword, 0),
		groups:   make(map[string]*Keyword, 0),
	}
}

func (k Keywords) Length() int {
	return len(k.keywords)
}

func (k *Keywords) Append(group string, key string) {
	if len(key) == 0 {
		return
	}

	if _, ok := k.groups[group]; !ok {
		k.keywords = append(k.keywords, Keyword{key, group})
		k.groups[group] = &k.keywords[len(k.keywords)-1]
	} else {
		k.groups[group].Key = key
	}
}

func (k Keywords) Groups() []string {
	groups := make([]string, 0)

	for grp, _ := range k.groups {
		groups = append(groups, grp)
	}

	return groups
}

func (k *Keywords) Keywords() []Keyword {
	return k.keywords
}

func (k Keywords) Group(name string) (*Keyword, bool) {
	v, ok := k.groups[name]
	return v, ok
}
