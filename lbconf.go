//  {
//    "tags": [
//      "hg_castor",
//      "loadbalancing",
//      "lbclient",
//      "castor::loadbalancing",
//      "class",
//      "lbd",
//      "castorpublic.cern.ch",
//      "alias",
//      "lbd::client",
//      "client",
//      "lbclient::alias",
//      "castor"
//    ],
//    "file": "/mnt/puppetnfsdir/environments/castor_flume_highmem/modules/lbclient/manifests/alias.pp",
//    "type": "Lbd::Client",
//    "title": "castorpublic.cern.ch",
//    "line": 15,
//    "resource": "fe689c50d9cfcea934c7678cb9c7f9be57745ac4",
//    "environment": "castor_flume_highmem",
//    "certname": "c2public-3.cern.ch",
//    "parameters": {
//      "comment": "castorpublic.cern.ch",
//      "lbalias": "castorpublic.cern.ch",
//      "clienthostgroup": "castor/c2public/headnode"
//    },
//    "exported": false
//  }

//{"meta": {"limit": 1000, "next": null, "offset": 0, "previous": null, "total_count": 990}, "objects": [{"AllowedNodes": null, "ForbiddenNodes": null, "alias_name":
// {
//    "AllowedNodes": "ipa-dev-3.ipa-dev.cern.ch,ipa-dev-4.ipa-dev.cern.ch",
//    "ForbiddenNodes": null,
//    "alias_name": "ldap-lb.ipa-dev.cern.ch",
//    "behaviour": "mindless",
//    "best_hosts": 2,
//    "clusters": "",
//    "cnames": [],
//    "external": "no",
//    "hostgroup": "ipa_dev",
//    "id": 3744,
//    "last_modification": "2020-05-28T00:00:00",
//    "metric": "cmsfrontier",
//    "polling_interval": 300,
//    "resource_uri": "/u/api/v1/alias/3744/",
//    "statistics": "long",
//    "tenant": "golang",
//    "ttl": 60,
//    "user": "psaiz"
//  },

package main

import (
	"encoding/json"
	"fmt"
	"gitlab.cern.ch/lb-experts/lbconf/connect"
	"net/http"
	"sort"
	//"strings"
)

const (
	Lbparamsurl = "https://aiermis.cern.ch/u/api/v1/alias/?format=json&limit=0"
	Localcacert = "/etc/pki/tls/certs/CERN-bundle.pem"
	Pdburl      = "https://constable.cern.ch:9082/pdb/query/v4/resources/Lbd::Client"
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

var resources []Resource
var SearchResp LbaliasBlob
var MembersPerAlias map[string][]string
var Clhostgroup map[string]string

func main() {
	Hostname := "aiadm83.cern.ch"
	Hostcert := fmt.Sprintf("/var/lib/puppet/ssl/certs/%s.pem", Hostname)
	//Hostprivkey := fmt.Sprintf("/var/lib/puppet/ssl/private_keys/%s.pem", Hostname)
	Hostprivkey := fmt.Sprintf("/afs/cern.ch/user/r/reguero/work/git/golbconf/%s.pem", Hostname)
	Lbpartition := "golang"
	//Configdir := "/usr/local/etc"
	//Reportfile := fmt.Sprintf("%s/load-balancing.report", Configdir)
	//Lbheader := fmt.Sprintf("%s/load-balancing.conf-header", Configdir)
	//Configfile := fmt.Sprintf("%s/load-balancing.conf", Configdir)

	pdb := connect.Connect{
		Ca:       Localcacert,
		HostCert: Hostcert,
		HostKey:  Hostprivkey,
		Url:      Pdburl,
		Client:   &http.Client{}}

	err, aliasresources := pdb.GetData()
	if err != nil {
		fmt.Sprintf("%s", err)
	} else {
		//fmt.Println(string(aliasresources))
		json.Unmarshal(aliasresources, &resources)
		//fmt.Printf("resources : %+v", resources)
		//  # Generate hash of hosts members per lbalias
		MembersPerAlias = make(map[string][]string)
		for _, r := range resources {
			lbalias := r.Parameters.Lbalias
			MembersPerAlias[lbalias] = append(MembersPerAlias[lbalias], r.Certname)
		}
		//  # Generate hash of hostgroup per host from the information in the resources
		Clhostgroup = make(map[string]string)
		for _, r := range resources {
			Clhostgroup[r.Certname] = r.Parameters.Clienthostgroup
		}

	}
	//fmt.Printf("MembersPerAlias : %+v", MembersPerAlias)
	//for k, v := range MembersPerAlias {
	//	fmt.Printf("key[%s] value[%s]\n", k, v)
	//}
	//for k, v := range Clhostgroup {
	//	fmt.Printf("key[%s] value[%s]\n", k, v)
	//}

	lbp := connect.Connect{
		Ca:       Localcacert,
		HostCert: Hostcert,
		HostKey:  Hostprivkey,
		Url:      Lbparamsurl,
		Client:   &http.Client{}}

	err, lbparams := lbp.GetData()
	if err != nil {
		fmt.Sprintf("%s", err)
	} else {
		//fmt.Println(string(lbparams))
		json.Unmarshal(lbparams, &SearchResp)
		//fmt.Printf("Meta : %+v", SearchResp.Meta)
		//fmt.Printf("Object Array : %+v", SearchResp.Objects)
		//for _, o := range SearchResp.Objects {
		//	fmt.Printf("Object : %+v\n", o)
		//}
		aliasdef := make(ObjectList, len(SearchResp.Objects))
		for i, v := range SearchResp.Objects {
			aliasdef[i] = v
		}
		sort.Sort(aliasdef)
		ttl := 60
		outputlst := make([]string, len(aliasdef))
		for _, o := range aliasdef {
			//fmt.Printf("Object : %+v\n", o)
			//fmt.Printf("Alias : %+v\n", o.Alias_name)
			// Filter by Lbpartition
			if o.Tenant != Lbpartition {
				continue
			}
			if o.Ttl != 0 {
				//fmt.Printf("Ttl : %+v ", o.Ttl)
				ttl = o.Ttl
			}
			line := fmt.Sprintf("parameters %s = behaviour#%s best_hosts#%d external#%s metric#%s polling_interval#%d statistics#%s ttl#%d", o.Alias_name, o.Behaviour, o.Best_hosts, o.External, o.Metric, o.Polling_interval, o.Statistics, ttl)
			outputlst = append(outputlst, line)
			fmt.Println(line)
		}
		//output := strings.Join(outputlst[:], "\n")
		//fmt.Printf(output)
		//fmt.Printf("outputlst : %+v\n", outputlst[:])
	}
}
