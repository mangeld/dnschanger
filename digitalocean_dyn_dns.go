//Run with cron (or whatever you want to use), for example every 5 or 10 mins.
//Change the access token to the one that digitalocean gave you (Personal access token)
//In order to this to work you have to have created first the domain that you want to use
//(it won't create it, it only updates it)
package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/digitalocean/godo"
	"github.com/getsentry/sentry-go"
	"golang.org/x/oauth2"
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
	return token, nil
}

func initSentry(options sentry.ClientOptions, wrapped func() error) {
	err := sentry.Init(options)
	if err != nil {
		log.Fatalf("Error initialising sentry: %s", err)
	}
	defer sentry.Flush(2 * time.Second)
	defer sentry.Recover()

	err = wrapped()
	if err != nil {
		sentry.CaptureException(err)
	}
}

func requireEnv(name string) string {
	value, present := os.LookupEnv(name)
	if !present {
		log.Fatalf("Please provide %s as an env variable", name)
	}
	return value
}

func main() {
	// Get accessToken and domain from env
	accessToken := requireEnv("DIGITALOCEAN_ACCESS_TOKEN")
	domain := requireEnv("DOMAIN")
	tokenSource := &TokenSource{AccessToken: accessToken}
	sentryDSN, sentryEnabled := os.LookupEnv("SENTRY_DSN")
	if sentryEnabled {
		log.Print("Sentry DSN provided, initializing...")
		initSentry(
			sentry.ClientOptions{Dsn: sentryDSN},
			func() error {
				return changeDnsIp(tokenSource, domain)
			},
		)
	} else {
		if err := changeDnsIp(tokenSource, domain); err != nil {
			log.Fatal(err)
		}
	}
}

func getOwnIp() (string, error) {
	resp, err := http.Get("http://ipinfo.io/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	response := MyIp{}
	err = json.Unmarshal(data, &response)
	if err != nil {
		return "", err
	}
	return response.Ip, nil
}

func changeDnsIp(accessToken *TokenSource, domainName string) error {
	oauth_client := oauth2.NewClient(context.Background(), accessToken)
	client := godo.NewClient(oauth_client)
	listOps := godo.ListOptions{Page: 1, PerPage: 50}
	var err error = nil

	records, _, err := client.Domains.Records(context.Background(), domainName, &listOps)
	if err != nil {
		return err
	}
	var ipRecord = godo.DomainRecord{}
	for _, r := range records {
		if r.Type == "A" {
			ipRecord = r
		}
	}

	ownIp, err := getOwnIp()
	if err != nil {
		return err
	}

	if ipRecord.Data == ownIp {
		log.Printf("Public ip (%s) didn't change, doing nothing...", ipRecord.Data)
		return nil
	}

	editRequest := godo.DomainRecordEditRequest{Data: ownIp}
	log.Printf("Updating record %v to new ip: %v\n", domainName, editRequest.Data)
	_, _, err = client.Domains.EditRecord(context.Background(), domainName, ipRecord.ID, &editRequest)
	if err != nil {
		return err
	}
	return nil
}
