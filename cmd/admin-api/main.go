/*
Copyright 2019 The Kunkka Authors.

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

package main

import (
	"fmt"
	"github.com/gostship/kunkka/cmd/admin-api/app"
	"math/rand"
	"os"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	apiCmd := app.GetApiCmd(os.Args[1:])

	if err := apiCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error:%v\n", err)
		os.Exit(-1)
	}
}
