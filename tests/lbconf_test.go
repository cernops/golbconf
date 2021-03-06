package main

import (
	"bytes"
	"fmt"
	"github.com/pmezard/go-difflib/difflib"
	"gitlab.cern.ch/lb-experts/lbconf/lbconfig"
	"io/ioutil"
	"log/syslog"
	"testing"
)

func TestLbconf(t *testing.T) {
	//Read files
	aliasresources, err := ioutil.ReadFile("aliasresources")
	if err != nil {
		t.Errorf("Error opening aliasresources: %s\n", err.Error())
	}
	lbparams, err := ioutil.ReadFile("lbparams")
	if err != nil {
		t.Errorf("Error opening lbparams: %s\n", err.Error())
	}
	expected, err := ioutil.ReadFile("expected")
	if err != nil {
		t.Errorf("Error opening expected: %s\n", err.Error())
	}

	Configdir := "."
	Reportfile := fmt.Sprintf("%s/load-balancing-go.report", Configdir)
	Lbheader := fmt.Sprintf("%s/load-balancing.conf-header", Configdir)
	Configfile := fmt.Sprintf("%s/load-balancing-go.conf", Configdir)

	lg := lbconfig.Log{Writer: syslog.Writer{}, Syslog: false, Stdout: false, Debugflag: true, TofilePath: Reportfile}

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
		t.Errorf("Error opening Configfile: %s\n", err.Error())
	}
	if !bytes.Equal(expected, obtained) {
		//t.Errorf("lbconf: got\n %v expected\n %v", obtained, expected)
		//t.Errorf("lbconf: got diffence.\nTry\ndiff  %s %s", Configfile, "expected")
		diff := difflib.UnifiedDiff{
			A:        difflib.SplitLines(string(expected)),
			B:        difflib.SplitLines(string(obtained)),
			FromFile: "Expected",
			ToFile:   "Obtained",
			Context:  3,
		}
		rawOutput, _ := difflib.GetUnifiedDiffString(diff)
		t.Errorf("lbconf: got diffence:\noutput of diff  %s %s :\n%s\n", Configfile, "./expected", rawOutput)
	}
}
