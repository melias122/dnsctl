package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/digitalocean/doctl/do"
	"github.com/digitalocean/godo"
)

func parseHostname(hostname string) (string, string, error) {
	p := strings.Split(hostname, ".")
	if len(p) < 2 {
		return "", "", fmt.Errorf("invalid hostname %s", hostname)
	}
	subdomain := strings.Join(p[:len(p)-2], ".")
	domain := strings.Join(p[len(p)-2:], ".")
	return subdomain, domain, nil
}

func findRecord(records []do.DomainRecord, hostname string, recordType string) *do.DomainRecord {
	for i, n := range records {
		if hostname == n.Name && recordType == n.Type {
			return &records[i]
		}
	}
	return nil
}

func myip(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	addr, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	return string(addr), nil
}

func run(ds do.DomainsService, hostname string, typ string, forceIpAddr string) error {
	subdomain, domain, err := parseHostname(hostname)
	if err != nil {
		return err
	}

	// if there is no subdomain, update domain itself
	if subdomain == "" {
		subdomain = "@"
	}

	ipAddr := forceIpAddr
	if ipAddr == "" {
		switch typ {
		case "A":
			ipv4, err := myip("https://v4.ident.me")
			if err != nil {
				return err
			}
			ipAddr = ipv4
		case "AAAA":
			ipv6, err := myip("https://v6.ident.me")
			if err != nil {
				return err
			}
			ipAddr = ipv6
		default:
			return fmt.Errorf("unknown typ %s", typ)
		}
	}

	request := do.DomainRecordEditRequest{
		Type: typ,
		Name: subdomain,
		Data: ipAddr,
		TTL:  3600,
	}

	records, err := ds.Records(domain)
	if err != nil {
		return err
	}

	record := findRecord(records, subdomain, "A")
	if record == nil {
		if ipAddr == "" {
			return fmt.Errorf("ipv4 is empty cannot create record")
		}

		log.Printf("creating new record:%s %s.%s %s\n", request.Type, request.Name, domain, ipAddr)
		_, err := ds.CreateRecord(domain, &request)
		return err
	} else if ipAddr == "" {
		log.Printf("deleting outdated record:%s %s.%s %s\n", record.Type, record.Name, domain, record.Data)
		return ds.DeleteRecord(domain, record.ID)
	} else if record.Data != ipAddr {
		log.Printf("updating record:%s %s.%s %s->%s\n", record.Type, record.Name, domain, record.Data, ipAddr)
		_, err := ds.EditRecord(domain, record.ID, &request)
		return err
	}

	return nil
}

func main() {
	log.SetFlags(0)

	var (
		token       = flag.String("token", "", "Digitalocean auth token")
		no4         = flag.Bool("no4", false, "Do not update A record")
		no6         = flag.Bool("no6", false, "Do not update AAAA record")
		hostname    = flag.String("hostname", "example.com", "Hostname to update. It could have subdomain (sub.example.com)")
		forceIPAddr = flag.String("forceIPAddr", "", "Force IP address override")
	)
	flag.Parse()

	if *token == "" {
		log.Fatal("token is not set")
	}

	client := godo.NewFromToken(*token)
	ds := do.NewDomainsService(client)

	if !*no4 {
		if err := run(ds, *hostname, "A", *forceIPAddr); err != nil {
			log.Fatal(err)
		}
	}
	if !*no6 {
		if err := run(ds, *hostname, "AAAA", *forceIPAddr); err != nil {
			log.Fatal(err)
		}
	}
}
