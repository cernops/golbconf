package lbconfig

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/lbconf/connect"
	"os"
	"sort"
	"strings"
)

type LbaliasObject struct {
	AllowedNodes      string   `json:"AllowedNodes"`
	ForbiddenNodes    string   `json:"ForbiddenNodes"`
	Alias_name        string   `json:"alias_name"`
	Behaviour         string   `json:"behaviour"`
	Best_hosts        int      `json:"best_hosts"`
	Clusters          string   `json:"clusters"`
	Cnames            []string `json:"cnames"`
	External          string   `json:"external"`
	Hostgroup         string   `json:"hostgroup"`
	Id                int      `json:"id"`
	Last_modification string   `json:"last_modification"`
	Metric            string   `json:"metric"`
	Polling_interval  int      `json:"polling_interval"`
	Resource_uri      string   `json:"resource_uri"`
	Statistics        string   `json:"statistics"`
	Tenant            string   `json:"tenant"`
	Ttl               int      `json:"ttl"`
	User              string   `json:"user"`
}

//ObjectList struct for the list
type ObjectList []LbaliasObject

func (p ObjectList) Len() int           { return len(p) }
func (p ObjectList) Less(i, j int) bool { return p[i].Alias_name < p[j].Alias_name }
func (p ObjectList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

type LbaliasMeta struct {
	Limit       int `json:"limit"`
	Next        int `json:"next"`
	Offset      int `json:"offset"`
	Previous    int `json:"previous"`
	Total_count int `json:"total_count"`
}

type LbaliasBlob struct {
	Meta    LbaliasMeta     `json:"meta"`
	Objects []LbaliasObject `json:"objects"`
}

type LbdClientParams struct {
	Comment         string `json:"comment"`
	Lbalias         string `json:"lbalias"`
	Clienthostgroup string `json:"clienthostgroup"`
}

type Resource struct {
	Parameters LbdClientParams `json:"parameters"`
	Line       int             `json:"line"`
	Exported   bool            `json:"exported"`
	Tags       []string        `json:"tags"`
	Title      string          `json:"title"`
	Type       string          `json:"type"`
	Resource   string          `json:"resource"`
	Certname   string          `json:"certname"`
}

type LBConfig struct {
	resources       []Resource
	searchResp      LbaliasBlob
	MembersPerAlias map[string][]string
	Clhostgroup     map[string]string
	Aliasdef        ObjectList
	outputlst       []string
	Lbpartition     string
	Debug           bool
	Rlog            *Log
}

func isIncludedIn(list []string, key string) bool {
	for _, a := range list {
		if len(a) == 0 {
			continue
		}
		if a == key {
			return true
		}
	}
	return false
}

func removeDuplicates(listwithdups []string) []string {
	allKeys := make(map[string]bool)
	outputlist := []string{}
	for _, item := range listwithdups {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			outputlist = append(outputlist, item)
		}
	}
	return outputlist
}

func (lbc *LBConfig) Get_alias_resources_from_pdb(pdb connect.Connect) error {
	err, aliasresources := pdb.GetData()
	if err != nil {
		if lbc.Debug {
			fmt.Printf("Error: %s\n", err.Error())
			return err
		}
	}
	if err := json.Unmarshal(aliasresources, &lbc.resources); err != nil {
		if lbc.Debug {
			fmt.Printf("Error: %s\n", err.Error())
			fmt.Printf("Here follows the aliasresources data : %s\n", string(aliasresources))
		}
		return err
	}
	// Generate hash of hosts members per LB alias
	lbc.MembersPerAlias = make(map[string][]string)
	for _, r := range lbc.resources {
		lbalias := r.Parameters.Lbalias
		lbc.MembersPerAlias[lbalias] = append(lbc.MembersPerAlias[lbalias], r.Certname)
	}
	// Generate hash of hostgroup per host from the information in the PDB resources
	lbc.Clhostgroup = make(map[string]string)
	for _, r := range lbc.resources {
		lbc.Clhostgroup[r.Certname] = r.Parameters.Clienthostgroup
	}

	return nil
}

