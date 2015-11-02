// +build unit

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
	"testing"
	"encoding/json"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/opencontainers/runc/libcontainer/cgroups"
)

var _ cgroups.BlkioStatEntry

func TestMap2NamespaceNoComposition(t *testing.T){
	Convey("Given json-tagged struct", t, func() {

		obj := struct{
			Foo int `json:"foo"`
			Bar string `json:"bar"`
		}{
			99,
			"str",
		}

		Convey("and marshaled-unmarshaled map", func() {

			var jmap map[string]interface{}
			jsondata, _ := json.Marshal(obj)
			_ = json.Unmarshal(jsondata, &jmap)

			Convey("When map is converted to namespace", func() {

				ns := []string{}
				tools := MyTools{}
				tools.Map2Namespace(jmap, "root", &ns)

				Convey("Then namespace contains two entries", func() {
					So(len(ns), ShouldEqual, 2)
				})

				Convey("Then namespace elements are correctly set", func() {
					So(ns, ShouldContain, "root/foo")
					So(ns, ShouldContain, "root/bar")
				})

			})
		})

	})
}

func TestMap2NamespaceSimpleComposition(t *testing.T) {

	Convey("Given json-tagged struct with simple composition", t , func() {
		obj := struct{
			Foo struct{
					Bar string `json:"bar"`
					Baz string `json:"baz"`
				} 		`json:"foo"`
			Qaz int 	`json:"qaz"`
		}{
			struct{Bar string `json:"bar"`
				   Baz string `json:"baz"`}{"for_bar", "for_baz"},
			99,
		}

		Convey("and marshaled-umarshaled map", func() {
			var jmap map[string]interface{}
			jsondata, _ := json.Marshal(obj)
			_ = json.Unmarshal(jsondata, &jmap)

			Convey("When map is converted to namespace", func() {
				ns := []string{}
				tools := MyTools{}
				tools.Map2Namespace(jmap, "root", &ns)

				Convey("Then namespace contains three entries", func() {
					So(len(ns), ShouldEqual, 3)
				})

				Convey("Then namespaces elements are correctly set", func() {
					So(ns, ShouldContain, "root/foo/bar")
					So(ns, ShouldContain, "root/foo/baz")
					So(ns, ShouldContain, "root/qaz")
				})
			})
		})

	})
}

