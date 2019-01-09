// Copyright 2018 Oliver Szabo
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package solr

import (
	"fmt"
	"net/url"
	"strings"
)

// CreateSolrQuery will create a new Solr query with empty queries
func CreateSolrQuery() *SolrQuery {
	q := new(SolrQuery)
	q.params = &url.Values{}
	return q
}

// AddParam add query parameter for Solr query
func (q *SolrQuery) AddParam(key, value string) {
	q.params.Add(key, value)
}

// SetParam set query parameter for Solr query
func (q *SolrQuery) SetParam(key, value string) {
	q.params.Set(key, value)
}

// Encode transform Solr query parameters to string
func (q *SolrQuery) Encode() string {
	return q.params.Encode()
}

// Query sets query string
func (q *SolrQuery) Query(query string) {
	q.AddParam("q", query)
}

// FilterQuery add filter query string
func (q *SolrQuery) FilterQuery(filterQuery string) {
	q.AddParam("fq", filterQuery)
}

// FacetQuery sets facet query string
func (q *SolrQuery) FacetQuery(query string) {
	q.SetParam("facet", "true")
	q.AddParam("fq", query)
}

// AddFacet add facet field
func (q *SolrQuery) AddFacet(field string) {
	q.SetParam("facet", "true")
	q.AddParam("facet.field", field)
}

// AddFields adding fields to Solr query
func (q *SolrQuery) AddFields(fields []string) {
	if len(fields) > 0 {
		q.AddParam("fl", strings.Join(fields, ","))
	}
}

// AddPivotFields adding pivot fields to Solr query
func (q *SolrQuery) AddPivotFields(pivotFields []string) {
	q.SetParam("facet", "true")
	if len(pivotFields) > 0 {
		q.AddParam("facet.pivot.field", strings.Join(pivotFields, ","))
	}
}

// Start sets start parameter for Solr query
func (q *SolrQuery) Start(start int) {
	q.SetParam("start", fmt.Sprintf("%d", start))
}

// Rows sets rows parameter for Solr query
func (q *SolrQuery) Rows(rows int) {
	q.SetParam("rows", fmt.Sprintf("%d", rows))
}

// Sort sets sort for Solr query
func (q *SolrQuery) Sort(sort string) {
	q.SetParam("sort", sort)
}
