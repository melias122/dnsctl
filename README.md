# dnsctl

A simple tool to automate updating of dynamic dns.

For know, only supports DigitalOcean as registrar.

to install: `go get github.com/Richard87/dnsctl`

`./dnsctl --hostname host.example.com --token <DIGITAL_OCEAN_TOKEN>`

 - `--4 to only update IPv4 records`
 - `--6 to only update IPv6 records`

It will automatically create,update and delete A and AAAA records and fetch IPv4 and IPv6 addresses from https://api.ident.me/