func (lbc *LBConfig) Get_alias_objects_from_ermis(lbp connect.Connect) error {
	err, lbparams := lbp.GetData()
	if err != nil {
		if lbc.Debug {
			fmt.Printf("Error: %s\n", err.Error())
		}
		return err
	}
	if err := json.Unmarshal(lbparams, &lbc.searchResp); err != nil {
		if lbc.Debug {
			fmt.Printf("Error: %s\n", err.Error())
			fmt.Printf("Here follows the lbparams data : %s\n", string(lbparams))
		}
		return err
	}
	// Get just the ObjectList to avoid the metadata
	lbc.Aliasdef = make(ObjectList, len(lbc.searchResp.Objects))
	for i, v := range lbc.searchResp.Objects {
		lbc.Aliasdef[i] = v
	}
	sort.Sort(lbc.Aliasdef)
	return nil
}

func (lbc *LBConfig) Gen_params() {
	lbc.outputlst = make([]string, len(lbc.Aliasdef)*2)
	for _, o := range lbc.Aliasdef {
		// Filter by Lbpartition
		if o.Tenant != lbc.Lbpartition {
			continue
		}
		ttl := 60
		if o.Ttl != 0 {
			ttl = o.Ttl
		}
		line := fmt.Sprintf("parameters %s = behaviour#%s best_hosts#%d external#%s metric#%s polling_interval#%d statistics#%s ttl#%d", o.Alias_name, o.Behaviour, o.Best_hosts, o.External, o.Metric, o.Polling_interval, o.Statistics, ttl)
		lbc.outputlst = append(lbc.outputlst, line)
	}

}

func (lbc *LBConfig) Gen_clusters() {
	// Let's populate all the possible cnames"
	cnames := make([]string, len(lbc.Aliasdef), len(lbc.Aliasdef)*8)
	for _, o := range lbc.Aliasdef {
		if len(o.Cnames) != 0 {
			cnames = append(cnames, o.Cnames...)
		}
	}
	// make list of valid aliases
	aliaslst := make([]string, len(lbc.Aliasdef))
	for _, o := range lbc.Aliasdef {
		if len(o.Alias_name) != 0 {
			aliaslst = append(aliaslst, o.Alias_name)
		}
	}
	for k, v := range lbc.MembersPerAlias {
		if !isIncludedIn(aliaslst, k) {
			if isIncludedIn(cnames, strings.Split(k, ".")[0]) {
				lbc.Write_to_report("WARN", fmt.Sprintf("category:alias_cname cluster:%s the alias is a canonical name record", k))
			} else {
				sort.Strings(v)
				lbc.Write_to_report("ERROR", fmt.Sprintf("category:alias_not_in_config cluster:%s alias not in configuration. Pointed by the following host(s) %s", k, strings.Join(v, " ")))
			}
		}
	}
	for _, o := range lbc.Aliasdef {
		if o.Tenant != lbc.Lbpartition {
			continue
		}
		fnodes := []string{}
		if o.ForbiddenNodes != "" {
			fnodes = strings.Split(o.ForbiddenNodes, ",")
		}
		mblist := make([]string, 50, 1000)
		if _, ok := lbc.MembersPerAlias[o.Alias_name]; ok {
			if o.Hostgroup != "" {
				sort.Strings(lbc.MembersPerAlias[o.Alias_name])
				for _, m := range lbc.MembersPerAlias[o.Alias_name] {
					oh := strings.Split(o.Hostgroup, "/")
					ch := strings.Split(lbc.Clhostgroup[m], "/")
					if oh[0] == ch[0] {
						if !isIncludedIn(fnodes, m) {
							mblist = append(mblist, m)
						}
					} else {
						lbc.Write_to_report("ERROR", fmt.Sprintf("category:wrong_hostgroup cluster:%s hostgroup=%s member=%s with wrong hostgroup=%s", o.Alias_name, o.Hostgroup, m, lbc.Clhostgroup[m]))
					}
				}
			} else {
				lbc.Write_to_report("ERROR", fmt.Sprintf("category:no_hostgroup cluster:%s has no hostgroup defined", o.Alias_name))
				for _, m := range lbc.MembersPerAlias[o.Alias_name] {
					if !isIncludedIn(fnodes, m) {
						mblist = append(mblist, m)
					}
					lbc.Write_to_report("ERROR", fmt.Sprintf("category:no_hostgroup cluster:%s member=%s hostgroup=%s", o.Alias_name, m, lbc.Clhostgroup[m]))
					lbc.Write_to_report("ERROR", fmt.Sprintf("category:no_hostgroup cluster:%s update ermis_api_alias set hostgroup = '%s'  where alias_name = '%s';", o.Alias_name, lbc.Clhostgroup[m], o.Alias_name))
				}
			}
			// Include allowed nodes that are not in PuppetDB
			anodes := strings.Split(o.AllowedNodes, ",")
			mblist = append(mblist, anodes...)
			uniqmblist := removeDuplicates(mblist)
			sort.Strings(uniqmblist)
			lbc.outputlst = append(lbc.outputlst, fmt.Sprintf("clusters %s =%s", o.Alias_name, strings.Join(uniqmblist, " ")))
		} else {
			// Include allowed nodes for an alias with no members
			if o.AllowedNodes != "" {
				anodes := strings.Split(o.AllowedNodes, ",")
				mblist = append(mblist, anodes...)
				uniqmblist := removeDuplicates(mblist)
				sort.Strings(uniqmblist)
				lbc.outputlst = append(lbc.outputlst, fmt.Sprintf("clusters %s =%s", o.Alias_name, strings.Join(uniqmblist, " ")))

			} else {
				lbc.Write_to_report("ERROR", fmt.Sprintf("category:no_members cluster:%s has no members", o.Alias_name))
			}
		}
	}
}

