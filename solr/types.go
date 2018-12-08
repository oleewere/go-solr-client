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
	"gopkg.in/jcmturner/gokrb5.v4/client"
	"net/http"
	"net/url"
)

// KerberosConfig holds kerberos related configurations
type KerberosConfig struct {
	keytab         string
	principal      string
	realm          string
	krb5confPath   string
	kerberosClient *client.Client
}

// BasicAuthConfig hold authentication credentials
type BasicAuthConfig struct {
	username string
	password string
}

// SecurityConfig holds security related configurations
type SecurityConfig struct {
	kerberosEnabled  bool
	basicAuthEnabled bool
	kerberosConfig   *KerberosConfig
	basicAuthConfig  *BasicAuthConfig
}

// TLSConfig holds TLS related configurations
type TLSConfig struct {
	cert    string
	enabled bool
}

// SolrConfig holds Solr related configurations
type SolrConfig struct {
	Url                   string
	Collection            string
	SecurityConfig        *SecurityConfig
	SolrUrlContext        string
	TlsConfig             TLSConfig
	Insecure              bool
	ConnectTimeoutSeconds int
}

// SolrClient represents a Solr connection that is used to communicate with Solr HTTP endpoints
type SolrClient struct {
	solrConfig *SolrConfig
	httpClient *http.Client
}

// SolrResponseData represents Solr response data that contains the response itself and the response header as well
type SolrResponseData struct {
	ResponseHeader SolrResponseHeader `json:"responseHeader"`
	Response       SolrResponse       `json:"response"`
}

// SolrDocument represents a Solr document (document map)
type SolrDocument map[string]interface{}

// SolrDocuments holds array of Solr documents
type SolrDocuments []SolrDocument

// SolrResponse represents a Solr HTTP response
type SolrResponse struct {
	NumFound int32          `json:"numFound,omitempty"`
	Start    int32          `json:"start,omitempty"`
	MaxScore float32        `json:"maxScore,omitempty"`
	Docs     []SolrDocument `json:"docs,omitempty"`
}

// SolrResponseHeader represents Solr request headers from Solr HTTP response
type SolrResponseHeader struct {
	Status int32             `json:"status,omitempty"`
	QTime  int32             `json:"QTime,omitempty"`
	Params map[string]string `json:"params,omitempty"`
}

// SSHConfig holds SSH related configs that is used by the data generator (to gather keytabs if kerberos is enabled)
type SSHConfig struct {
	Enabled          bool
	Username         string
	PrivateKeyPath   string
	DownloadLocation string
	RemoteKrb5Conf   string
	RemoteKeytab     string
	Hostname         string
}

// SolrQuery represents a solr query object
type SolrQuery struct {
	params *url.Values
}
