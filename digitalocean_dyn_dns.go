//Run with cron (or whatever you want to use), for example every 5 or 10 mins.
//Change the access token to the one that digitalocean gave you (Personal access token)
//In order to this to work you have to have created first the domain that you want to use
//(it won't create it, it only updates it)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/digitalocean/godo"
	"golang.org/x/oauth2"
)

const (
	accessToken = "your token"
	domain      = "your domain"
)

type MyIp struct {
	Ip       string `json:"ip"`
	Hostname string `json:""`
	City     string `json:""`
	Region   string `json:""`
	Country  string `json:""`
	Loc      string `json:""`
	Org      string `json:""`
}

type TokenSource struct {
	AccessToken string
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
	token := &oauth2.Token{AccessToken: t.AccessToken}
	fmt.Printf("Using token: %v... for domain: %v\n", accessToken[:5], domain)
	return token, nil
}

func main() {
	tokenSource := &TokenSource{AccessToken: accessToken}
	changeDnsIp(tokenSource, domain)
}

func getOwnIp() string {
	resp, err := http.Get("http://ipinfo.io/json")
	if err != nil {
		fmt.Printf(err.Error())
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	response := MyIp{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		fmt.Printf(err.Error())
	}
	return response.Ip
}

func changeDnsIp(accessToken *TokenSource, domainName string) {
	oauth_client := oauth2.NewClient(context.Background(), accessToken)
	client := godo.NewClient(oauth_client)
	listOps := godo.ListOptions{Page: 1, PerPage: 50}
	var err error = nil

	records, _, err := client.Domains.Records(context.Background(), domainName, &listOps)
	if err != nil {
		fmt.Println("Error getting records")
		fmt.Println(err.Error())
	}
	var ipRecord = godo.DomainRecord{}
	for _, r := range records {
		if r.Type == "A" {
			ipRecord = r
		}
	}

	editRequest := godo.DomainRecordEditRequest{Data: getOwnIp()}
	fmt.Printf("Updating record %v to new ip: %v\n", domainName, editRequest.Data)
	_, _, err = client.Domains.EditRecord(context.Background(), domainName, ipRecord.ID, &editRequest)
	if err != nil {
		fmt.Printf(err.Error())
	}
}
