// Package e2e will expect a running instance of the app, where this
// test file will hit the supported endpoints to ensure they are working
// as intended
//
// to run the tests, run the app via Docker or
// `sudo go run examples/core_memmap/main.go`
package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/miekg/dns"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/transport/httpapi"
)

type dnsContainer struct {
	testcontainers.Container
	HTTPURI string
	UDPURI  string
}

func initService() (*dnsContainer, error) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:    "./../..",
			Dockerfile: "Dockerfile",
		},
		Name: "dns",
		ExposedPorts: []string{
			"8080/tcp",
			"53/udp",
		},
		Privileged: true,
		Env: map[string]string{
			"DNS_AUTOSTART": "0",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	ip, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	httpPort, err := container.MappedPort(ctx, "8080/tcp")
	if err != nil {
		return nil, err
	}
	udpPort, err := container.MappedPort(ctx, "53/udp")
	if err != nil {
		return nil, err
	}

	httpURI := fmt.Sprintf("http://%s:%s", ip, httpPort.Port())
	udpURI := fmt.Sprintf("%s:%s", ip, udpPort.Port())

	return &dnsContainer{
		Container: container,
		HTTPURI:   httpURI,
		UDPURI:    udpURI,
	}, nil
}

func httpReq(targetURI, endpoint string, data []byte) ([]byte, int, error) {
	var (
		res *http.Response
		err error
	)

	switch data {
	case nil:
		res, err = http.Get(fmt.Sprintf("%s%s", targetURI, endpoint))
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, 0, err
		}
	default:
		buf := &bytes.Buffer{}
		_, err := buf.Write(data)
		if err != nil {
			return nil, 0, err
		}
		res, err = http.Post(fmt.Sprintf("%s%s", targetURI, endpoint), "application/json", buf)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, 0, err
		}
	}

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, 0, err
	}
	return b, res.StatusCode, nil
}

func dnsQuestion(targetURI, rtype, domain string) *dns.Msg {
	message := new(dns.Msg)
	out := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(domain), store.RecordTypeInts[rtype])
	client := &dns.Client{
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		Net:          "udp",
	}
	in, _, err := client.Exchange(message, targetURI)
	if err != nil || len(in.Answer) == 0 {
		return message
	}
	out.Answer = append(out.Answer, in.Answer...)

	return out
}

func TestTransport(t *testing.T) {
	dnsC, err := initService()
	if err != nil {
		t.Errorf("unexpected error starting test container: %v", err)
	}
	defer func() {
		err := dnsC.Terminate(context.Background())
		if err != nil {
			t.Errorf("failed to terminate test container: %v", err)
			return
		}
	}()

	time.Sleep(time.Second * 2)
	t.Run("HTTP", func(t *testing.T) {
		t.Run("DNS", func(t *testing.T) {
			t.Run("StartDNS", func(t *testing.T) {
				b, status, err := httpReq(dnsC.HTTPURI, "/dns/start", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}

				t.Log("[ok] /dns/start")
			})
			t.Run("ReloadDNS", func(t *testing.T) {
				b, status, err := httpReq(dnsC.HTTPURI, "/dns/reload", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}

				t.Log("[ok] /dns/reload")
			})
		})
		t.Run("Store", func(t *testing.T) {
			t.Run("AddRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"added record successfully","record":{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records/add", []byte(`{"name":"not.a.dom.ain","type":"A","address":"192.168.0.10"}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records/add")
			})
			t.Run("ListRecords", func(t *testing.T) {
				wants := `{"success":true,"message":"listing all records","records":[{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}]}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records")
			})
			t.Run("GetRecordByDomainAndType", func(t *testing.T) {
				wants := `{"success":true,"message":"fetched record for domain not.a.dom.ain","record":{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records/getAddress", []byte(`{"name":"not.a.dom.ain","type":"A"}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records/getAddress")
			})
			t.Run("GetRecordByAddress", func(t *testing.T) {
				wants := `{"success":true,"message":"listing all records for IP address 192.168.0.10","records":[{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}]}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records/getDomains", []byte(`{"address":"192.168.0.10"}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records/getDomains")
			})
			t.Run("UpdateRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"updated record successfully","record":{"type":"A","name":"really.not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records/update", []byte(`{"target":"not.a.dom.ain","record":{"name":"really.not.a.dom.ain","type":"A","address":"192.168.0.10"}}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records/update")
			})
		})
	})
	t.Run("UDP", func(t *testing.T) {
		t.Run("LocalQuery", func(t *testing.T) {
			m := dnsQuestion(dnsC.UDPURI, "A", "really.not.a.dom.ain")

			if m == nil {
				t.Errorf("unexpected nil DNS message")
				return
			}
			if len(m.Answer) != 1 {
				t.Errorf("unexpected answer length: wanted %v ; got %v", 1, len(m.Answer))
				return
			}

			if !strings.Contains(m.Answer[0].String(), "really.not.a.dom.ain") {
				t.Errorf("missing expected domain name: %s", m.Answer[0].String())
			}
			if !strings.Contains(m.Answer[0].String(), "A") {
				t.Errorf("missing expected record type: %s", m.Answer[0].String())
			}
			if !strings.Contains(m.Answer[0].String(), "192.168.0.10") {
				t.Errorf("missing expected IP address: %s", m.Answer[0].String())
			}

			t.Log("[ok] local DNS query")
		})
		t.Run("ExternalQuery", func(t *testing.T) {
			addrRgx := regexp.MustCompile(`([\d]+?\.){3}[\d]+`)
			m := dnsQuestion(dnsC.UDPURI, "A", "google.com")

			if m == nil {
				t.Errorf("unexpected nil DNS message")
				return
			}
			if len(m.Answer) != 1 {
				t.Errorf("unexpected answer length: wanted %v ; got %v", 1, len(m.Answer))
				return
			}

			if !strings.Contains(m.Answer[0].String(), "google.com") {
				t.Errorf("missing expected domain name: %s", m.Answer[0].String())
			}
			if !strings.Contains(m.Answer[0].String(), "A") {
				t.Errorf("missing expected record type: %s", m.Answer[0].String())
			}
			if !addrRgx.MatchString(m.Answer[0].String()) {
				t.Errorf("missing expected IP address: %s", m.Answer[0].String())
			}

			t.Log("[ok] external DNS query")
		})
	})
	t.Run("Health", func(t *testing.T) {
		b, status, err := httpReq(dnsC.HTTPURI, "/health", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if status != 200 {
			t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
			return
		}

		res := &httpapi.HealthResponse{}
		_ = json.Unmarshal(b, res)

		if res.Report.Status != health.Healthy {
			t.Errorf("service status is not as expected: wanted %v ; got %v", health.Healthy, res.Report.Status)
			return
		}

		t.Log("[ok] /health")
	})
	t.Run("Shutdown", func(t *testing.T) {
		t.Run("Store", func(t *testing.T) {
			t.Run("DeleteRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"record deleted successfully"}`

				b, status, err := httpReq(dnsC.HTTPURI, "/records/delete", []byte(`{"name":"really.not.a.dom.ain","type":"A"}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
					return
				}

				if string(b) != wants {
					t.Errorf("unexpected response: wanted %s ; got %s", wants, string(b))
					return
				}

				t.Log("[ok] /records/delete")
			})
		})
		t.Run("DNS", func(t *testing.T) {
			t.Run("StopDNS", func(t *testing.T) {
				b, status, err := httpReq(dnsC.HTTPURI, "/dns/stop", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}

				t.Log("[ok] /dns/stop")
			})
		})
	})
}
