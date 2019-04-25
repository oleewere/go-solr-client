## Go Solr Client

[![Build Status](https://travis-ci.org/oleewere/go-solr-client.svg?branch=master)](https://travis-ci.org/oleewere/go-solr-client)
[![Go Report Card](https://goreportcard.com/badge/github.com/oleewere/go-solr-client)](https://goreportcard.com/report/github.com/oleewere/go-solr-client)
![license](http://img.shields.io/badge/license-Apache%20v2-blue.svg)

### Install

```bash
go get -u github.com/oleewere/solr-client
```

### Usage

```go
import (
	"github.com/oleewere/go-solr-client/solr"
)

func main() {
	securityConfig := SecurityConfig{}
	if kerberosEnabled {
		securityConfig = InitSecurityConfig(krb5Path, keytabPath, principal, realm)
	}
	
	// ...
	
	solrUrl := "http://localhost:8886"
	solrCollection := "mycollection"
	solrConext := "/solr"
	tlsConfig := TLSConfig{}
	
	// ...
	
	solrConfig := SolrConfig{solrUrl, solrCollection, &securityConfig, solrContext,
		tlsConfig, false, solrConnectionTimeout}
	// ...
	
	solrClient, err := NewSolrClient(solrConfig)
	// Create a query - example
	solrQuery := solr.CreateSolrQuery()
	solrQuery.Query("*:*")
	// you can set params one-by-one with solrQuery.AddParam or solrQuery.SetParam etc.
	solrClient.Query(&solrQuery)
	
	// Update docs - example 
	solrDoc1 := make(map[string]interface{})
	solrDoc1["id"] = uuid.NewV4().String()
	// ...
	solrDoc2 := make(map[string]interface{})
	solrDoc2["id"] = uuid.NewV4().String()
	// ...
	solrDocs := make([]interface{}, 0)
	solrDocs = append(solrDocs, solrDoc1)
	solrDocs = append(solrDocs, solrDoc2)
	// ...
	solrClient.Update(solrDocs, nil, true)
}
```

### Developement

```bash
make build
```

### Key features
- Basic auth support
- Kerberos support
