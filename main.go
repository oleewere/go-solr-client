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

package main

import "os"
import (
	"flag"
	"fmt"
	"github.com/oleewere/go-solr-client/solr"
	"log"
)

// Version built-in version type
var Version string

// GitRevString built-in git revision string
var GitRevString string

// ActionType type of the action, currently only generator is supported (which is built-in)
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
}