// readLines reads a whole file into memory and returns a slice of lines.
func readLines(path string) (lines []string, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

//remove empty lines from array
func removeEmpty(list []string) []string {
	res := []string{}
	for _, s := range list {
		if s != "" {
			res = append(res, s)
		}
	}
	return res
}

func (lbc *LBConfig) Create_config_file(Lbheader string, Configfile string) error {
	prevfile := Configfile[0:len(Configfile)-5] + "prev"
	newfile := Configfile[0:len(Configfile)-5] + "new"
	headerlines, err := readLines(Lbheader)
	if err != nil {
		if lbc.Debug {
			fmt.Printf("readLines Error: %s\n", err.Error())
		}
		return err
	}
	f, err := os.OpenFile(newfile, os.O_TRUNC|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		if lbc.Debug {
			fmt.Printf("can not open %v for writing: %v", newfile, err)
		}
		return err
	}
	headerlines = append(headerlines, "")
	for _, l := range append(headerlines, removeEmpty(lbc.outputlst)...) {
		_, err = fmt.Fprintf(f, "%s\n", l)
		if err != nil {
			if lbc.Debug {
				fmt.Printf("can not write to %v: %v", newfile, err)
			}
		}
	}
	f.Close()

	if _, err := os.Stat(Configfile); err == nil {
		// File exists
		if err = os.Rename(Configfile, prevfile); err != nil {
			if lbc.Debug {
				fmt.Printf("can not rename %v to %v: %v", Configfile, prevfile, err)
			}
			return err
		}

	} else if !os.IsNotExist(err) {
		if lbc.Debug {
			fmt.Printf("Stat Error: %s\n", err.Error())
		}
		return err
	}
	if err = os.Rename(newfile, Configfile); err != nil {
		if lbc.Debug {
			fmt.Printf("can not rename %v to %v: %v", newfile, Configfile, err)
		}
		return err
	}
	return nil

}