func TestMap2NamespaceSpecialCaseSlice(t *testing.T){

	Convey("Given json-tagged struct with slice of len = 3", t, func() {
		obj := struct{
			Foo []string 	`json:"foo"`
			Bar int			`json:"bar"`
		}{
			[]string{"11", "22", "33"},
			99,
		}

		Convey("and marshaled-unmarshaled map", func() {
			var jmap map[string]interface{}
			jsondata, _ := json.Marshal(obj)
			_ = json.Unmarshal(jsondata, &jmap)

			Convey("When map is converted to namespace", func() {
				ns := []string{}
				tools := MyTools{}
				tools.Map2Namespace(jmap, "root", &ns)

				Convey("Then namespace contains four entries", func() {
					So(len(ns), ShouldEqual, 4)
				})

				Convey("Then namespace elements are correctly set", func() {
					So(ns, ShouldContain, "root/foo/0")
					So(ns, ShouldContain, "root/foo/1")
					So(ns, ShouldContain, "root/foo/2")
					So(ns, ShouldContain, "root/bar")
				})
			})
		})
	})
}
/*
func TestGetValueByNamespaceNoComposition(t *testing.T){
	Convey("Given json-tagged struct with no composition", t, func() {
		obj := struct{
			Foo string 	`json:"foo"`
			Bar int		`json:"bar"`
		}{
			"foo_val",
			43,
		}

		Convey("When value by foo tag is requested", func() {
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, []string{"foo"})
			Convey("Then correct foo value is returned", func() {
				So(val.(string), ShouldEqual, "foo_val")
			})
		})

		Convey("When value by bar tag is requested", func() {
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, []string{"bar"})
			Convey("Then correct bar value is returned", func() {
				So(val.(int), ShouldEqual, 43)
			})
		})

	})
}

func TestGetValueByNamespaceSimpleComposition(t *testing.T) {

	Convey("Given json-tagged composite struct", t, func() {
		obj := struct{
			Foo struct{
				Bar int 	`json:"bar"`
				Baz string	`json:"baz"`
				}			`json:"foo"`
			Qaz string 		`json:"qaz"`
		}{
			struct{
				Bar int		`json:"bar"`
				Baz	string	`json:"baz"`}{43, "baz_val"},
			"qaz_val",
		}

		Convey("When value by foo/bar namespace is requested", func() {
			ns := []string{"foo", "bar"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar value is returned", func() {
				So(val.(int), ShouldEqual, 43)
			})
		})

		Convey("When value by foo/baz namespace is requested", func() {
			ns := []string{"foo", "baz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/baz value is returned", func() {
				So(val.(string), ShouldEqual, "baz_val")
			})
		})

		Convey("When value by qaz namespace is requested", func() {
			ns := []string{"qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct qaz value is returned", func() {
				So(val.(string), ShouldEqual, "qaz_val")
			})
		})
	})
}

func TestGetValueByNamespaceSliceCompositionInt(t *testing.T) {

	Convey("Given json-tagged composite struct with slice", t, func() {
		obj := struct{
			Foo struct{
				Bar []uint64 	`json:"bar"`
				Baz string		`json:"baz"`
				}				`json:"foo"`
			Qaz string 			`json:"qaz"`
		}{
			struct{
				Bar []uint64		`json:"bar"`
				Baz	string	`json:"baz"`}{[]uint64{43, 1}, "baz_val"},
			"qaz_val",
		}

		Convey("When values by foo/bar/* namespace is requested", func() {
			ns := []string{"foo", "bar", "0"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/0 value is returned", func() {
				So(val.(uint64), ShouldEqual, 43)
			})

			ns = []string{"foo", "bar", "1"}
			val = tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/1 value is returned", func() {
				So(val.(uint64), ShouldEqual, 1)
			})
		})

		Convey("When value by foo/baz namespace is requested", func() {
			ns := []string{"foo", "baz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/baz value is returned", func() {
				So(val.(string), ShouldEqual, "baz_val")
			})
		})

		Convey("When value by qaz namespace is requested", func() {
			ns := []string{"qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct qaz value is returned", func() {
				So(val.(string), ShouldEqual, "qaz_val")
			})
		})
	})
}

func TestGetValueByNamespaceSliceCompositionBlkio(t *testing.T) {

	Convey("Given json-tagged composite struct with slice", t, func() {
		obj := struct{
			Foo struct{
				Bar []cgroups.BlkioStatEntry 	`json:"bar"`
				Baz string						`json:"baz"`
				}								`json:"foo"`
			Qaz string 							`json:"qaz"`
		}{
			struct{
				Bar []cgroups.BlkioStatEntry	`json:"bar"`
				Baz	string	`json:"baz"`}{
					[]cgroups.BlkioStatEntry{cgroups.BlkioStatEntry{Major: 1, Minor: 2, Op: "op", Value: 4}},
					"baz_val"},
				"qaz_val",
		}

		Convey("When values by foo/bar/0/* namespace is requested", func() {
			ns := []string{"foo", "bar", "0", "major"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/0/major value is returned", func() {
				So(val.(uint64), ShouldEqual, 1)
			})

			ns = []string{"foo", "bar", "0", "minor"}
			val = tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/0/minor value is returned", func() {
				So(val.(uint64), ShouldEqual, 2)
			})

			ns = []string{"foo", "bar", "0", "op"}
			val = tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/0/op value is returned", func() {
				So(val.(string), ShouldEqual, "op")
			})

			ns = []string{"foo", "bar", "0", "value"}
			val = tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/0/vaue value is returned", func() {
				So(val.(uint64), ShouldEqual, 4)
			})
		})

		Convey("When value by foo/baz namespace is requested", func() {
			ns := []string{"foo", "baz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/baz value is returned", func() {
				So(val.(string), ShouldEqual, "baz_val")
			})
		})

		Convey("When value by qaz namespace is requested", func() {
			ns := []string{"qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct qaz value is returned", func() {
				So(val.(string), ShouldEqual, "qaz_val")
			})
		})
	})
}

func TestGetValueByNamespaceSliceUnsupported(t *testing.T) {

	Convey("Given json-tagged composite struct with slice", t, func() {
		obj := struct{
			Foo struct{
				Bar []string 	`json:"bar"`
				Baz string		`json:"baz"`
				}				`json:"foo"`
			Qaz string 			`json:"qaz"`
		}{
			struct{
				Bar []string	`json:"bar"`
				Baz	string	`json:"baz"`}{
					[]string{"unsupported"},
					"baz_val"},
				"qaz_val",
		}

		Convey("When values by unsupported foo/bar/0 namespace is requested", func() {
			ns := []string{"foo", "bar", "0"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then nil value is returned", func() {
				So(val, ShouldBeNil)
			})
		})
	})
}

func TestGetValueByNamespaceMemoryMapComposition(t *testing.T) {

	Convey("Given json-tagged composite struct with map", t, func() {
		obj := struct{
			Foo struct{
				Bar map[string]uint64 	`json:"bar"`
				Baz string				`json:"baz"`
				}						`json:"foo"`
			Qaz string 					`json:"qaz"`
		}{
			struct{
				Bar map[string]uint64	`json:"bar"`
				Baz	string	`json:"baz"`}{map[string]uint64{"tar": 43, "far": 1}, "baz_val"},
			"qaz_val",
		}

		Convey("When values by foo/bar/* namespace is requested", func() {
			ns := []string{"foo", "bar", "tar"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/tar value is returned", func() {
				So(val.(uint64), ShouldEqual, 43)
			})

			ns = []string{"foo", "bar", "far"}
			val = tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/far value is returned", func() {
				So(val.(uint64), ShouldEqual, 1)
			})
		})

		Convey("When value by foo/baz namespace is requested", func() {
			ns := []string{"foo", "baz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/baz value is returned", func() {
				So(val.(string), ShouldEqual, "baz_val")
			})
		})

		Convey("When value by qaz namespace is requested", func() {
			ns := []string{"qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct qaz value is returned", func() {
				So(val.(string), ShouldEqual, "qaz_val")
			})
		})
	})
}

func TestGetValueByNamespaceHugetlbMapComposition(t *testing.T) {

	Convey("Given json-tagged composite struct with map", t, func() {
		obj := struct{
			Foo struct{
				Bar map[string]cgroups.HugetlbStats 	`json:"bar"`
				Baz string								`json:"baz"`
				}										`json:"foo"`
			Qaz string 									`json:"qaz"`
		}{
			struct{
				Bar map[string]cgroups.HugetlbStats		`json:"bar"`
				Baz	string	`json:"baz"`}{
					map[string]cgroups.HugetlbStats{
						"2MB": cgroups.HugetlbStats{Usage: 43, MaxUsage: 99, Failcnt: 7}},
					"baz_val"},
				"qaz_val",
		}

		Convey("When values by foo/bar/2MB/usage namespace is requested", func() {
			ns := []string{"foo", "bar", "2MB", "usage"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/2MB/usage value is returned", func() {
				So(val.(uint64), ShouldEqual, 43)
			})
		})

		Convey("When values by foo/bar/2MB/max_usage namespace is requested", func() {
			ns := []string{"foo", "bar", "2MB", "max_usage"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/2MB/max_usage value is returned", func() {
				So(val.(uint64), ShouldEqual, 99)
			})
		})

		Convey("When values by foo/bar/2MB/failcnt namespace is requested", func() {
			ns := []string{"foo", "bar", "2MB", "failcnt"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/bar/2MB/failcnt value is returned", func() {
				So(val.(uint64), ShouldEqual, 7)
			})
		})

		Convey("When value by foo/baz namespace is requested", func() {
			ns := []string{"foo", "baz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct foo/baz value is returned", func() {
				So(val.(string), ShouldEqual, "baz_val")
			})
		})

		Convey("When value by qaz namespace is requested", func() {
			ns := []string{"qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then correct qaz value is returned", func() {
				So(val.(string), ShouldEqual, "qaz_val")
			})
		})
	})
}

func TestGetValueByNamespaceUnsupportedMapComposition(t *testing.T) {

	Convey("Given json-tagged composite struct with unsupported map type", t, func() {
		obj := struct{
			Foo struct{
				Bar map[string]string 	`json:"bar"`
				Baz string				`json:"baz"`
				}						`json:"foo"`
			Qaz string 					`json:"qaz"`
		}{
			struct{
				Bar map[string]string	`json:"bar"`
				Baz	string	`json:"baz"`}{
					map[string]string{
						"2MB": "value"},
					"baz_val"},
				"qaz_val",
		}

		Convey("When values by foo/bar/2MB namespace is requested", func() {
			ns := []string{"foo", "bar", "2MB"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)
			Convey("Then nil value is returned", func() {
				So(val, ShouldBeNil)
			})
		})
	})
}

func TestGetValuesByTagWrongNamespace(t *testing.T){
	Convey("Given json-tagged composite struct", t, func() {
		obj := struct{
			Foo struct{
				Bar uint64 		`json:"bar"`
				Baz string		`json:"baz"`
				}				`json:"foo"`
			Qaz string 			`json:"qaz"`
		}{
			struct{
				Bar uint64	`json:"bar"`
				Baz	string	`json:"baz"`}{43, "baz_val"},
			"qaz_val",
		}

		Convey("When value for wrong namespace is requestes", func() {
			ns := []string{"foo", "qaz"}
			tools := MyTools{}
			val := tools.GetValueByNamespace(obj, ns)

			Convey("Then there's no value to return", func() {
				So(val, ShouldBeNil)
			})
		})
	})
}

*/