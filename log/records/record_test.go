package records

import (
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/zalgonoise/x/log/attr"
	"github.com/zalgonoise/x/log/level"
)

var (
	testTime  = time.Unix(1668802887, 0)
	testMsg   = "test message"
	testLevel = level.Info

	ta1 = attr.String("a_key", "value")
	ta2 = attr.Int("b_test_no", 1)
	ta3 = attr.Float("c_success_rate", 1.0)
	ta4 = attr.New("d_custom", struct {
		key   string
		value int
	}{
		key:   "custom_key",
		value: 2,
	})

	testAttrs = []attr.Attr{
		ta1, ta2, ta3, ta4,
	}
)

func TestNew(t *testing.T) {
	t.Run("Simple", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     []attr.Attr{},
		}

		a := New(testTime, testLevel, testMsg)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Complex", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     testAttrs,
		}

		a := New(testTime, testLevel, testMsg, testAttrs...)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}

func TestRecordAddAttr(t *testing.T) {
	t.Run("AddOne", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     []attr.Attr{ta1, ta2},
		}

		r := New(testTime, testLevel, testMsg, ta1)
		a := r.AddAttr(ta2)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("AddMany", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     testAttrs,
		}

		r := New(testTime, testLevel, testMsg, ta1)
		a := r.AddAttr(ta2, ta3, ta4)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("AddNone", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     []attr.Attr{ta1},
		}

		r := New(testTime, testLevel, testMsg, ta1)
		a := r.AddAttr()

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}

func TestRecordAttr(t *testing.T) {
	t.Run("IdxZero", func(t *testing.T) {
		wants := ta1

		r := New(testTime, testLevel, testMsg, ta1, ta2)
		a := r.Attr(0)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("IdxThree", func(t *testing.T) {
		wants := ta4

		r := New(testTime, testLevel, testMsg, ta1, ta2, ta3, ta4)
		a := r.Attr(3)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("IdxOOB", func(t *testing.T) {
		r := New(testTime, testLevel, testMsg, ta1, ta2)
		a := r.Attr(90)

		if a != nil {
			t.Errorf("expected value to be nil; got %v of type %T", a, a)
		}
	})
}

func TestRecordAttrs(t *testing.T) {
	t.Run("GetOne", func(t *testing.T) {
		wants := []attr.Attr{ta1}

		r := New(testTime, testLevel, testMsg, ta1)
		a := r.Attrs()
		sort.Slice(a, func(i, j int) bool {
			return a[i].Key()[0] < a[j].Key()[0]
		})

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("GetMany", func(t *testing.T) {
		wants := []attr.Attr{ta1, ta2, ta3, ta4}

		r := New(testTime, testLevel, testMsg, testAttrs...)
		a := r.Attrs()
		sort.Slice(a, func(i, j int) bool {
			return a[i].Key()[0] < a[j].Key()[0]
		})

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("GetNone", func(t *testing.T) {
		wants := []attr.Attr{}

		r := New(testTime, testLevel, testMsg)
		a := r.Attrs()
		sort.Slice(a, func(i, j int) bool {
			return a[i].Key()[0] < a[j].Key()[0]
		})

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}
