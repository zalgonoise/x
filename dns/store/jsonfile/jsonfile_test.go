package jsonfile

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/zalgonoise/x/dns/store"
)

var target = os.TempDir() + "/dns-test.list"

func rm(t *testing.T) {
	err := os.Remove(target)
	if err != nil {
		t.Errorf("failed to remove tempfile %s: %v", target, err)
	}
}

func TestNew(t *testing.T) {
	t.Run("SuccessNonExistingList", func(t *testing.T) {
		repo := New(target)
		if repo == nil {
			t.Errorf("repository was unexpectedly nil")
		}
		rm(t)
	})

	t.Run("SuccessNoItemsInList", func(t *testing.T) {
		f, err := os.Create(target)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		err = f.Sync()
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		repo := New(target)
		if repo == nil {
			t.Errorf("repository was unexpectedly nil")
		}
		rm(t)
	})

	t.Run("SuccessWithItemsInListFromJSON", func(t *testing.T) {
		ctx := context.Background()
		wants := store.New().Addr("192.168.0.10").Type("A").Name("not.a.dom.ain").Build()

		err := os.WriteFile(
			target,
			[]byte(`{"types":[{"type":"A","records":[{"address":"192.168.0.10","domains":["not.a.dom.ain"]}]}]}`),
			os.FileMode(store.OS_ALL_RW),
		)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		repo := New(target)
		if repo == nil {
			t.Errorf("repository was unexpectedly nil")
		}

		rs, err := repo.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 1 {
			t.Errorf("unexpected store list length: wanted 0, got %v", len(rs))
		}
		if !reflect.DeepEqual(wants, rs[0]) {
			t.Errorf("output mismatch error; wanted %v ; got %v", wants, rs[0])
		}
		rm(t)
	})

	t.Run("SuccessWithItemsInListFromYAML", func(t *testing.T) {
		ctx := context.Background()
		wants := store.New().Addr("192.168.0.10").Type("A").Name("not.a.dom.ain").Build()

		err := os.WriteFile(
			target,
			[]byte(`
types:
- type: A
  records:
  - address: 192.168.0.10
    domains:
    - not.a.dom.ain`),
			os.FileMode(store.OS_ALL_RW),
		)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		repo := New(target)
		if repo == nil {
			t.Errorf("repository was unexpectedly nil")
		}

		rs, err := repo.List(ctx)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(rs) != 1 {
			t.Errorf("unexpected store list length: wanted 0, got %v", len(rs))
		}
		if !reflect.DeepEqual(wants, rs[0]) {
			t.Errorf("output mismatch error; wanted %v ; got %v", wants, rs[0])
		}
		rm(t)
	})
}
