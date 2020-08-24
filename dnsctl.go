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
	ipv4 := flag.Bool("4", false, "Only update IPv4")
	ipv6 := flag.Bool("6", false, "Only update IPv6")
	token := flag.String("token", "", "AUTH-Token for digitalocean")
	flag.Parse()

	domainParts := strings.Split(*hostname, ".")
	if len(domainParts) <= 2 {
		panic(errors.New("hostname require at least 3 parts"))
	}

	client := godo.NewFromToken(*token)
	ds := do.NewDomainsService(client)

	ipv4address := getIpV4Address()
	ipv6address := getIpV6Address()

	domain := fmt.Sprintf("%s.%s", domainParts[len(domainParts)-2], domainParts[len(domainParts)-1])
	host := strings.Join(domainParts[:len(domainParts)-2], ".")
	fmt.Printf("Finding records for %s\n", host)
	records, err := ds.Records(domain)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Updating %s\n", *hostname)

	recordRequestv4 := do.DomainRecordEditRequest{
		Type: "A",
		Name: host,
		Data: ipv4address,
		TTL:  60,
	}
	recordRequestv6 := do.DomainRecordEditRequest{
		Type: "AAAA",
		Name: host,
		Data: ipv6address,
		TTL:  60,
	}
	if *ipv4 {

		record := findRecord(records, host, "A")
		if record == nil {
			fmt.Printf("Creating new IPv4: %s\n", ipv4address)
			_, err := ds.CreateRecord(domain, &recordRequestv4)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("Updating only IPv4: %s\n", ipv4address)
			_, err := ds.EditRecord(domain, record.ID, &recordRequestv4)
			if err != nil {
				panic(err)
			}
		}
	} else if *ipv6 {

		record := findRecord(records, host, "AAAA")
		if record == nil {
			fmt.Printf("Creating new IPv6: %s\n", ipv6address)
			_, err := ds.CreateRecord(domain, &recordRequestv6)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("Updating only IPv6: %s\n", ipv6address)
			_, err := ds.EditRecord(domain, record.ID, &recordRequestv6)
			if err != nil {
				panic(err)
			}
		}
	} else {
		fmt.Printf("Updating IPv4 and IPv6: %s / %s\n", ipv4address, ipv6address)

		record := findRecord(records, host, "AAAA")
		if record == nil {
			fmt.Printf("Creating new IPv6: %s\n", ipv6address)
			_, err := ds.CreateRecord(domain, &recordRequestv6)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("Updating only IPv6: %s\n", ipv6address)
			_, err := ds.EditRecord(domain, record.ID, &recordRequestv6)
			if err != nil {
				panic(err)
			}
		}

		record = findRecord(records, host, "A")
		if record == nil {
			fmt.Printf("Creating new IPv4: %s\n", ipv4address)
			_, err := ds.CreateRecord(domain, &recordRequestv4)
			if err != nil {
				panic(err)
			}
		} else {
			fmt.Printf("Updating only IPv4: %s\n", ipv4address)
			_, err := ds.EditRecord(domain, record.ID, &recordRequestv4)
			if err != nil {
				panic(err)
			}
		}
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

func getIpV4Address() string {
	response, err := http.Get("https://v4.ident.me/")
	if err != nil {
		panic(err)
	}
	address, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}
	return string(address)
}

func getIpV6Address() string {
	response, err := http.Get("https://v6.ident.me/")
	if err != nil {
		panic(err)
	}
	address, err := ioutil.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		panic(err)
	}
	return string(address)
}
