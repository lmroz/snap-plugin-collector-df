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

package df

import (
	"fmt"
	"os/exec"
)

var optionsKB = []string{"--no-sync", "-P", "-T"}
var optionsINode = []string{"--no-sync", "-P", "-T", "-i"}

func runDf() (kBOutput, iNodeOutput string, err error) {
	kBOutputB, erri := exec.Command("df", optionsKB...).Output()
	if erri != nil {
		err = fmt.Errorf("df run for kB output failed: %v", erri)
		return
	}
	kBOutput = string(kBOutputB)

	iNodeOutputB, erri := exec.Command("df", optionsINode...).Output()
	if erri != nil {
		err = fmt.Errorf("df run for inodes output failed: %v", erri)
		return
	}
	iNodeOutput = string(iNodeOutputB)

	err = nil

	return
}
