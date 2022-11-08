# dns

*A basic but modular domain name server*

___________________


## Overview

This project explores domain-driven design (DDD) in Go, with a DNS application.

The application will contain a UDP transport (for anwering DNS questions) as well as a HTTP transport with endpoints for various purposes (CRUD operations against the DNS records store, enabling / disabling the UDP server, a health-check and status endpoint).

This allows working with a service layer composed of two (basic) elements, using a (records) store repository and a(n answering) DNS repository. The implmentations available are very simple, with the DNS server being based on [`github.com/miekg/dns`](https://github.com/miekg/dns) and the records store a simple (in-memory) Go map, optionally wrapped with a writer that stores this data to a file. These implementations satisfy said store repository, DDD style.

To simplify spawning these components, there is also a factory implementation, to simplify some of the initialization process; as well as strong support for CLI / container env / file-based configuration, making it seamless to spawn a new instance of the app and keep it configured as such.

______________

## Entities

### [Record](./store/record.go#L4)

A Record will contain information about a DNS record, holding its record type, domain name and IP address.

```go
type Record struct {
	Type string `json:"type,omitempty"`
	Name string `json:"name,omitempty"`
	Addr string `json:"address,omitempty"`
}
```

### [RecordWithTarget](./store/record.go#L10)

A RecordWithTarget will wrap a Record with a target domain name, used for updating a certain record.

```go
type RecordWithTarget struct {
	Record `json:"record,omitempty"`
	Target string `json:"target,omitempty"`
}
```

____________________

## Repositories

### [Store Repository](./store/repository.go#L12)

A (DNS records) store repository defines the methods for accessing, registering and changing DNS records. 

It exposes basic create, read, update, delete (CRUD) operations, as well as specific filter methods to satisfy all needed queries.

```go
type Repository interface {
	Create(context.Context, ...*Record) error
	List(context.Context) ([]*Record, error)
	FilterByTypeAndDomain(context.Context, string, string) (*Record, error)
	FilterByDomain(context.Context, string) ([]*Record, error)
	FilterByDest(context.Context, string) ([]*Record, error)
	Update(context.Context, string, *Record) error
	Delete(context.Context, *Record) error
}
```

#### Implementations

##### [In-Memory Map (`memmap`)](./store/memmap/memmap.go#L18) 

A basic implementation of a store repository with a Go map grouping a set of record types to domain names to IP addresses. The access to the data is protected with a `sync.RWMutex`.

The reason for the order of the elements in the map (record types > domain names > IP addresses) is to favor DNS queries, that will ask for a certain record type and domain name. This is the most effective way to group this data for these kinds of queries; while sacrificing write operations with longer times. 

```go
type MemoryStore struct {
	// maps a set of record types to domain names to IPs
	Records map[string]map[string]string
	mtx     sync.RWMutex
}
```

##### [File (`file`)](./store/file/file.go#L20)

A wrapper for `memmap`, this implementation will flush all records to a file in either JSON or YAML format, when any change is done.

As the struct implies, it will simply leverage the `memmap` implementation and call its methods (while writing the store contents to a file on any type of mutation).

```go
type FileStore struct {
	Path  string `json:"path,omitempty" yaml:"path,omitempty"`
	store store.Repository
	enc   encoder.EncodeDecoder
	mtx   sync.RWMutex
}
```

### [DNS Repository](./dns/repository.go#L12)

A DNS (answering service) repository will define the methods for replying to DNS questions for both stored domains as well as to fallback to a secondary DNS in case no records are found for a certain domain.

For the moment, as there are no other DNS implementations (the server logic), it strictly follows a model based on [`miekg/dns`](https://github.com/miekg/dns)

```go
type Repository interface {
	Answer(*store.Record, *dns.Msg)
	Fallback(*store.Record, *dns.Msg)
}
```

#### Implementations

##### [Core - ](./dns/core/core.go#L28)[`miekg/dns`](https://github.com/miekg/dns)[ (`core`)](./dns/core/core.go#L28)

While its Answer method will simply pass the record type, domain name and IP address from the input `*store.Record` into the input `*dns.Msg.Answer` as a `*dns.RR`; the repository also handles a fallback scenario where the record is not found in the record store (for instance).

That is where its Fallback method kicks in, spawning a DNS client to forward the same question to each of the configured fallback DNS, until a valid answer is retrieved. Then, it is appended to `*dns.Msg.Answer` as in the Answer method, and the function ends.

```go
type DNSCore struct {
	fallbackDNS []string
}
```

### [Health Repository](./health/repository.go#L14)

A Health repository will define methods for a health-check / status report on the application's running services.

This involves basic tests against the services to provide the user with a summary of the current status of the application.

```go
type Repository interface {
	Store(int, time.Duration) *StoreReport
	DNS(address string, fallback string, records *store.Record) *DNSReport
	HTTP(port int) *HTTPReport
	Merge(*StoreReport, *DNSReport, *HTTPReport) *Report
}
```

#### Implementations

##### [Simple-health (`simplehealth`)](./health/simplehealth/simple.go#L)

While the service layer actually performs the calls to feed its methods, this implementation will generate a quick and simple report (with sane defaults) providing context on the status of the application.

Each of the probed services will be awarded a `health.Status` value, which will be determined from the input metadata.

When all three reports are generated, they can be fed into the `Merge` method which will merge all the data and also determining an overall status of the app.

```go
type shealth struct{}
```
__________


## [Service](./service/service.go#L28)

The service layer is what glues the different repositories for the different services, and allows them to interact with eachother.

Its interfaces are sharded into different scopes, with a general `Service` interface containing all functionalities. This allows contiguring (for instance), transport elements one level above with a more limited list of method it can access. This is evident in the UDP transport, where the UDP server only contains a `service.Answering` interface as one of its elements.

A service instance is spawned with a DNS Repository, a Store Repository, a Health Repository and a Config. 

```go
type Service interface {
	StoreService
	DNSService
	HealthService
}

type StoreService interface {
	AddRecord(ctx context.Context, r *store.Record) error
	AddRecords(ctx context.Context, rs ...*store.Record) error
	ListRecords(ctx context.Context) ([]*store.Record, error)
	GetRecordByTypeAndDomain(ctx context.Context, rtype, domain string) (*store.Record, error)
	GetRecordByAddress(ctx context.Context, address string) ([]*store.Record, error)
	UpdateRecord(ctx context.Context, domain string, r *store.Record) error
	DeleteRecord(ctx context.Context, r *store.Record) error
}

type DNSService interface {
	AnswerDNS(r *store.Record, m *dnsr.Msg)
}

type HealthService interface {
	StoreHealth() *health.StoreReport
	DNSHealth() *health.DNSReport
	HTTPHealth() *health.HTTPReport
	Health() *health.Report
}

type StoreWithHealth interface {
	StoreService
	HealthService
}

type Answering interface {
	GetRecordByTypeAndDomain(context.Context, string, string) (*store.Record, error)
	AnswerDNS(*store.Record, *dnsr.Msg)
}
```

### [Middleware](./service/middleware)

The service layer exposes middleware too, which are none other than wrappers for the Service interface, to perform a certain set of actions before or after (or both) to Service method calls.

For the moment, a logger middleware [is available in its own folder](./service/middleware/logger/logger.go#L18)


_______________

## Transports

The app works with two transport types, UDP for answering DNS questions and HTTP to expose certain endpoints, providing users with controls over the DNS records store, the DNS server, and health checks.

### [UDP](./transport/udp/server.go#L12)

The UDP transport will listen on DNS queries, while interacting with the service, with its `service.Answering` interface.

```go
type Server interface {
	Start() error
	Stop() error
	Running() bool
}
```

While there is only one implementation of `udp.Server` with [miekg/dns](https://github.com/miekg/dns), there is *room* to expand the app with a new implementation of the server.

#### Implementations

##### [`miekgdns`](./transport/udp/miekgdns/server.go#L10)

This implementation leverages the [`miekg/dns`](https://github.com/miekg/dns) library to serve as a DNS server. It's also configured with a `service.Answering` interface to interact with the DNS records store.

```go
type udps struct {
	on   bool
	ans  service.Answering
	conf *udp.DNS
	srv  *dns.Server
	err  error
}
```

### [HTTP](./transport/httpapi/server.go#L14)

HTTP will expose endpoints to provide users with access to the DNS records store, the DNS server and health-checks. 

```go
type Server interface {
	Start() error
	Stop() error
}

type HTTPAPI interface {
	StartDNS(w http.ResponseWriter, r *http.Request)
	StopDNS(w http.ResponseWriter, r *http.Request)
	ReloadDNS(w http.ResponseWriter, r *http.Request)

	AddRecord(w http.ResponseWriter, r *http.Request)
	ListRecords(w http.ResponseWriter, r *http.Request)
	GetRecordByDomain(w http.ResponseWriter, r *http.Request)
	GetRecordByAddress(w http.ResponseWriter, r *http.Request)
	UpdateRecord(w http.ResponseWriter, r *http.Request)
	DeleteRecord(w http.ResponseWriter, r *http.Request)

	Health(w http.ResponseWriter, r *http.Request)
}
```

#### Endpoints

The endpoints are configured with a standard-library HTTP server and muxer. Similarly, there isn't much complexity with the endpoints themselves, mostly based on GET / POST HTTP requests (without actually specifying any data in the URL path or parameters, at most as POST data).

The endpoints can be implemented with any HTTP library, provided they can satisfy the `HTTPAPI` interface (and the endpoints themselves are accessible).

Below is a list of all endpoints and their characteristics:

Endpoint | Method | Action | Description | Post Data
:-------:|:------:|:------:|:-----------:|:---------:
`/dns/start` | `GET` | [`StartDNS`](./transport/httpapi/endpoints/dns.go#L9) | Starts the DNS server | N/A 
`/dns/stop` | `GET` | [`StopDNS`](./transport/httpapi/endpoints/dns.go#L34) | Stops the DNS server | N/A 
`/dns/reload` | `GET` | [`ReloadDNS`](./transport/httpapi/endpoints/dns.go#L54) | Stops and then starts the DNS server | N/A 
`/records/add` | `POST` | [`AddRecord`](./transport/httpapi/endpoints/store.go#L12) | Adds a new entry to the DNS records store | `{"name":"not.a.dom.ain","type":"A","address":"192.168.0.10"}`
`/records` | `GET` | [`ListRecords`](./transport/httpapi/endpoints/store.go#L72) | Lists all DNS records in the store | N/A
`/records/getAddress` | `POST` | [`GetRecordByDomain`](./transport/httpapi/endpoints/store.go#L100) | Gets the IP Address of a record, filtered by domain name and by record type | `{"name":"not.a.dom.ain","type":"A"}`
`/records/getDomains` | `POST` | [`GetRecordByAddress`](./transport/httpapi/endpoints/store.go#L149) | Gets a list of record types and associated domains, filtered by IP address  | `{"address":"192.168.0.10"}`
`/records/update` | `POST` | [`UpdateRecord`](./transport/httpapi/endpoints/store.go#L202) | Updates a certain record by targetting its domain name | `{"target":"not.a.dom.ain","record":{"name":"really.not.a.dom.ain","type":"A","address":"192.168.0.10"}}`
`/records/delete` | `POST` | [`DeleteRecord`](./transport/httpapi/endpoints/store.go#L261) | Removes a record from the store, by targetting its domain name and record type | `{"name":"really.not.a.dom.ain","type":"A"}`
`/health` | `GET` | [`DeleteRecord`](./transport/httpapi/endpoints/health.go#L9) | Generates a health-check / status report on the app's services | N/A

_________________


## Factories

To make it easier to spawn each of these, be it a repository, service or anything else, there is a `factory` package available.

This package will simply do all the *manual* configuration work and spit out what you really need:

```go
func StoreRepository(rtype string, path string) store.Repository
func DNSRepository(rtype string, fallbackDNS ...string) dns.Repository
func HealthRepository(rtype string) health.Repository
func Service(dnsRepo dns.Repository, storeRepo store.Repository, healthRepo health.Repository, conf *config.Config) service.Service
func UDPServer(stype, address, prefix, proto string, svc service.Service) udp.Server 
func Server(dnstype, dnsAddress, dnsPrefix, dnsProto string, httpPort int, svc service.Service) (httpapi.Server, udp.Server) 
func From(conf *config.Config) httpapi.Server
```

While most of these are granular enough to *compose* your configuration along the way, it is worth underlining that the most streamlined option is to leverage the `From(*config.Config) httpapi.Server` function, and in one-shot set up the app (from a configuration, even if default).

_______________


## CLI

The command-line interface for this app is set up in `cmd`; who will also manage the configuration structures for the app.

### Config

For each service or major feature, there will be a dedicated data structure used to configure it. 

Every single configuration struct will need to satisfy the `ConfigOption` interface, which only contains one method which applies said configuration to a pointer to a Config:

```go
type ConfigOption interface {
	Apply(*Config)
}
```

The Config itself will be composed of many modules as pointed out above:

```go
type Config struct {
	DNS       *DNSConfig       `json:"dns,omitempty" yaml:"dns,omitempty"`
	Store     *StoreConfig     `json:"store,omitempty" yaml:"store,omitempty"`
	HTTP      *HTTPConfig      `json:"http,omitempty" yaml:"http,omitempty"`
	Logger    *LoggerConfig    `json:"logger,omitempty" yaml:"logger,omitempty"`
	Autostart *AutostartConfig `json:"autostart,omitempty" yaml:"autostart,omitempty"`
	Health    *HealthConfig    `json:"health,omitempty" yaml:"health,omitempty"`
	Type      string           `json:"type,omitempty" yaml:"type,omitempty"`
	Path      string           `json:"path,omitempty" yaml:"path,omitempty"`
}

type DNSConfig struct {
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`
	FallbackDNS string `json:"fallback,omitempty" yaml:"fallback,omitempty"`
	Address     string `json:"address,omitempty" yaml:"address,omitempty"`
	Prefix      string `json:"prefix,omitempty" yaml:"prefix,omitempty"`
	Proto       string `json:"proto,omitempty" yaml:"proto,omitempty"`
}

type StoreConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
}

type HTTPConfig struct {
	Port int `json:"port,omitempty" yaml:"port,omitempty"`
}

type LoggerConfig struct {
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

type AutostartConfig struct {
	DNS bool `json:"dns,omitempty" yaml:"dns,omitempty"`
}

type HealthConfig struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}
```

As it is ensured that using the `config.New()` or `config.Default()` functions will correctly initialize a Config (even ready for usage, if you wanted an in-memory store), the configuration struct will simply work with the target field. Any validation for the input is done on a separate function. Take as an example `config.HTTPPort(p int) ConfigOption`:

```go
func HTTPPort(p int) ConfigOption {
	if p > 65535 || p == 0 {
		return nil
	}
	return &httpPort{
		p: p,
	}
}

type httpPort struct {
	p int
}

// Apply implements the ConfigOption interface
func (h *httpPort) Apply(c *Config) {
	c.HTTP.Port = h.p
}
```

#### CLI parameters

Below is a list of all CLI parameters (flags) you can pass when starting the app:

Flag | Type | Default | Description
:---:|:----:|:-------:|:-----------:
`-dns-addr` | `string` | `:53` | the address to listen to for DNS queries
`-dns-fallback` | `string` |  | use a secondary DNS to parse unsuccessful queries
`-dns-prefix` | `string` | `.` | the prefix for DNS queries / answers. Usually it's a period (.) 
`-dns-proto` | `string` | `udp` | the protocol for the DNS server
`-dns-type` | `string` | `miekgdns` | use a specific domain-name server implementation 
`-file` | `string` |  | load a config from a file
`-health-type` | `string` | `simplehealth` | the type of health / status report 
`-http-port` | `int` | `8080` | port to use for the HTTP API, defaults to :8080
`-log-path` | `string` |  | the log file's path, to register events
`-log-type` | `string` | `text` | the type of formatter to use for the logger (text, json, yaml)
`-start-dns` | `bool` | `true` | automatically start the DNS server
`-store-path` | `string` |  | the record store file path, if stored to a file
`-store-type` |`string` | `memmap` | the record store implementation to use (memmap, yamlfile, jsonfile)

#### OS environment variables

Below is a list of all OS environment variables you can set before starting the app:

Variable name | Type | Description
:------------:|:----:|:-----------:
`DNS_ADDRESS` | `string` | the address to listen to for DNS queries
`DNS_FALLBACK` | `string` | use a secondary DNS to parse unsuccessful queries
`DNS_PREFIX` | `string`  | the prefix for DNS queries / answers. Usually it's a period (.) 
`DNS_PROTO` | `string`  | the protocol for the DNS server
`DNS_TYPE` | `string`  | use a specific domain-name server implementation 
`DNS_CONFIG_PATH` | `string`  | load a config from a file
`DNS_HEALTH_TYPE` | `string`  | the type of health / status report 
`DNS_API_PORT` | `int`  | port to use for the HTTP API, defaults to :8080
`DNS_LOGGER_PATH` | `string`  | the log file's path, to register events
`DNS_LOGGER_TYPE` | `string`  | the type of formatter to use for the logger (text, json, yaml)
`DNS_AUTOSTART` | `string`  | automatically start the DNS server
`DNS_STORE_PATH` | `string` | the record store file path, if stored to a file
`DNS_STORE_TYPE` |`string` | the record store implementation to use (memmap, yamlfile, jsonfile)

#### From file

Below is the content of an example configuration file, in YAML format:

```yaml
dns:
  type: miekgdns
  fallback: 1.1.1.1
  address: :53
  prefix: .
  proto: udp
store:
  type: yamlfile
  path: /tmp/dns/dns.list
http:
  port: 8080
logger:
  type: text
  path: /tmp/dns/dns.log
autostart:
  dns: true
health:
  type: simplehealth
type: yaml
path: /tmp/dns/dns.conf
```
_______________

## Build / Test


### Go

**Building** - From the root of the repository, run:

```shell
go build -o dns .
```

This generates a binary you can execute directly.

**Testing** - From the root of the repository, run:

```shell
go test -v -timeout=0 ./...
```

### Bazel

**Building** - From the root of the repository, run:

```shell
bazel build //...
```

This builds a binary which you can use with Bazel (run, test, etc).

**Testing** - From the root of the repository, run:

```shell
bazel test //...
```


### Docker

The app can be deployed to a container easily via the [Dockerfile](./Dockerfile) in the repository's root directory.

The Dockerfile will perform a multi-stage build with `golang:alpine` fetching the dependencies and building the binary -- which is then copied to the final `alpine:edge` container.

**Building** - From the root of the repository, run:

```shell
docker build -t dns:local .
```

### Docker-compose

To deploy the app (and also build+deploy) you can use the [`docker-compose.yaml` file](./docker-compose.yaml) where you can launch the app with a certain configuration (and also in an isolated container).

While the default file configures the container with a `network_mode: host` setting, a setup that fits neatly in a home-based DNS deployment, you may prefer to set it up for an isolated network of containers -- for that you can comment-out the `network_mode: host` line and uncomment the `privileged` and `ports` elements.

**Executing** - From the root of the repository, run:

```shell
docker compose up -d dns
```