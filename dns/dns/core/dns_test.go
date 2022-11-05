package core

import (
	"regexp"
	"strings"
	"testing"

	dns "github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

const (
	testName       = "not.a.dom.ain"
	testRealDomain = "google.com"
	testType       = "A"
	testAddr       = "192.168.0.10"
)

func TestAnswer(t *testing.T) {
	core := New()

	t.Run("Success", func(t *testing.T) {
		r := store.New().Name(testName).Type(testType).Addr(testAddr).Build()
		m := new(dns.Msg)

		core.Answer(r, m)

		if len(m.Answer) != 1 {
			t.Errorf("unexpected answer length: wanted %v ; got %v", 1, len(m.Answer))
			return
		}
		if !strings.Contains(m.Answer[0].String(), testName) {
			t.Errorf("unexpected answer: should contain domain %s ; got %s", testName, m.Answer[0].String())
		}
		if !strings.Contains(m.Answer[0].String(), testType) {
			t.Errorf("unexpected answer: should contain record type %s ; got %s", testType, m.Answer[0].String())
		}
		if !strings.Contains(m.Answer[0].String(), testAddr) {
			t.Errorf("unexpected answer: should contain address %s ; got %s", testAddr, m.Answer[0].String())
		}
	})
}

func TestFallback(t *testing.T) {
	core := New()
	addrRgx := regexp.MustCompile(`([\d]+?\.){3}[\d]+`)

	t.Run("Success", func(t *testing.T) {
		r := store.New().Name(testRealDomain).Type(testType).Build()
		m := new(dns.Msg)

		core.Fallback(r, m)

		if len(m.Answer) != 1 {
			t.Errorf("unexpected answer length: wanted %v ; got %v", 1, len(m.Answer))
		}
		if !strings.Contains(m.Answer[0].String(), testRealDomain) {
			t.Errorf("unexpected answer: should contain domain %s ; got %s", testRealDomain, m.Answer[0].String())
		}
		if !strings.Contains(m.Answer[0].String(), testType) {
			t.Errorf("unexpected answer: should contain record type %s ; got %s", testType, m.Answer[0].String())
		}
		if !addrRgx.MatchString(m.Answer[0].String()) {
			t.Errorf("unexpected answer: should contain an IP address; got %s", m.Answer[0].String())
		}
	})

	t.Run("FailNoResults", func(t *testing.T) {
		r := store.New().Name(testName).Type(testType).Build()
		m := new(dns.Msg)

		core.Fallback(r, m)

		if len(m.Answer) != 0 {
			t.Errorf("unexpected answer length: wanted %v ; got %v", 0, len(m.Answer))
		}
	})
}
