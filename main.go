package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

type Space struct {
	Used, Available uint64
	Leaf            bool
}

func (self Space) Add(s2 Space) Space {
	res := self
	res.Available += s2.Available
	res.Used += s2.Used

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

var x = `
System plików           Typ        1024-bl     użyte  dostępne pojemność zamont. na
/dev/mapper/fedora-root ext4      51475068  12286688  36550556       26% /
devtmpfs                devtmpfs   8110252         0   8110252        0% /dev
tmpfs                   tmpfs      8121252     68408   8052844        1% /dev/shm
tmpfs                   tmpfs      8121252      1372   8119880        1% /run
tmpfs                   tmpfs      8121252         0   8121252        0% /sys/fs/cgroup
tmpfs                   tmpfs      8121252   2115744   6005508       27% /tmp
/dev/sda1               ext4        487652    163058    294898       36% /boot
/dev/mapper/fedora-home ext4     401182688 260885336 119895384       69% /home
tmpfs                   tmpfs      1624252         4   1624248        1% /run/user/989
tmpfs                   tmpfs      1624252        44   1624208        1% /run/user/1000
`

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

func makeKey(s []string) string {
	return strings.Join(s, "/")
}

func parse(output string) (map[string]Space, error) {
	var err error
	res := map[string]Space{}
	lines := strings.FieldsFunc(x, isNewline)
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
		if len(columns) < COLUMNS_LEN {
			return nil, fmt.Errorf("cannot parse available space column: got %v", columns[COLUMN_AVAILABLE])
		}
		leaf.Used, err = strconv.ParseUint(columns[COLUMN_USED], 10, 64)
		if len(columns) < COLUMNS_LEN {
			return nil, fmt.Errorf("cannot parse used space column: got %v", columns[COLUMN_USED])
		}

		for i := len(parsedPath) - 1; i >= 0; i-- {
			key := makeKey(parsedPath[0 : i+1])
			res[key] = res[key].Add(leaf)
			//tmpfs is special case
			if strings.Contains(columns[COLUMNS_FS], "tmpfs") {
				setAvailable := res[key]
				setAvailable.Available = leaf.Available
				res[key] = setAvailable
			}
		}

		setLeaf := res[makeKey(parsedPath)]
		setLeaf.Leaf = true
		res[makeKey(parsedPath)] = setLeaf

	}

	return res, nil
}

type MetricKind int

const (
	KIND_AVAILABLE = iota
	KIND_USED
	KIND_PERCENTAGE
)

var namespacePrefix = []string{"intel", "disk"}
var suffixToKind = map[string]MetricKind{"available_space": KIND_AVAILABLE}

func makeNamespace(path string, kind MetricKind, aggregated bool) []string {

	var metric string
	switch kind {
	case KIND_AVAILABLE:

	}
}

func parseNamespace(ns []string) (path string, kind MetricKind, aggregated bool) {

}

func main() {
	fmt.Printf("%+v\n", parse(x))
}
