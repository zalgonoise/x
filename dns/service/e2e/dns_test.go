package e2e

import (
	"context"
	"testing"

	"github.com/miekg/dns"
	"github.com/zalgonoise/x/dns/store"
)

func TestDNS(t *testing.T) {
	s := initializeService()

	err := s.AddRecords(context.Background(), record1, record2)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	t.Run("AnswerDNS", func(t *testing.T) {
		input := store.New().Type(record1.Type).Name(record1.Name).Build()
		m := new(dns.Msg)
		wants := "not.a.dom.ain.	3600	IN	A	192.168.0.10"

		s.AnswerDNS(input, m)

		if len(m.Answer) != 1 {
			t.Errorf("unexpected answers list length: wanted %v ; got %v", 1, len(m.Answer))
			return
		}

		if m.Answer[0].String() != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", record1, m.Answer[0])
			return
		}
	})
}
