package file

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/dns/store"
)

var (
	test1 = store.New().Name("not.a.dom.ain").Type("A").Addr("192.168.0.10").Build()
	test2 = store.New().Name("really.not.a.dom.ain").Type("A").Addr("192.168.0.10").Build()
	// test3 = store.New().Name("really.not.a.dom.ain").Type("CNAME").Addr("am.i.not.a.dom.ain.").Build()
	// test4 = store.New().Name("am.i.not.a.dom.ain").Type("A").Addr("192.168.0.15").Build()
)

func TestCreate(t *testing.T) {
	t.Run("SuccessJSON", func(t *testing.T) {
		ctx := context.Background()
		s := New("json", target)
		wants := `{"types":[{"type":"A","records":[{"address":"192.168.0.10","domains":["not.a.dom.ain"]}]}]}`

		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		b, err := os.ReadFile(target)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(wants, string(b)) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, string(b))
		}
	})
	t.Run("SuccessYAML", func(t *testing.T) {
		ctx := context.Background()
		s := New("yaml", target)
		wants := `types:
- type: A
  records:
  - address: 192.168.0.10
    domains:
    - not.a.dom.ain
`
		err := s.Create(ctx, test1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		b, err := os.ReadFile(target)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(wants, string(b)) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, string(b))
		}
	})
}

func TestList(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New("yaml", target)

		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rs) != 1 {
			t.Errorf("output length error: wanted %v ; got %v", 1, len(rs))
		}
		if !reflect.DeepEqual(test1, rs[0]) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test1, rs[0])
		}
	})
}

func TestFilterByTypeAndDomain(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New("yaml", target)

		r, err := s.FilterByTypeAndDomain(ctx, test1.Type, test1.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test1, r) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test1, r)
		}
	})
}

func TestFilterByDest(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		s := New("yaml", target)

		rs, err := s.FilterByDest(ctx, test1.Addr)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rs) != 1 {
			t.Errorf("output length error: wanted %v ; got %v", 1, len(rs))
		}
		if !reflect.DeepEqual(test1, rs[0]) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test1, rs[0])
		}
	})
}
