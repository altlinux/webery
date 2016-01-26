package dbutil

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/mgo.v2/bson"

	"github.com/altlinux/webery/pkg/keywords"
)

type fieldParams struct {
	Public    bool
	Keyword   bool
	Lowercase bool
}

func parseTag(tag string) *fieldParams {
	res := &fieldParams{}
	for _, word := range strings.Split(tag, ",") {
		switch word {
		case "private":
			res.Public = false
		case "public":
			res.Public = true
		case "keyword":
			res.Keyword = true
		case "lowercase":
			res.Lowercase = true
		}
	}
	return res
}

func UpdateBSON(val reflect.Value) ([]bson.M, error) {
	if val.Kind() == reflect.Interface && !val.IsNil() {
		elm := val.Elem()
		if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
			val = elm
		}
	}
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	kwds := keywords.NewKeywords()
	changeSet := bson.M{}

	for i := 0; i < val.NumField(); i++ {
		valueField := val.Field(i)
		typeField := val.Type().Field(i)

		tags := parseTag(typeField.Tag.Get("webery"))

		if !tags.Public {
			continue
		}

		if valueField.Kind() == reflect.Interface && !valueField.IsNil() {
			elm := valueField.Elem()
			if elm.Kind() == reflect.Ptr && !elm.IsNil() && elm.Elem().Kind() == reflect.Ptr {
				valueField = elm
			}
		}

		if valueField.Kind() == reflect.Ptr {
			if valueField.IsNil() {
				continue
			}
			valueField = valueField.Elem()
		}

		name := typeField.Tag.Get("json")

		if valueField.Kind() == reflect.String {
			var str string

			if tags.Lowercase {
				str = strings.ToLower(valueField.Interface().(string))
			} else {
				str = valueField.Interface().(string)
			}

			if tags.Keyword {
				kwds.Append(name, str)
			}
			changeSet[name] = str
		} else if valueField.Kind() == reflect.Int64 {
			if tags.Keyword {
				str := fmt.Sprintf("%d", valueField.Int())
				kwds.Append(name, str)
			}
			changeSet[name] = valueField.Interface()
		} else {
			changeSet[name] = valueField.Interface()
		}
	}

	res := make([]bson.M, 0)
	changeBSON := bson.M{}

	if kwds.Length() > 0 {
		res = append(res, bson.M{
			"$pull": bson.M{
				"search": bson.M{
					"group": bson.M{
						"$in": kwds.Groups(),
					},
				},
			},
		})

		changeBSON["$addToSet"] = bson.M{
			"search": bson.M{
				"$each": kwds.Keywords(),
			},
		}
	}

	if len(changeSet) > 0 {
		changeBSON["$set"] = changeSet
	}

	if len(changeBSON) > 0 {
		res = append(res, changeBSON)
	}

	return res, nil
}
