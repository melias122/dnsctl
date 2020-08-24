# dnsctl

A simple tool to automate updating of dynamic dns

For know, only supports DigitalOcean as registrar

`./dnsctl --hostname host.example.com --token <DIGITAL_OCEAN_TOKEN>`

 - `--4 to only update IPv4 records`
 - `--6 to only update IPv6 records`

It will automatically create A and AAAA records (if you have public ipv4 and/or ipv6 addresses)
