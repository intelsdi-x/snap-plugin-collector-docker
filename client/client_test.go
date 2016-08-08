/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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

package client

import (
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGetShortID(t *testing.T) {

	Convey("Get short docker container ID", t, func() {

		Convey("successful get short container ID (12 chars)", func() {
			mockContainerID := strings.Repeat("1", 60)

			shortID, err := GetShortID(mockContainerID)
			So(err, ShouldBeNil)
			So(len(shortID), ShouldEqual, 12)
		})

		Convey("return an error when container ID contains less that 12 chars", func() {
			mockContainerID := strings.Repeat("1", 11)

			shortID, err := GetShortID(mockContainerID)
			So(err, ShouldNotBeNil)
			So(shortID, ShouldBeBlank)
		})

	})
}
