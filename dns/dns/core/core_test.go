package core

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("ZeroFallbackDNS", func(t *testing.T) {
		wants := &DNSCore{
			fallbackDNS: defaultFallback,
		}

		dnsCore := New()

		if !reflect.DeepEqual(wants, dnsCore) {
			t.Errorf("output mismatch error -- wanted %v ; got %v", wants, dnsCore)
		}
	})

	t.Run("OneFallbackDNS", func(t *testing.T) {
		wants := &DNSCore{
			fallbackDNS: []string{"1.1.1.1:53"},
		}

		dnsCore := New("1.1.1.1")

		if !reflect.DeepEqual(wants, dnsCore) {
			t.Errorf("output mismatch error -- wanted %v ; got %v", wants, dnsCore)
		}
	})

	t.Run("ManyFallbackDNS", func(t *testing.T) {
		wants := &DNSCore{
			fallbackDNS: []string{"1.1.1.1:53", "8.8.8.8:53"},
		}

		dnsCore := New("1.1.1.1:53", "8.8.8.8", "")

		if !reflect.DeepEqual(wants, dnsCore) {
			t.Errorf("output mismatch error -- wanted %v ; got %v", wants, dnsCore)
		}
	})
}
