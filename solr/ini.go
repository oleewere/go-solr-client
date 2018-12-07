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
	"github.com/go-ini/ini"
	"log"
)

// GenerateIniFile create an ini file to a specific location
func GenerateIniFile(iniFileLocation string) {
	log.Println("Generating new INI config file: " + iniFileLocation)
	cfg := ini.Empty()

	cfg.NewSection("security")
	cfg.Section("security").NewKey("kerberosEnabled", "false")
	cfg.Section("security").NewKey("kerberosKeytab", "/tmp/solr.keytab")
	cfg.Section("security").NewKey("kerberosPrincipal", "solr/myhostname")
	cfg.Section("security").NewKey("kerberosRealm", "EXAMPLE.COM")
	cfg.Section("security").NewKey("kerberosKrb5Path", "/tmp/krb5.conf")

	cfg.NewSection("solr")
	cfg.Section("solr").NewKey("url", "http://localhost:8983")
	cfg.Section("solr").NewKey("context", "/solr")
	cfg.Section("solr").NewKey("collection", "hadoop_logs")
	cfg.Section("solr").NewKey("ssl", "false")
	cfg.Section("solr").NewKey("connection_timeout", "60")

	cfg.NewSection("ssh")
	cfg.Section("ssh").NewKey("enabled", "false")
	cfg.Section("ssh").NewKey("username", "root")
	cfg.Section("ssh").NewKey("hostname", "myremotehost")
	cfg.Section("ssh").NewKey("private_key_path", "/keys/private_key")
	cfg.Section("ssh").NewKey("download_location", "/tmp")
	cfg.Section("ssh").NewKey("remote_krb5_conf", "/etc/krb5.conf")
	cfg.Section("ssh").NewKey("remote_keytab", "/etc/security/keytabs/solr.service.keytab")

	cfg.NewSection("generator")
	cfg.Section("generator").NewKey("num_writes", "10")
	cfg.Section("generator").NewKey("num_docs_per_write", "1000")
	cfg.Section("generator").NewKey("clusters_field", "cluster")
	cfg.Section("generator").NewKey("clusters_num", "10")
	cfg.Section("generator").NewKey("filterable_field", "host")
	cfg.Section("generator").NewKey("filterable_field_num", "1000")
	cfg.Section("generator").NewKey("level_field", "level")
	cfg.Section("generator").NewKey("level_values", "INFO,DEBUG,FATAL,WARN,ERROR,UNKNOWN,TRACE")
	cfg.Section("generator").NewKey("type_field", "type")
	cfg.Section("generator").NewKey("type_values", "ambari_server,ambari_agent,ambari_config,ambari_eclipselink,hdfs_name_node,hdfs_secondary_name_node")
	cfg.Section("generator").NewKey("date_field", "logtime")
	cfg.Section("generator").NewKey("message_fields", "log_message")
	cfg.Section("generator").NewKey("num_fields", "seq_num")
	cfg.SaveTo(iniFileLocation)
}

// GenerateSolrConfig create sample ini file for Solr data generation
func GenerateSolrConfig(iniFileLocation string) (SolrConfig, SSHConfig) {
	cfg, err := ini.Load(iniFileLocation)
	if err != nil {
		log.Fatal("Fail to read file: " + iniFileLocation)
	}

	kerberosEnabled, _ := cfg.Section("security").Key("kerberosEnabled").Bool()
	keytabPath := cfg.Section("security").Key("kerberosKeytab").String()
	principal := cfg.Section("security").Key("kerberosPrincipal").String()
	realm := cfg.Section("security").Key("kerberosRealm").String()
	krb5Path := cfg.Section("security").Key("kerberosKrb5Path").String()

	solrUrl := cfg.Section("solr").Key("url").String()
	solrContext := cfg.Section("solr").Key("context").String()
	solrCollection := cfg.Section("solr").Key("collection").String()
	solrTlsEnabled, _ := cfg.Section("solr").Key("ssl").Bool()
	solrConnectionTimeout, _ := cfg.Section("solr").Key("connection_timeout").Int()

	sshEnabled, _ := cfg.Section("ssh").Key("enabled").Bool()
	sshUsername := cfg.Section("ssh").Key("username").String()
	sshHostname := cfg.Section("ssh").Key("hostname").String()
	sshPrivateKeyPath := cfg.Section("ssh").Key("private_key_path").String()
	sshDownloadLocation := cfg.Section("ssh").Key("download_location").String()
	remoteKrb5Conf := cfg.Section("ssh").Key("remote_krb5_conf").String()
	remoteKeytab := cfg.Section("ssh").Key("remote_keytab").String()

	securityConfig := SecurityConfig{}
	if kerberosEnabled {
		securityConfig = InitSecurityConfig(krb5Path, keytabPath, principal, realm)
	}

	solrConfig := SolrConfig{solrUrl, solrCollection, &securityConfig, solrContext,
		TLSConfig{}, !solrTlsEnabled, solrConnectionTimeout}

	sshConfig := SSHConfig{Enabled: sshEnabled, Username: sshUsername, PrivateKeyPath: sshPrivateKeyPath,
		DownloadLocation: sshDownloadLocation, RemoteKrb5Conf: remoteKrb5Conf, RemoteKeytab: remoteKeytab, Hostname: sshHostname}

	return solrConfig, sshConfig
}
