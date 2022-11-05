// Package e2e will expect a running instance of the app, where this
// test file will hit the supported endpoints to ensure they are working
// as intended
//
// to run the tests, run the app via Docker or
// `sudo go run examples/core_memmap/main.go`
package e2e

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

var (
	record1 = store.New().Type("A").Name("not.a.dom.ain").Addr("192.168.0.10").Build()
	record2 = store.New().Type("A").Name("really.not.a.dom.ain").Addr("192.168.0.10").Build()
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
				b, status, err := httpReq("/records/add", []byte(`{"name":"not.a.dom.ain","type":"A","address":"192.168.0.10"}`))
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}

				if status != 200 {
					t.Errorf("unexpected status code: wanted %v ; got %v -- body: %s", 200, status, string(b))
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
		})
	})
	t.Run("UDP", func(t *testing.T) {

	})

	t.Run("HTTP", func(t *testing.T) {
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
}
