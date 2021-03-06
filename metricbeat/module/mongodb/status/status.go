// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package status

import (
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"
	"github.com/elastic/beats/metricbeat/mb"
	"github.com/elastic/beats/metricbeat/module/mongodb"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

/*
TODOs:
	* add metricset for "locks" data
	* add a metricset for "metrics" data
*/

var debugf = logp.MakeDebug("mongodb.status")

func init() {
	mb.Registry.MustAddMetricSet("mongodb", "status", New,
		mb.WithHostParser(mongodb.ParseURL),
		mb.DefaultMetricSet(),
	)
}

// MetricSet type defines all fields of the MetricSet
// As a minimum it must inherit the mb.BaseMetricSet fields, but can be extended with
// additional entries. These variables can be used to persist data or configuration between
// multiple fetch calls.
type MetricSet struct {
	mb.BaseMetricSet
	dialInfo *mgo.DialInfo
}

// New creates a new instance of the MetricSet
// Part of new is also setting up the configuration by processing additional
// configuration entries if needed.
func New(base mb.BaseMetricSet) (mb.MetricSet, error) {
	dialInfo, err := mgo.ParseURL(base.HostData().URI)
	if err != nil {
		return nil, err
	}
	dialInfo.Timeout = base.Module().Config().Timeout

	return &MetricSet{
		BaseMetricSet: base,
		dialInfo:      dialInfo,
	}, nil
}

// Fetch methods implements the data gathering and data conversion to the right format
// It returns the event which is then forward to the output. In case of an error, a
// descriptive error must be returned.
func (m *MetricSet) Fetch() (common.MapStr, error) {

	// instantiate direct connections to each of the configured Mongo hosts
	mongoSession, err := mongodb.NewDirectSession(m.dialInfo)
	if err != nil {
		return nil, err
	}
	defer mongoSession.Close()

	result := map[string]interface{}{}
	if err := mongoSession.DB("admin").Run(bson.D{{Name: "serverStatus", Value: 1}}, &result); err != nil {
		return nil, err
	}

	data, _ := schema.Apply(result)
	return data, nil
}
