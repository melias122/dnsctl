package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
	"io/ioutil"
	"net/http"
	"strings"
)

func main() {
	// hostname, ipv4 or ipv6 or both
	hostname := flag.String("hostname", "example.com", "Which hostname to update")
	ignoreIpv6 := flag.Bool("4", false, "Only update IPv4")
	ignoreIpv4 := flag.Bool("6", false, "Only update IPv6")
	token := flag.String("token", "", "AUTH-Token for digitalocean")
	flag.Parse()

	domainParts := strings.Split(*hostname, ".")
	if len(domainParts) <= 2 {
		panic(errors.New("hostname require at least 3 parts"))
	}

	client := godo.NewFromToken(*token)
	ds := do.NewDomainsService(client)

	domain := fmt.Sprintf("%s.%s", domainParts[len(domainParts)-2], domainParts[len(domainParts)-1])
	host := strings.Join(domainParts[:len(domainParts)-2], ".")
	fmt.Printf("Fetching existing records for %s\n", *hostname)
	records, err := ds.Records(domain)
	if err != nil {
		panic(err)
	}

	if *ignoreIpv6 == false {
		err := handleIpV6(records, host, ds, domain)
		if err != nil {
			panic(err)
		}
	}

	if *ignoreIpv4 == false {
		err := handleIpV4(records, host, ds, domain)
		if err != nil {
			panic(err)
		}
	}
}

func handleIpV4(records do.DomainRecords, host string, ds do.DomainsService, domain string) error {
	ipv4address := getContent("https://v4.ident.me/")
	if ipv4address != "" {
		fmt.Printf("Found public IPv4: %s\n", ipv4address)
	}

	record := findRecord(records, host, "A")
	request := do.DomainRecordEditRequest{
		Type: "A",
		Name: host,
		Data: ipv4address,
		TTL:  60,
	}

	if ipv4address == "" {
		if record != nil {
			fmt.Printf("Deleting outdated IPv4 record: %s\n", record.Data)
			err := ds.DeleteRecord(domain, record.ID)
			return err
		}
		fmt.Printf("No IPv4 record or address\n")
		return nil
	}

	if record == nil {
		fmt.Printf("Creating new IPv4: %s\n", ipv4address)
		_, err := ds.CreateRecord(domain, &request)
		return err
	}

	if record.Data != ipv4address {
		fmt.Printf("Updating existing IPv4: %s\n", ipv4address)
		_, err := ds.EditRecord(domain, record.ID, &request)
		return err
	} else {
		fmt.Printf("No changes for IPv4\n")
		return nil
	}
}

func handleIpV6(records do.DomainRecords, host string, ds do.DomainsService, domain string) error {

	ipv6address := getContent("https://v6.ident.me/")
	if ipv6address != "" {
		fmt.Printf("Found public IPv6: %s\n", ipv6address)
	}
	record := findRecord(records, host, "AAAA")
	request := do.DomainRecordEditRequest{
		Type: "AAAA",
		Name: host,
		Data: ipv6address,
		TTL:  60,
	}

	if ipv6address == "" {
		if record != nil {
			fmt.Printf("Deleting outdated IPv6 record: %s\n", record.Data)
			err := ds.DeleteRecord(domain, record.ID)
			return err
		}
		fmt.Printf("No IPv6 record or address\n")
		return nil
	}

	if record == nil {
		fmt.Printf("Creating new IPv6: %s\n", ipv6address)
		_, err := ds.CreateRecord(domain, &request)
		return err
	}

	if record.Data != ipv6address {
		fmt.Printf("Updating existing IPv6: %s\n", ipv6address)
		_, err := ds.EditRecord(domain, record.ID, &request)
		return err
	} else {
		fmt.Printf("No changes for IPv6\n")
		return nil
	}
}

func findRecord(records []do.DomainRecord, hostname string, recordType string) *do.DomainRecord {
	for i, n := range records {
		if hostname == n.Name && recordType == n.Type {
			return &records[i]
		}
	}
	return nil
}

func getContent(url string) string {
	response, err := http.Get(url)
	if err != nil {
		return ""
	}
	address, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return ""
	}
	return string(address)
}
