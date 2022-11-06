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

### UDP

### HTTP

#### Endpoints

_________________


## Factories

_______________


## CLI

### Config

#### CLI parameters

#### OS environment variables

#### From file

_______________


## Docker

### Docker-compose