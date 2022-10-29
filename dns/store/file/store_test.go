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
)

func TestJSON(t *testing.T) {
	s := New("json", target)
	defer rm(t)

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
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
	t.Run("List", func(t *testing.T) {
		ctx := context.Background()

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
	t.Run("FilterByTypeAndDomain", func(t *testing.T) {
		ctx := context.Background()

		r, err := s.FilterByTypeAndDomain(ctx, test1.Type, test1.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test1, r) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test1, r)
		}
	})
	t.Run("FilterByDest", func(t *testing.T) {
		ctx := context.Background()

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
	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		wants := `{"types":[{"type":"A","records":[{"address":"192.168.0.10","domains":["really.not.a.dom.ain"]}]}]}`

		err := s.Update(ctx, test1.Name, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		r, err := s.FilterByTypeAndDomain(ctx, test2.Type, test2.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test2, r) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test2, r)
		}

		b, err := os.ReadFile(target)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(wants, string(b)) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, string(b))
		}
	})
	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()
		wants := `{}`

		err := s.Delete(ctx, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rs) != 0 {
			t.Errorf("unexpected records length: wanted %v ; got %v", 0, len(rs))
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

func TestYAML(t *testing.T) {
	s := New("yaml", target)
	defer rm(t)

	t.Run("Create", func(t *testing.T) {
		ctx := context.Background()
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
	t.Run("List", func(t *testing.T) {
		ctx := context.Background()

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
	t.Run("FilterByTypeAndDomain", func(t *testing.T) {
		ctx := context.Background()

		r, err := s.FilterByTypeAndDomain(ctx, test1.Type, test1.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test1, r) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test1, r)
		}
	})
	t.Run("FilterByDest", func(t *testing.T) {
		ctx := context.Background()

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
	t.Run("Update", func(t *testing.T) {
		ctx := context.Background()
		wants := `types:
- type: A
  records:
  - address: 192.168.0.10
    domains:
    - really.not.a.dom.ain
`
		err := s.Update(ctx, test1.Name, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		r, err := s.FilterByTypeAndDomain(ctx, test2.Type, test2.Name)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(test2, r) {
			t.Errorf("output mismatch error: wanted %s ; got %s", test2, r)
		}

		b, err := os.ReadFile(target)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !reflect.DeepEqual(wants, string(b)) {
			t.Errorf("output mismatch error: wanted %s ; got %s", wants, string(b))
		}
	})
	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()
		wants := `{}
`

		err := s.Delete(ctx, test2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		rs, err := s.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(rs) != 0 {
			t.Errorf("unexpected records length: wanted %v ; got %v", 0, len(rs))
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
