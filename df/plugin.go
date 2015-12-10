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
	_ "os"
	"strings"
	_ "time"

	"github.com/intelsdi-x/snap/control/plugin"
	"github.com/intelsdi-x/snap/control/plugin/cpolicy"
)

const (
	// Name of plugin
	Name = "df"
	// Version of plugin
	Version = 1
	// Type of plugin
	Type = plugin.CollectorPluginType
)

const (
	KIND_AVAILABLE = iota
	KIND_USED
	KIND_PERCENTAGE
)

type MetricKind int

var namespacePrefix = []string{"intel", "disk"}
var stringToKind = map[string]MetricKind{"available_space": KIND_AVAILABLE,
	"used_space": KIND_USED, "percentage_used": KIND_PERCENTAGE}
var kindToString = []string{"available_space", "used_space", "percentage_used"}
var true_false = []bool{true, false}

func makeNamespace(path string, kind MetricKind, aggregated, inode bool) []string {
	ns := []string{}
	ns = append(ns, namespacePrefix...)
	ns = append(ns, strings.Split(path, "/")...)
	ns = append(ns, "space")
	metric := ""
	if aggregated {
		metric = "aggregated_"
	}
	metric += kindToString[kind]
	if inode {
		metric += "_inodes"
	} else {
		metric += "_kB"
	}
	ns = append(ns, metric)
	return ns
}

func parseNamespace(ns []string) (path string, kind MetricKind, aggregated, inode bool) {
	woPrefix := ns[len(namespacePrefix):]
	n := len(woPrefix)
	path = strings.Join(woPrefix[:n-2], "/")
	trimmed := strings.TrimPrefix(woPrefix[n-1], "aggregated_")
	trimmed = strings.TrimSuffix(trimmed, "_kB")
	trimmed = strings.TrimSuffix(trimmed, "_inodes")
	kind = stringToKind[trimmed]
	aggregated = strings.HasPrefix(woPrefix[n-1], "aggregated_")
	inode = strings.HasSuffix(woPrefix[n-1], "_inodes")
	return
}

func collect() (kbRes, inodeRes DfOutput, err error) {
	outKB, outINode, err := runDf()
	if err != nil {
		return
	}

	kbRes, err = parse(outKB)
	if err != nil {
		err = fmt.Errorf("parsing kB data failed: %v", err)
		return
	}

	inodeRes, err = parse(outINode)
	if err != nil {
		err = fmt.Errorf("parsing inode data failed: %v", err)
		return
	}

	return
}

type DfCollector struct {
}

/*
func (p *DfCollector) CollectMetrics(mts []plugin.PluginMetricType) ([]plugin.PluginMetricType, error) {

}*/

func (p *DfCollector) GetMetricTypes(_ plugin.PluginConfigType) ([]plugin.PluginMetricType, error) {
	kbData, inodeData, err := collect()
	if err != nil {
		return nil, err
	}
	mts := []plugin.PluginMetricType{}
	data_all := map[bool]*DfOutput{false: &kbData, true: &inodeData}
	for inodeUsed, data := range data_all {

		fmt.Println(">>>>>>>>>>>>>> ", data.ByMountPoint())
		for key, info := range *data {
			for kind, _ := range kindToString {
				mt := plugin.PluginMetricType{Namespace_: makeNamespace(key.Path, MetricKind(kind), !info.Leaf, inodeUsed)}
				fmt.Printf("%v\t\t%v\n", info.MountPoint, mt.Namespace_)
				mts = append(mts, mt)
			}

		}
	}

	return mts, nil
}

// GetConfigPolicy
func (p *DfCollector) GetConfigPolicy() (*cpolicy.ConfigPolicy, error) {
	c := cpolicy.New()
	return c, nil
}

// Creates new instance of plugin and returns pointer to initialized object.
func NewDfCollector() *DfCollector {
	return &DfCollector{}
}

// Returns plugin's metadata
func Meta() *plugin.PluginMeta {
	return plugin.NewPluginMeta(Name, Version, Type, []string{plugin.SnapGOBContentType}, []string{plugin.SnapGOBContentType})
}
