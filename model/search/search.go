/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package search

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/storage"
)

type Keyword struct {
	Key   string
	Group string
}

type Keywords struct {
	keywords []Keyword
	groups   map[string]bool
}

func NewKeywords() *Keywords {
	return &Keywords{
		keywords: make([]Keyword, 0),
		groups:   make(map[string]bool, 0),
	}
}

func (k *Keywords) Length() int {
	return len(k.keywords)
}

func (k *Keywords) Append(group string, key string) {
	if len(key) > 0 {
		k.keywords = append(k.keywords, Keyword{key, group})
	}
	k.groups[group] = true
}

func (k *Keywords) Groups() []string {
	groups := make([]string, 0)

	for grp, _ := range k.groups {
		groups = append(groups, grp)
	}

	return groups
}

func (k *Keywords) Keywords() []Keyword {
	return k.keywords
}

func FindKey(st *storage.MongoStorage, collname string, prefix string, limit int) *mgo.Iter {
	query := st.Coll(collname).Find(bson.M{
		"search.key": prefix,
	})

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.Iter()
}

func FindPrefix(st *storage.MongoStorage, collname string, prefix string, limit int) *mgo.Iter {
	query := st.Coll(collname).Find(bson.M{
		"search.key": bson.M{
			"$regex": "^" + prefix,
		},
	})

	if limit > 0 {
		query = query.Limit(limit)
	}

	return query.Iter()
}
