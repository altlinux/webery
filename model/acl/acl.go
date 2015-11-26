/*
 * Copyright (C) 2015 Alexey Gladkov <gladkov.alexey@gmail.com>
 *
 * This file is covered by the GNU General Public License,
 * which should be included with webery as the file COPYING.
 */

package acl

import (
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/storage"
)

const collection_pkg = "acl_packages"
const collection_grp = "acl_groups"

type Member struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	Leader bool   `json:"leader"`
}

type ACL struct {
	Repo    string   `json:"repo"`
	Name    string   `json:"name"`
	Members []Member `json:"members"`
}

func ListPackagesByRepo(st *storage.MongoStorage, repo string) *mgo.Query {
	return st.Coll(collection_pkg).Find(bson.M{"repo": repo})
}

func ListGroupsByRepo(st *storage.MongoStorage, repo string) *mgo.Query {
	return st.Coll(collection_grp).Find(bson.M{"repo": repo})
}

func GetPackageACL(st *storage.MongoStorage, repo string, name string) (*ACL, error) {
	res := &ACL{}

	err := st.Coll(collection_pkg).
		Find(bson.M{
		"repo": repo,
		"name": name,
	}).
		One(res)

	return res, err
}

func GetGroupACL(st *storage.MongoStorage, repo string, name string) (*ACL, error) {
	res := &ACL{}

	err := st.Coll(collection_grp).
		Find(bson.M{
		"repo": repo,
		"name": name,
	}).
		One(res)

	return res, err
}
