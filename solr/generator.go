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
	"math/rand"
	"time"
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
	docsPerWrite, err := cfg.Section("generator").Key("num_docs_per_write").Int()
	clusterField := cfg.Section("generator").Key("cluster_field").String()
	clusterNum, err := cfg.Section("generator").Key("cluster_num").Int()
	filterableField := cfg.Section("generator").Key("filterable_field").String()
	filterableFieldNum, err := cfg.Section("generator").Key("filterable_field_num").Int()
	levelField := cfg.Section("generator").Key("level_field").String()
	levels := strings.Split(cfg.Section("generator").Key("level_values").String(), ",")
	typeField := cfg.Section("generator").Key("type_field").String()
	types := strings.Split(cfg.Section("generator").Key("type_values").String(), ",")
	dateField := cfg.Section("generator").Key("date_field").String()
	messageFields := strings.Split(cfg.Section("generator").Key("message_fields").String(), ",")
	numFields := strings.Split(cfg.Section("generator").Key("num_fields").String(), ",")

	solrClient, err := NewSolrClient(solrConfig)

	for i := 1; i <= numWrites; i++ {
		putDocs := SolrDocuments{}
		for j := 1; j <= docsPerWrite; j++ {
			solrDoc :=  make(map[string]interface{})
			solrDoc["id"] = uuid.NewV4().String()
			solrDoc[clusterField] = "cluster" + fmt.Sprintf("%d", rand.Intn(clusterNum))
			solrDoc[filterableField] = "random-name-" + fmt.Sprintf("%d", rand.Intn(filterableFieldNum))
			solrDoc[levelField] = levels[rand.Int() % len(levels)]
			solrDoc[typeField] = types[rand.Int() % len(types)]
			solrDoc[dateField] = time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
			for _, msgField := range messageFields {
				solrDoc[msgField] = "Random message: " + uuid.NewV4().String()
			}

			for _, nField := range numFields {
				solrDoc[nField] = rand.Intn(3000)
			}

			putDocs = append(putDocs, solrDoc)
		}
		randomMsg := fmt.Sprintf("Sending %d documents to Solr: %d/%d ...", docsPerWrite, i, numWrites)
		log.Println(randomMsg)
		solrClient.Update(putDocs, nil, false)
	}
	log.Println("Solr random documents generation has finished.")
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
