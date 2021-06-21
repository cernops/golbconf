package main

import (
	"bytes"
	"fmt"
	"gitlab.cern.ch/lb-experts/lbconf/lbconfig"
	"gitlab.cern.ch/lb-experts/lbconf/runner"
	"io/ioutil"
	"log/syslog"
	"testing"
)

func TestLbconf(t *testing.T) {
	//Read files
	aliasresources, err := ioutil.ReadFile("aliasresources")
	if err != nil {
		panic(err)
	}
	lbparams, err := ioutil.ReadFile("lbparams")
	if err != nil {
		panic(err)
	}
	expected, err := ioutil.ReadFile("expected")
	if err != nil {
		panic(err)
	}

	Configdir := "."
	Reportfile := fmt.Sprintf("%s/load-balancing-go.report", Configdir)
	Lbheader := fmt.Sprintf("%s/load-balancing.conf-header", Configdir)
	Configfile := fmt.Sprintf("%s/load-balancing-go.conf", Configdir)

	log, e := syslog.New(syslog.LOG_NOTICE, "lbconf")
	lg := lbconfig.Log{Writer: *log, Syslog: false, Stdout: false, Debugflag: true, TofilePath: Reportfile}
	if e != nil {
		t.Errorf("Error opening log for report: %s\n", err.Error())
	}

	lbconfig := lbconfig.LBConfig{}
	lbconfig.Lbpartition = "golang"
	lbconfig.Debug = true
	lbconfig.Rlog = &lg

	err = lbconfig.Get_alias_resources_from_pdb(aliasresources)
	if err != nil {
		t.Errorf("Get_alias_resources_from_pdb Error: %s\n", err.Error())
	}

	err = lbconfig.Get_alias_objects_from_ermis(lbparams)
	if err != nil {
		t.Errorf("Get_alias_objects_from_ermis Error: %s\n", err.Error())
	}
	lbconfig.Gen_params()
	lbconfig.Gen_clusters()
	err = lbconfig.Create_config_file(Lbheader, Configfile)
	if err != nil {
		t.Errorf("Create_config_file Error: %s\n", err.Error())
	}
	obtained, err := ioutil.ReadFile(Configfile)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(expected, obtained) {
		//t.Errorf("lbconf: got\n %v expected\n %v", obtained, expected)
		//t.Errorf("lbconf: got diffence.\nTry\ndiff  %s %s", Configfile, "expected")
		rawOutput, _, _ := runner.Run("/usr/bin/diff", 0, Configfile, "./expected")
		t.Errorf("lbconf: got diffence:\noutput of diff  %s %s :\n%s\n", Configfile, "./expected", rawOutput)
	}
}
