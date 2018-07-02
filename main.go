package main

import "os"
import (
    "flag"
	"fmt"
	"github.com/oleewere/native-solr-client/solr"
)

func main()  {
	fmt.Print("Start Solr Cloud Client ...\n")

	var collection string
    var url string
    var krb5Path string
    var keytabPath string
    var principal string
    var realm string

	flag.StringVar(&url, "url", "http://localhost:8983" , "URL name for Solr or Solr proxy")
	flag.StringVar(&collection, "collection", "hadoop_logs" , "Collection name for the Solr client")
	flag.StringVar(&krb5Path, "krb-conf-path", "" , "Kerberos config location")
	flag.StringVar(&keytabPath, "keytab-path", "" , "Kerberos keytab location")
	flag.StringVar(&principal, "principal", "" , "Kerberos principal")
	flag.StringVar(&realm, "realm", "" , "Kerberos Realm e.g.: EXAMPLE.COM")

	flag.Parse()

	securityConfig := solr.InitSecurityConfig(krb5Path, keytabPath, principal, realm)

	solrConfig := solr.SolrConfig{ url, "hadoop_logs", &securityConfig, "/solr",
	solr.TLSConfig{}, true, 60, }
	solrClient, err := solr.NewSolrClient(url, collection, &solrConfig)
	solrClient.Query("q=*:*")

	if err != nil {
		fmt.Print(err)
	}

	os.Exit(0)
}
