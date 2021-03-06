package main

import (
	"flag"
	"fmt"
	"gitlab.cern.ch/lb-experts/lbconf/connect"
	"gitlab.cern.ch/lb-experts/lbconf/lbconfig"
	"log/syslog"
	"net/http"
	"os"
)

const (
	Lbparamsurl = "https://aiermis.cern.ch/u/api/v1/alias/?format=json&limit=0"
	Localcacert = "/etc/pki/tls/certs/CERN-bundle.pem"
	Pdburl      = "https://constable.cern.ch:9082/pdb/query/v4/resources/Lbd::Client"
)

var configDirFlag = flag.String("configdir", "/usr/local/etc", "specify configuration directory path")
var PartitionFlag = flag.String("partition", "golang", "specify lbd partition")
var debugFlag = flag.Bool("debug", false, "set lbconf in debug mode")
var stdoutFlag = flag.Bool("stdout", false, "send report to stdtout")

func main() {
	flag.Parse()
	Hostname, err := os.Hostname()
	if err != nil {
		fmt.Printf("Hostname Error: %s\n", err.Error())
		os.Exit(1)
	}
	Hostcert := fmt.Sprintf("/var/lib/puppet/ssl/certs/%s.pem", Hostname)
	Hostprivkey := fmt.Sprintf("/var/lib/puppet/ssl/private_keys/%s.pem", Hostname)
	//Hostprivkey := fmt.Sprintf("/afs/cern.ch/user/r/reguero/work/git/golbconf/%s.pem", Hostname)
	Configdir := *configDirFlag
	Reportfile := fmt.Sprintf("%s/load-balancing.report", Configdir)
	Lbheader := fmt.Sprintf("%s/load-balancing.conf-header", Configdir)
	Configfile := fmt.Sprintf("%s/load-balancing.conf", Configdir)

	lg := lbconfig.Log{Writer: syslog.Writer{}, Syslog: false, Stdout: *stdoutFlag, Debugflag: *debugFlag, TofilePath: Reportfile}

	lbconfig := lbconfig.LBConfig{}
	lbconfig.Lbpartition = *PartitionFlag
	lbconfig.Debug = *debugFlag
	lbconfig.Rlog = &lg

	pdb := connect.Connect{
		Ca:       Localcacert,
		HostCert: Hostcert,
		HostKey:  Hostprivkey,
		Url:      Pdburl,
		Client:   &http.Client{}}

	err, aliasresources := pdb.GetData()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	lbp := connect.Connect{
		Ca:       Localcacert,
		HostCert: Hostcert,
		HostKey:  Hostprivkey,
		Url:      Lbparamsurl,
		Client:   &http.Client{}}

	err, lbparams := lbp.GetData()
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	err = lbconfig.Get_alias_resources_from_pdb(aliasresources)
	if err != nil {
		fmt.Printf("Get_alias_resources_from_pdb Error: %s\n", err.Error())
		os.Exit(1)
	}

	err = lbconfig.Get_alias_objects_from_ermis(lbparams)
	if err != nil {
		fmt.Printf("Get_alias_objects_from_ermis Error: %s\n", err.Error())
		os.Exit(1)
	}
	lbconfig.Gen_params()
	lbconfig.Gen_clusters()
	err = lbconfig.Create_config_file(Lbheader, Configfile)
	if err != nil {
		fmt.Printf("Create_config_file Error: %s\n", err.Error())
		os.Exit(1)
	}
}
