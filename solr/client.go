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
	"time"
	"net/http"
	"fmt"
	"bytes"
	"io/ioutil"
	"gopkg.in/jcmturner/gokrb5.v4/keytab"
	"gopkg.in/jcmturner/gokrb5.v4/client"
	"gopkg.in/jcmturner/gokrb5.v4/config"
	"log"
	"encoding/json"
	"net/url"
)

func NewSolrClient(url string, collection string, solrConfig *SolrConfig) (*SolrClient, error) {
	httpClient := &http.Client{
		Transport: &http.Transport{
			Proxy:                 http.ProxyFromEnvironment,
			MaxIdleConns:          100,
			IdleConnTimeout:       30 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	solrConfig.Collection = collection
	solrConfig.Url = url
	solrConfig.SolrUrlContext = "/solr"

	securityConfig := solrConfig.SecurityConfig

	if securityConfig.kerberosConfig != nil && len(securityConfig.kerberosConfig.keytab) != 0 {
		securityConfig.kerberosEnabled = true
	}

	if securityConfig.kerberosEnabled {
		keytabPath := securityConfig.kerberosConfig.keytab
		principalName := securityConfig.kerberosConfig.principal
		krb5confPath := securityConfig.kerberosConfig.krb5confPath
		realm := securityConfig.kerberosConfig.realm

		kt, err := keytab.Load(keytabPath)

		if (err != nil) {
			log.Fatal(err)
		}
		cfg, err := config.Load(krb5confPath)
		cl := client.NewClientWithKeytab(principalName, realm, kt)
		cl.WithConfig(cfg)
		errLogin := cl.Login()

		if (errLogin != nil) {
			log.Fatal(errLogin)
		}
		solrConfig.SecurityConfig.kerberosConfig.kerberosClient = &cl
	}
	solrClient := SolrClient{httpClient: httpClient, solrConfig: solrConfig}
	return &solrClient, nil
}

// Set initial security config on start
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

// Add WWW-Authenticate header (SPNEGO) in case of kerberos is enabled
func AddNegotiateHeader(request *http.Request , solrConfig *SolrConfig) {
	if solrConfig.SecurityConfig.kerberosConfig != nil && solrConfig.SecurityConfig.kerberosEnabled {
		spn := ""
		kcl := solrConfig.SecurityConfig.kerberosConfig.kerberosClient
		kcl.SetSPNEGOHeader(request, spn)
	}
}

// Get Solr collection url with url context (if exists) and url suffix
// e.g.: url - https://myurl:8886, context: /solr, suffix: /update/json/docs = https://myurl:8886/solr/update/json/docs
func GetSolrCollectionUri(solrConfig *SolrConfig, uriSuffix string) string {
	var uriPrefix = solrConfig.Url
	if len(solrConfig.SolrUrlContext) != 0 {
		uriPrefix = uriPrefix + "" + solrConfig.SolrUrlContext
	}
	uri := fmt.Sprintf("%s/%s/%s", uriPrefix, solrConfig.Collection, uriSuffix)
	return uri
}

func (solrClient* SolrClient) Update(docs interface{}, parameters *url.Values, commit bool) error {
	httpClient := solrClient.httpClient
	uri := GetSolrCollectionUri(solrClient.solrConfig, "update/json/docs")
	var buf bytes.Buffer
	if docs != nil {
		encoder := json.NewEncoder(&buf)
		if err := encoder.Encode(docs); err != nil {
			return err
		}
	}
	request, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		log.Fatal(err)
	}

	request.Header.Add("Content-Type", "application/json")
	AddNegotiateHeader(request, solrClient.solrConfig)

	response, err := httpClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	return nil
}

func (solrClient *SolrClient) Query(parameters *url.Values) SolrResponseData {
	httpClient := solrClient.httpClient
	uri := GetSolrCollectionUri(solrClient.solrConfig, "select")
	var buf bytes.Buffer
	request, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		log.Fatal(err)
	}

	if parameters == nil {
		parameters = &url.Values{}
		parameters.Add("q", "*:*")
	}

	request.URL.RawQuery = parameters.Encode()
	request.Header.Add("Content-Type", "application/json")

	AddNegotiateHeader(request, solrClient.solrConfig)
	log.Print("Query: ", uri)

	response, err := httpClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if (err != nil) {
		log.Fatal(err)
	}

	var solrResponse SolrResponseData
	json_err := json.Unmarshal(bodyBytes, &solrResponse)
	if json_err != nil {
		log.Fatal(err)
	}
	return solrResponse
}
