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

	if len(securityConfig.kerberosConfig.keytab) != 0 {
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
			log.Fatal(err)
		}
		solrConfig.SecurityConfig.kerberosConfig.kerberosClient = &cl
	}

	solrClient := SolrClient{httpClient: httpClient, solrConfig: solrConfig}

	return &solrClient, nil
}

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

func (solrClient *SolrClient) Query(queryString string) SolrResponseData {
	httpClient := solrClient.httpClient
	var uriPrefix = solrClient.solrConfig.Url

	if len(solrClient.solrConfig.SolrUrlContext) != 0 {
		uriPrefix = uriPrefix + "" + solrClient.solrConfig.SolrUrlContext
	}

	uri := fmt.Sprintf("%s/%s/select", uriPrefix, solrClient.solrConfig.Collection)
	var buf bytes.Buffer
	request, err := http.NewRequest("POST", uri, &buf)
	if err != nil {
		log.Fatal(err)
	}
	request.URL.RawQuery = queryString
	request.Header.Add("Content-Type", "application/json")

	if solrClient.solrConfig.SecurityConfig.kerberosConfig != nil && solrClient.solrConfig.SecurityConfig.kerberosEnabled {
		spn := ""
		kcl := solrClient.solrConfig.SecurityConfig.kerberosConfig.kerberosClient
		kcl.SetSPNEGOHeader(request, spn)
	}
	log.Print("query: ", uri)

	response, err := httpClient.Do(request)

	if err != nil {
		log.Fatal(err)
	}

	defer response.Body.Close()

	bodyBytes, err := ioutil.ReadAll(response.Body)

	if (err != nil) {
		log.Fatal(err)
	}
	body := string(bodyBytes)

	fmt.Println(body)
	var solrResponse SolrResponseData
	json_err := json.Unmarshal(bodyBytes, &solrResponse)
	if json_err != nil {
		log.Fatal(err)
	}
	return solrResponse
}
