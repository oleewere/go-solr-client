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
	"fmt"
	"os"
	"io"
)

// Use to generate Solr data, also scp keytab file to local if kerberos and ssl config is enabled
func GenerateSolrData(solrConfig *SolrConfig, sshConfig *SSHConfig) {

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
		sshClient, err := ssh.Dial("tcp", sshConfig.Hostname + ":22", clientConfig)
		if err != nil {
			panic("Failed to dial: " + err.Error())
		}
		session, err := sshClient.NewSession()
		if err != nil {
			panic("Failed to create session: " + err.Error())
		}
		defer session.Close()

		r, err := session.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		name := fmt.Sprintf("%s/krb5.conf", sshConfig.DownloadLocation)
		file, err := os.OpenFile(name, os.O_APPEND | os.O_WRONLY | os.O_CREATE, 0644)
		if err != nil {
			log.Fatal(err)
		}

		if err := session.Start("cat /etc/krb5.conf"); err != nil {
			log.Fatal(err)
		}

		n, err := io.Copy(file, r)
		if err != nil {
			log.Fatal(err)
		}

		log.Print(n)

		if err := session.Wait(); err != nil {
			log.Fatal(err)
		}

	}
}
