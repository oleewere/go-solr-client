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
	"fmt"
	"github.com/go-ini/ini"
	"github.com/oleewere/go-buffered-processor/processor"
	"github.com/pkg/sftp"
	"github.com/satori/go.uuid"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

// SolrDataProcessor type for processing Solr data
type SolrDataProcessor struct {
	Mutex      *sync.Mutex
	SolrClient *SolrClient
}

// Process send gathered data to Solr
func (p SolrDataProcessor) Process(batchContext *processor.BatchContext) error {
	log.Println("Processing...")
	p.Mutex.Lock()
	defer p.Mutex.Unlock()
	_, _, err := p.SolrClient.Update(batchContext.BufferData, nil, true)
	return err
}

// HandleError handle errors during time based buffer processing (it is not used by this generator)
func (p SolrDataProcessor) HandleError(batchContext *processor.BatchContext, err error) {
	fmt.Println(err)
}

// GenerateSolrData Use to generate Solr data, also scp keytab file to local if kerberos and ssl config is enabled
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
		client, err := ssh.Dial("tcp", sshConfig.Hostname+":22", clientConfig)
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
	numWrites, _ := cfg.Section("generator").Key("num_writes").Int()
	docsPerWrite, _ := cfg.Section("generator").Key("num_docs_per_write").Int()
	clusterField := cfg.Section("generator").Key("cluster_field").String()
	clusterNum, _ := cfg.Section("generator").Key("cluster_num").Int()
	filterableField := cfg.Section("generator").Key("filterable_field").String()
	filterableFieldNum, _ := cfg.Section("generator").Key("filterable_field_num").Int()
	levelField := cfg.Section("generator").Key("level_field").String()
	levels := strings.Split(cfg.Section("generator").Key("level_values").String(), ",")
	typeField := cfg.Section("generator").Key("type_field").String()
	types := strings.Split(cfg.Section("generator").Key("type_values").String(), ",")
	dateField := cfg.Section("generator").Key("date_field").String()
	messageFields := strings.Split(cfg.Section("generator").Key("message_fields").String(), ",")
	numFields := strings.Split(cfg.Section("generator").Key("num_fields").String(), ",")

	solrClient, err := NewSolrClient(solrConfig)
	if err != nil {
		log.Fatal(err)
	}

	batchContext := processor.CreateDefaultBatchContext()
	batchContext.MaxBufferSize = docsPerWrite
	batchContext.MaxRetries = 20
	batchContext.RetryTimeInterval = 10

	proc := SolrDataProcessor{SolrClient: solrClient, Mutex: &sync.Mutex{}}

	for i := 1; i <= numWrites; i++ {
		for j := 1; j <= docsPerWrite; j++ {
			solrDoc := createRandomSolrDoc(clusterField, clusterNum, filterableField, filterableFieldNum, levelField, levels, typeField, types, dateField, messageFields, numFields)

			processor.ProcessData(solrDoc, batchContext, proc)
		}
		randomMsg := fmt.Sprintf("Sending %d documents to Solr: %d/%d ...", docsPerWrite, i, numWrites)
		log.Println(randomMsg)
	}
	proc.Process(batchContext)
	log.Println("Solr random documents generation has finished.")
}

func createRandomSolrDoc(clusterField string, clusterNum int, filterableField string, filterableFieldNum int, levelField string, levels []string,
	typeField string, types []string, dateField string, messageFields []string, numFields []string) map[string]interface{} {
	solrDoc := make(map[string]interface{})
	solrDoc["id"] = uuid.NewV4().String()
	solrDoc[clusterField] = "cluster" + fmt.Sprintf("%d", rand.Intn(clusterNum))
	solrDoc[filterableField] = "random-name-" + fmt.Sprintf("%d", rand.Intn(filterableFieldNum))
	solrDoc[levelField] = levels[rand.Int()%len(levels)]
	solrDoc[typeField] = types[rand.Int()%len(types)]
	solrDoc[dateField] = time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
	for _, msgField := range messageFields {
		solrDoc[msgField] = "Random message: " + uuid.NewV4().String()
	}
	for _, nField := range numFields {
		solrDoc[nField] = rand.Intn(3000)
	}
	return solrDoc
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
