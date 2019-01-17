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
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/jcmturner/gokrb5.v4/client"
	"gopkg.in/jcmturner/gokrb5.v4/config"
	"gopkg.in/jcmturner/gokrb5.v4/keytab"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

// NewSolrClient initialize a new Solr client based on configuration type
func NewSolrClient(solrConfig *SolrConfig) (*SolrClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}

	var securityConfig SecurityConfig
	if solrConfig.SecurityConfig == nil {
		solrConfig.SecurityConfig = new(SecurityConfig)
	}

	if securityConfig.kerberosConfig != nil && len(securityConfig.kerberosConfig.keytab) != 0 {
		securityConfig.kerberosEnabled = true
	}

	if securityConfig.kerberosEnabled {
		keytabPath := securityConfig.kerberosConfig.keytab
		principalName := securityConfig.kerberosConfig.principal
		krb5confPath := securityConfig.kerberosConfig.krb5confPath
		realm := securityConfig.kerberosConfig.realm

		kt, err := keytab.Load(keytabPath)

		if err != nil {
			log.Fatal(err)
		}
		cfg, err := config.Load(krb5confPath)
		if err != nil {
			log.Fatal(err)
		}
		cl := client.NewClientWithKeytab(principalName, realm, kt)
		cl.WithConfig(cfg)
		errLogin := cl.Login()

		if errLogin != nil {
			log.Fatal(errLogin)
		}
		solrConfig.SecurityConfig.kerberosConfig.kerberosClient = &cl
	}
	solrClient := SolrClient{httpClient: httpClient, solrConfig: solrConfig}
	return &solrClient, nil
}

// InitSecurityConfig set initial security config on start
func InitSecurityConfig(krb5Path string, keytabPath string, principal string, realm string) SecurityConfig {
	var securityConfig SecurityConfig
	if len(keytabPath) > 0 {
		kerberosConfig := KerberosConfig{krb5confPath: krb5Path, keytab: keytabPath, principal: principal, realm: realm}
		securityConfig = SecurityConfig{kerberosEnabled: true, kerberosConfig: &kerberosConfig}
	} else {
		securityConfig = SecurityConfig{}
	}
	return securityConfig
}

// AddBasicAuthHeader add Auth header with basic auth credentials
func AddBasicAuthHeader(request *http.Request, solrConfig *SolrConfig) {
	if solrConfig.SecurityConfig.basicAuthConfig != nil && solrConfig.SecurityConfig.basicAuthEnabled {
		request.SetBasicAuth(solrConfig.SecurityConfig.basicAuthConfig.username, solrConfig.SecurityConfig.basicAuthConfig.password)
	}
}

// AddNegotiateHeader add WWW-Authenticate header (SPNEGO) in case of kerberos is enabled
func AddNegotiateHeader(request *http.Request, solrConfig *SolrConfig) {
	if solrConfig.SecurityConfig.kerberosConfig != nil && solrConfig.SecurityConfig.kerberosEnabled {
		spn := ""
		kcl := solrConfig.SecurityConfig.kerberosConfig.kerberosClient
		kcl.SetSPNEGOHeader(request, spn)
	}
}

// GetSolrCollectionUri gather Solr collection url with url context (if exists) and url suffix
// e.g.: url - https://myurl:8886, context: /solr, suffix: /update/json/docs = https://myurl:8886/solr/update/json/docs
func GetSolrCollectionUri(solrConfig *SolrConfig, uriSuffix string) string {
	var uriPrefix = solrConfig.Url
	if len(solrConfig.SolrUrlContext) != 0 {
		uriPrefix = uriPrefix + "" + solrConfig.SolrUrlContext
	}
	uri := fmt.Sprintf("%s/%s/%s", uriPrefix, solrConfig.Collection, uriSuffix)
	return uri
}

// Update send documents to Solr
func (solrClient *SolrClient) Update(docs interface{}, parameters *url.Values, commit bool) (bool, *SolrResponseData, error) {
	httpClient := solrClient.httpClient
	uri := GetSolrCollectionUri(solrClient.solrConfig, "update")
	var buf bytes.Buffer
	if docs != nil {
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(docs); err != nil {
			return false, nil, err
		}
	}
	request, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		return false, nil, err
	}

	request.Header.Add("Content-Type", "application/json")

	AddBasicAuthHeader(request, solrClient.solrConfig)
	AddNegotiateHeader(request, solrClient.solrConfig)

	response, err := httpClient.Do(request)

	if err != nil {
		return false, nil, err
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)

	var solrResponse SolrResponseData
	jsonErr := json.Unmarshal(bodyBytes, &solrResponse)
	if jsonErr != nil {
		return false, nil, jsonErr
	}

	return true, &solrResponse, err
}

// Query get Solr data based on parameters
func (solrClient *SolrClient) Query(solrQuery *SolrQuery) (bool, *SolrResponseData, error) {
	httpClient := solrClient.httpClient
	uri := GetSolrCollectionUri(solrClient.solrConfig, "select")
	var buf bytes.Buffer
	request, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		return false, nil, err
	}

	if solrQuery == nil {
		solrQuery = CreateSolrQuery()
	}

	request.URL.RawQuery = solrQuery.Encode()
	request.Header.Add("Content-Type", "application/json")

	AddBasicAuthHeader(request, solrClient.solrConfig)
	AddNegotiateHeader(request, solrClient.solrConfig)

	log.Print("Query: ", uri)

	response, err := httpClient.Do(request)

	if err != nil {
		return false, nil, err
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return false, nil, err
	}

	var solrResponse SolrResponseData
	jsonErr := json.Unmarshal(bodyBytes, &solrResponse)
	if jsonErr != nil {
		return false, nil, jsonErr
	}
	return true, &solrResponse, nil
}
