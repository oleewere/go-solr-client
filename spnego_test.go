package main

import (
	"os"
	"fmt"
	"net/http"
	"github.com/jcmturner/gokrb5/keytab"
	"github.com/jcmturner/gokrb5/client"
	"github.com/jcmturner/gokrb5/config"
	"io/ioutil"
	"testing"
)

func TestGet(t *testing.T) {
	keytabsPath := "/Users/oliverszabo/Projects/ambari-vagrant/centos6.4/"
	solrKeytab := keytabsPath + "ambari-infra-solr.service.keytab"
	krb5ConfPath := keytabsPath + "krb5.conf"
	servicePrincipal := "infra-solr/c6401.ambari.apache.org";
	realm := "AMBARI.APACHE.ORG"
	request := "http://c6402.ambari.apache.org:8886/v2/c/audit_logs"
	requestType := "GET"
	GetWithSPNego(solrKeytab, krb5ConfPath, servicePrincipal, realm, request, requestType)
}

func GetWithSPNego(keytabPath string, krb5confPath string, servicePrincipal string, realm string, request string, requestType string) {

	kt, err := keytab.Load(keytabPath)

	if (err != nil) {
		os.Exit(1)
	}

	cfg, err := config.Load(krb5confPath)

	if (err != nil) {
		os.Exit(1)
	}

	cl := client.NewClientWithKeytab(servicePrincipal, realm, kt)
	cl.WithConfig(cfg)
	errLogin := cl.Login()

	if (errLogin != nil) {
		os.Exit(1)
	}

	r, _ := http.NewRequest(requestType, request, nil)

	spn := ""
	cl.SetSPNEGOHeader(r, spn)
	resp, err := http.DefaultClient.Do(r)

	if (err != nil) {
		os.Exit(1)
	}

	defer resp.Body.Close()

	bodyBytes, err := ioutil.ReadAll(resp.Body)

	if (err != nil) {
		os.Exit(1)
	}
	body := string(bodyBytes)

	fmt.Print(body)
}
