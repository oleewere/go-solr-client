package main

import "os"
import (
	"flag"
	"fmt"
	"github.com/oleewere/go-solr-client/solr"
	"log"
)

var Version string
var GitRevString string
var ActionType string

func main() {

	var isVersionCheck bool
	var iniFileLocation string

	flag.BoolVar(&isVersionCheck, "version", false, "Print application version and git revision if available")
	flag.StringVar(&iniFileLocation, "ini-file", "", "INI config file location")
	if len(ActionType) == 0 {
		flag.StringVar(&ActionType, "action-type", "generator", "action")
	}

	flag.Parse()

	if isVersionCheck {
		if GitRevString == "" {
			fmt.Println("version:", Version)
		} else {
			fmt.Printf("version: %s (git revision: %s)\n", Version, GitRevString)
		}
		os.Exit(0)
	}

	if len(iniFileLocation) == 0 {
		log.Fatal("INI config file option (--ini-file) is missing.")
	}

	if _, err := os.Stat(iniFileLocation); os.IsNotExist(err) {
		solr.GenerateIniFile(iniFileLocation)
		os.Exit(0)
	}
	log.Println("Starting Solr Client ...")
	solrConfig, sshConfig := solr.GenerateSolrConfig(iniFileLocation)
	solr.GenerateSolrData(&solrConfig, &sshConfig, iniFileLocation)
    /*
	solrClient, err := solr.NewSolrClient(&solrConfig)
	_, response, _ := solrClient.Query(nil)
	docs := response.Response.Docs
	for _, doc := range docs {
		fmt.Printf("----------------------")
		for k, v := range doc {
			fmt.Print("key: ", k)
			fmt.Println(" , value: ", v)
		}
		fmt.Printf("----------------------")
	}
    */
}
