// Package e2e will expect a running instance of the app, where this
// test file will hit the supported endpoints to ensure they are working
// as intended
//
// to run the tests, run the app via Docker or
// `sudo go run examples/core_memmap/main.go`
package e2e

import (
	"bytes"
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
	"github.com/zalgonoise/x/dns/health"
	"github.com/zalgonoise/x/dns/store"
	"github.com/zalgonoise/x/dns/transport/httpapi/endpoints"
)

func httpReq(endpoint string, data []byte) ([]byte, int, error) {
	var (
		res *http.Response
		err error
	)

	switch data {
	case nil:
		res, err = http.Get(fmt.Sprintf("http://localhost:%v%s", 8080, endpoint))
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, 0, err
		}
	default:
		buf := &bytes.Buffer{}
		_, err := buf.Write(data)
		if err != nil {
			return nil, 0, err
		}
		res, err = http.Post(fmt.Sprintf("http://localhost:%v%s", 8080, endpoint), "application/json", buf)
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

func dnsQuestion(rtype, domain string) *dns.Msg {
	message := new(dns.Msg)
	out := new(dns.Msg)
	message.SetQuestion(dns.Fqdn(domain), store.RecordTypeInts[rtype])
	client := &dns.Client{
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
		Net:          "udp",
	}
	in, _, err := client.Exchange(message, "localhost:53")
	if err != nil || len(in.Answer) == 0 {
		return message
	}
	out.Answer = append(out.Answer, in.Answer...)

	return out
}

func TestTransport(t *testing.T) {
	t.Run("HTTP", func(t *testing.T) {
		t.Run("DNS", func(t *testing.T) {
			t.Run("StartDNS", func(t *testing.T) {
				b, status, err := httpReq("/dns/start", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}
			})
			t.Run("ReloadDNS", func(t *testing.T) {
				b, status, err := httpReq("/dns/reload", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}
			})
		})
		t.Run("Store", func(t *testing.T) {
			t.Run("AddRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"added record successfully","record":{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq("/records/add", []byte(`{"name":"not.a.dom.ain","type":"A","address":"192.168.0.10"}`))
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
			})
			t.Run("ListRecords", func(t *testing.T) {
				wants := `{"success":true,"message":"listing all records","records":[{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}]}`

				b, status, err := httpReq("/records", nil)
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
			})
			t.Run("GetRecordByDomainAndType", func(t *testing.T) {
				wants := `{"success":true,"message":"fetched record for domain not.a.dom.ain","record":{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq("/records/getAddress", []byte(`{"name":"not.a.dom.ain","type":"A"}`))
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
			})
			t.Run("GetRecordByAddress", func(t *testing.T) {
				wants := `{"success":true,"message":"listing all records for IP address 192.168.0.10","records":[{"type":"A","name":"not.a.dom.ain","address":"192.168.0.10"}]}`

				b, status, err := httpReq("/records/getDomains", []byte(`{"address":"192.168.0.10"}`))
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
			})
			t.Run("UpdateRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"updated record successfully","record":{"type":"A","name":"really.not.a.dom.ain","address":"192.168.0.10"}}`

				b, status, err := httpReq("/records/update", []byte(`{"target":"not.a.dom.ain","record":{"name":"really.not.a.dom.ain","type":"A","address":"192.168.0.10"}}`))
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
			})
		})
	})
	t.Run("UDP", func(t *testing.T) {
		t.Run("LocalQuery", func(t *testing.T) {
			m := dnsQuestion("A", "really.not.a.dom.ain")

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
		})
		t.Run("ExternalQuery", func(t *testing.T) {
			addrRgx := regexp.MustCompile(`([\d]+?\.){3}[\d]+`)
			m := dnsQuestion("A", "google.com")

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
		})
	})
	t.Run("Health", func(t *testing.T) {
		b, status, err := httpReq("/health", nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
			return
		}

		if status != 200 {
			t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
			return
		}

		res := &endpoints.HealthResponse{}
		_ = json.Unmarshal(b, res)

		if res.Report.Status != health.Healthy {
			t.Errorf("service status is not as expected: wanted %v ; got %v", health.Healthy, res.Report.Status)
			return
		}

	})
	t.Run("Shutdown", func(t *testing.T) {
		t.Run("Store", func(t *testing.T) {
			t.Run("DeleteRecord", func(t *testing.T) {
				wants := `{"success":true,"message":"record deleted successfully"}`

				b, status, err := httpReq("/records/delete", []byte(`{"name":"really.not.a.dom.ain","type":"A"}`))
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

			})
		})
		t.Run("DNS", func(t *testing.T) {
			t.Run("StopDNS", func(t *testing.T) {
				b, status, err := httpReq("/dns/stop", nil)
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
				}
			})
		})
	})
}
