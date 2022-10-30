# Cartographer
Cartographer is a tool intended to be used for internal audits to scan hosts and generate a CSV file with all results. The core principle is to get the server list from a domain controller, resolve IP addresses and perform scans on hosts. This avoids having to scan large IP ranges to discover servers. This list can be enriched with an additional network scan for servers outside of the domain on IP ranges. The tool will perform a port scan on all resolved hosts and additional modules for Windows servers in the domain are available: 
- List shares exposed
- List sessions of domain administrators (works only in a shell launched with runas for now)
- Check if Webdav is enabled
- List RPC interfaces (to identify print spooler for example)

## Installation
Make sur Go is installed.
```bash
git clone https://github.com/hugopicq/cartographergo.git
cd cartographergo
go get
go build
```
At this point you should have cartographergo.exe in the directory.

## Manual

### Cheatsheet

##### Basic Usage
This command will ask the DC for all servers, resolve IP addresses, run port scan on all servers and run all modules on the servers.
```text
.\cartographergo -s <DOMAIN_CONTROLLER> -u <USERNAME> -p <PASSWORD> -d <DOMAIN> -o output.csv
```

##### Restrictions
This command will ask the DC for all servers, resolve IP addresses, keep only servers which IP addresses are in whitelist.txt and not in blacklist.txt, run port scan on these servers and run RPC and Webdav modules.
```text
.\cartographergo -s <DOMAIN_CONTROLLER> -u <USERNAME> -p <PASSWORD> -d <DOMAIN> -o output.csv --whitelist-file whitelist.txt --blacklist-file blacklist.txt --modules rpc,webdav
```
##### Complete scan
This command will do the same as the first one, but will also perform port scans on all hosts in the specified additional IP ranges which have not been discovered in the initial domain scan. It won't run modules on the additional hosts.
```text
.\cartographergo -s <DOMAIN_CONTROLLER> -u <USERNAME> -p <PASSWORD> -d <DOMAIN> -o output.csv -a 192.168.230.0/24,192.168.10.0/24
```

##### Port scan only without AD reconnaissance
This command will simply perform a basic port scan on the provided IP range.
```text
.\cartographergo -a 192.168.0.0/16 --skip-domain -o output.csv
```

##### Port scan and RPC dump of a specific server
This command will perform a port scan on the specified IP address and dump RPC interfaces on this host.
```text
.\cartographergo -s <DOMAIN_CONTROLLER> -u <USERNAME> -p <PASSWORD> -d <DOMAIN> -o output.csv -a 192.168.0.5 --skip-domain --run-modules-additional --modules rpc
```

### Options

```text
USAGE : 
  cartographer [flags]

Flags:
  -a, --additional-IP string      Additional IP to scan. They must be in CIDR format separated by commas
  -b, --batchsize uint16          Batch size (default 4500)
      --blacklist-file string     Blacklist IP files in IP or CIDR format (1 line per CIDR)
  -d, --domain string             Active Directory Domain
  -s, --domaincontroller string   IP of Domain Controller
  -h, --help                      help for cartographer
      --include-workstations      Include workstations for scans, by default only computers matching *Server* in the OS will be scanned
  -m, --modules string            Modules to run on resolved domain computers separated by commas (all by default). Available: all, adminsessions, webdav, shares, rpc (default "all")
  -o, --outputfile string         Output filepath
  -p, --password string           Active Directory Password
      --run-modules-additional    Run chosen modules on additional IPs
      --skip-domain               Skip AD reconnaissance with modules to go straight to additional IP port scan
  -t, --timeout uint              Timeout in milliseconds (default 1500)
  -u, --user string               Active Directory User
      --whitelist-file string     Whitelist IP files in IP or CIDR format (1 line per CIDR)
```
