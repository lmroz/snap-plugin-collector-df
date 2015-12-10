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
	"strconv"
	"strings"
	"unicode"
)

type Space struct {
	Used, Available uint64
	Percentage      float64
	Leaf            bool
	MountPoint      string
}

type OutputKey struct {
	Path, MountPoint string
}

type DfOutput map[OutputKey]Space

func (self DfOutput) ByMountPoint() map[string]Space {
	res := map[string]Space{}
	for key, val := range self {
		res[key.MountPoint] = val
	}

	return res
}

func (self Space) Add(s2 Space) Space {
	res := self
	res.Available += s2.Available
	res.Used += s2.Used

	res.Percentage = float64(res.Used) / float64(res.Available)

	return res
}

func (self Space) WithMountPoint(mp string) Space {
	res := self
	res.MountPoint = mp
	return res
}

const (
	COLUMN_PATH = iota
	COLUMNS_FS
	COLUMN_BLOCKSIZE
	COLUMN_USED
	COLUMN_AVAILABLE
	COLUMN_PERCENT
	COLUMN_MOUNT_POINT
	COLUMNS_LEN
)

func isNewline(c rune) bool {
	return c == '\n' || c == '\r'
}

func isDiskSplit(c rune) bool {
	return c == '/' || c == '-'
}

func splitOnDigit(s string) []string {
	res := []string{""}
	for i, c := range s {
		if i > 0 && unicode.IsDigit(rune(c)) && !unicode.IsDigit(rune(s[i-1])) {
			res = append(res, string(c))
		} else {
			res[len(res)-1] += string(c)
		}
	}
	return res
}

func partitionExtract(s []string) []string {
	if s[0] == "dev" {
		return append(s[1:len(s)-1], splitOnDigit(s[len(s)-1])...)
	}
	return s
}

func makeKey(s []string, MountPoint string) OutputKey {
	return OutputKey{Path: strings.Join(s, "/"), MountPoint: MountPoint}
}

func parse(output string) (DfOutput, error) {
	var err error
	res := DfOutput{}
	lines := strings.FieldsFunc(output, isNewline)
	for columnNr, line := range lines {
		columns := strings.Fields(line)

		if len(columns) < COLUMNS_LEN {
			return nil, fmt.Errorf("unexpected number of columns: got %d expected %d", len(columns), COLUMNS_LEN)
		}

		// check len also on first row
		if columnNr < 1 {
			continue
		}

		path := strings.FieldsFunc(columns[COLUMN_PATH], isDiskSplit)
		parsedPath := partitionExtract(path)

		leaf := Space{}
		leaf.Available, err = strconv.ParseUint(columns[COLUMN_AVAILABLE], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse available space column: got %v", columns[COLUMN_AVAILABLE])
		}
		leaf.Used, err = strconv.ParseUint(columns[COLUMN_USED], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse used space column: got %v", columns[COLUMN_USED])
		}

		MountPoint := columns[COLUMN_MOUNT_POINT]
		LeafMountPoit := MountPoint

		for i := len(parsedPath) - 1; i >= 0; i-- {
			key := makeKey(parsedPath[0:i+1], MountPoint)
			res[key] = res[key].Add(leaf).WithMountPoint(MountPoint)
			MountPoint = ""
			//tmpfs is special case
			//TODO: check if that makes sense
			/*
				if strings.Contains(columns[COLUMNS_FS], "tmpfs") {
					setAvailable := res[key]
					setAvailable.Available = leaf.Available
					res[key] = setAvailable
				}
			*/
		}

		setLeaf := res[makeKey(parsedPath, LeafMountPoit)]
		setLeaf.Leaf = true
		res[makeKey(parsedPath, LeafMountPoit)] = setLeaf

	}

	return res, nil
}
