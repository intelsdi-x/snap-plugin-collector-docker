// +build linux

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2015 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tools

import (
	"fmt"
	"strings"
	"reflect"
	"strconv"
	"path/filepath"

	"github.com/oleiade/reflections"
	_"github.com/docker/docker/vendor/src/github.com/opencontainers/runc/libcontainer/cgroups"
)

type ToolsInterface interface {
	Map2Namespace(stats map[string]interface{}, current string, out *[]string)
	GetValueByField(object interface{}, fields []string) interface{}
	GetValueByNamespace(object interface{}, ns []string) interface{}
}

type MyTools struct {}

func (t *MyTools) Map2Namespace(stats map[string]interface{}, current string, namespaces *[]string) {
	for key, val := range stats {
		// handling special cases
		switch reflect.TypeOf(val).Kind() {
		case reflect.Map:
			// modify current path and go deeper, do not populate output
			c := filepath.Join(current, key)
			t.Map2Namespace(val.(map[string]interface{}), c, namespaces)
		case reflect.Slice:
			val := reflect.ValueOf(val)
			for i := 0; i < val.Len(); i++ {
				// modify current path for each slice element
				c := filepath.Join(current, key, strconv.Itoa(i))
				if reflect.TypeOf(val.Index(i).Interface()).Kind() == reflect.Map {
					// go deeper
					t.Map2Namespace(val.Index(i).Interface().(map[string]interface{}), c, namespaces)
				} else {
					// leaf reached
					*namespaces = append(*namespaces, c)
				}
			}
		// max depth reached for current
		default:
			// modify current path
			c := filepath.Join(current, key)
			// create new output entry
			*namespaces = append(*namespaces, c)
		}
	}
}

func (t *MyTools) GetValueByField(object interface{}, fields []string) interface{} {
	// current level of struct composition
	field := fields[0]
	// reflect value for current composition by name
	val, err := reflections.GetField(object, field)
	if err != nil {
		fmt.Printf("Value for %s not found\n", field)
		return nil
	}
	// if it's the deepest level of struct composition return it's value
	if len(fields) == 1 {
		return val
	}
	// or go deeper
	return t.GetValueByField(val, fields[1:])
}

func (t *MyTools) GetValueByNamespace(object interface{}, ns []string) interface{} {
	// current level of namespace
	current := ns[0]
	fields, err := reflections.Fields(object)
	if err != nil {
		fmt.Printf("Could not return fields for object{%v}\n", object)
		return nil
	}

	for _, field := range fields {
		tag, err := reflections.GetFieldTag(object, field, "json")
		if err != nil {
			fmt.Printf("Could not find tag for field{%s}\n", field)
			return nil
		}
		// remove omitempty from tag
		tag = strings.Replace(tag, ",omitempty", "", -1)
		if tag == current {
			val, _ := reflections.GetField(object, field)

			// handling of special cases for slice and map
			switch reflect.TypeOf(val).Kind() {
			case reflect.Slice:
				idx, _ := strconv.Atoi(ns[1])
				val := reflect.ValueOf(val)
				if val.Index(idx).Kind() == reflect.Struct {
					return t.GetValueByNamespace(val.Index(idx).Interface(), ns[2:])
				} else {
					return val.Index(idx)
				}
			case reflect.Map:
				key := ns[1]
				// try uint64 map (memory_stats case)
				if vi, ok := val.(map[string]uint64); ok {
					return vi[key]
				}
				// try with hugetlb map (hugetlb_stats case)
				val := reflect.ValueOf(val)
				kval := reflect.ValueOf(key)
				if reflect.TypeOf(val.MapIndex(kval).Interface()).Kind() == reflect.Struct {
					return t.GetValueByNamespace(val.MapIndex(kval).Interface(), ns[2:])
				}
			default:
				// last ns, return value found
				if len(ns) == 1 {
					return val
				} else {
				// or go deeper
					return t.GetValueByNamespace(val, ns[1:])
				}
			}
		}
	}
	return nil
}
