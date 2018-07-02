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
	"net/http"
	"gopkg.in/jcmturner/gokrb5.v4/client"
)

type KerberosConfig struct {
	keytab         string
	principal      string
	realm          string
	krb5confPath   string
	kerberosClient *client.Client
}

type BasicAuthConfig struct {
	username string
	password string
}

type SecurityConfig struct {
	kerberosEnabled  bool
	basicAuthEnabled bool
	kerberosConfig   *KerberosConfig
	basicAuthConfig  *BasicAuthConfig
}

type TLSConfig struct {
	cert    string
	enabled bool
}

type SolrConfig struct {
	Url                   string
	Collection            string
	SecurityConfig        *SecurityConfig
	SolrUrlContext        string
	TlsConfig             TLSConfig
	Insecure              bool
	ConnectTimeoutSeconds int
}

type SolrClient struct {
	solrConfig *SolrConfig
	httpClient *http.Client
}
