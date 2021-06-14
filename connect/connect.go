package connect

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Connect struct {
	Ca       string
	HostCert string
	HostKey  string
	Url      string
	Client   *http.Client
}

func (self *Connect) GetData() (error, []byte) {
	caCert, err := ioutil.ReadFile(self.Ca)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	cert, err := tls.LoadX509KeyPair(self.HostCert, self.HostKey)
	if err != nil {
		log.Fatal(err)
	}

	self.Client = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs:      caCertPool,
				Certificates: []tls.Certificate{cert},
			},
		},
	}
	response, err := self.Client.Get(self.Url)
	if err != nil {
		fmt.Sprintf("%s", err)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Sprintf("%s", err)
		}
		return err, contents
	}
	return err, []byte{byte(0)}
}

func (self *Connect) doReq(payload []byte, soapAction, httpMethod string) bool {
	req, err := http.NewRequest(httpMethod, self.Url, bytes.NewReader(payload))
	if err != nil {
		log.Fatal("Error on creating request object. ", err.Error())
		return false
	}
	req.Header.Set("Content-type", "text/xml")
	req.Header.Set("SOAPAction", "urn:"+soapAction)
	resp, err := self.Client.Do(req)
	if err != nil {
		log.Fatal("Error on dispatching request. ", err.Error())
		return false
	}

	htmlData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer resp.Body.Close()
	fmt.Printf("Status is %v\n", resp.Status)
	fmt.Printf(string(htmlData) + "\n")
	return true
}
