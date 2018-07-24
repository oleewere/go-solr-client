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
	"io/ioutil"
	"golang.org/x/crypto/ssh"
	"log"
	"os"
	"fmt"
	"github.com/pkg/sftp"
	"github.com/go-ini/ini"
	"github.com/satori/go.uuid"
	"strings"
)

// Use to generate Solr data, also scp keytab file to local if kerberos and ssl config is enabled
func GenerateSolrData(solrConfig *SolrConfig, sshConfig *SSHConfig, iniFileLocation string) {
	if sshConfig.Enabled {
		privateKeyContent, err := ioutil.ReadFile(sshConfig.PrivateKeyPath)
		if err != nil {
			log.Fatal(err)
		}
		signer, _ := ssh.ParsePrivateKey([]byte(privateKeyContent))
		clientConfig := &ssh.ClientConfig{
			User: sshConfig.Username,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		client, err := ssh.Dial("tcp", sshConfig.Hostname + ":22", clientConfig)
		if err != nil {
			panic("Failed to dial: " + err.Error())
		}
		sftpClient, err := sftp.NewClient(client)
		if err != nil {
			log.Fatal(err)
		}
		defer sftpClient.Close()

		copyFileToLocal(sshConfig.RemoteKrb5Conf, solrConfig.SecurityConfig.kerberosConfig.krb5confPath, sftpClient)
		copyFileToLocal(sshConfig.RemoteKeytab, solrConfig.SecurityConfig.kerberosConfig.keytab, sftpClient)
	}

	cfg, err := ini.Load(iniFileLocation)
	if err != nil {
		log.Fatal("Fail to read file: " + iniFileLocation)
	}
	numWrites, err := cfg.Section("generator").Key("num_writes").Int()
	clusterField := cfg.Section("generator").Key("cluster_field").String()
	clusterNum, err := cfg.Section("generator").Key("cluster_num").Int()
	hostnameField := cfg.Section("generator").Key("hostname_field").String()
	hostnameNum, err := cfg.Section("generator").Key("hostname_num").Int()
	levelField := cfg.Section("generator").Key("level_field").String()
	levels := strings.Split(cfg.Section("generator").Key("level_values").String(), ",")
	typeField := cfg.Section("generator").Key("type_field").String()
	types := strings.Split(cfg.Section("generator").Key("type_values").String(), ",")
	dateField := cfg.Section("generator").Key("date_field").String()
	datePattern := cfg.Section("generator").Key("date_pattern").String()
	messageFields := strings.Split(cfg.Section("generator").Key("message_fields").String(), ",")
	numFields := strings.Split(cfg.Section("generator").Key("num_fields").String(), ",")

	for i := 1; i <= numWrites; i++ {
		putDocs := SolrDocuments{}
		for i := 1; i <= 10; i++ {
			solrDoc :=  make(map[string]interface{})
			solrDoc["id"] = uuid.NewV4()

			putDocs = append(putDocs, solrDoc)
		}
	}

	os.Exit(0)

	solrClient, err := NewSolrClient(solrConfig)

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

	if err != nil {
		fmt.Print(err)
	}
}
func copyFileToLocal(srcFilePath string, destFilePath string, sftpClient *sftp.Client) {
	srcFile, err := sftpClient.Open(srcFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer srcFile.Close()
	destFile, err := os.Create(destFilePath)
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()
	srcFile.WriteTo(destFile)
}
