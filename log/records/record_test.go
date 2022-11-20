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
	t.Run("ZeroTime", func(t *testing.T) {
		wants := record{
			message: testMsg,
			level:   testLevel,
			attrs:   testAttrs,
		}

		a := New(time.Time{}, testLevel, testMsg, testAttrs...)

		if a.Time().IsZero() {
			t.Errorf("expected time not to be zero ; got %v", a.Time())
		}
		wants.timestamp = a.Time()

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("ZeroUnixTime", func(t *testing.T) {
		wants := record{
			message: testMsg,
			level:   testLevel,
			attrs:   testAttrs,
		}

		a := New(time.Unix(0, 0), testLevel, testMsg, testAttrs...)

		if a.Time().IsZero() {
			t.Errorf("expected time not to be zero ; got %v", a.Time())
		}
		wants.timestamp = a.Time()

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("NoLevel", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   testMsg,
			level:     testLevel,
			attrs:     testAttrs,
		}

		a := New(testTime, nil, testMsg, testAttrs...)

		if !reflect.DeepEqual(wants, a) {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("NoMessage", func(t *testing.T) {
		wants := record{
			timestamp: testTime,
			message:   "",
			level:     testLevel,
			attrs:     testAttrs,
		}
		a := New(testTime, testLevel, "", testAttrs...)

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

func TestRecordAttrLen(t *testing.T) {
	t.Run("ZeroLen", func(t *testing.T) {
		wants := 0
		r := New(testTime, testLevel, testMsg)
		a := r.AttrLen()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("OneLen", func(t *testing.T) {
		wants := 1
		r := New(testTime, testLevel, testMsg, ta1)
		a := r.AttrLen()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("FourLen", func(t *testing.T) {
		wants := 4
		r := New(testTime, testLevel, testMsg, ta1, ta2, ta3, ta4)
		a := r.AttrLen()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}

func TestMessage(t *testing.T) {
	t.Run("Short", func(t *testing.T) {
		wants := "m"
		r := New(testTime, testLevel, "m", ta1, ta2, ta3, ta4)
		a := r.Message()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Long", func(t *testing.T) {
		wants := testMsg
		r := New(testTime, testLevel, testMsg, ta1, ta2, ta3, ta4)
		a := r.Message()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}

func TestLevel(t *testing.T) {
	t.Run("Trace", func(t *testing.T) {
		wants := level.Trace.Int()
		r := New(testTime, level.Trace, "m", ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Debug", func(t *testing.T) {
		wants := level.Debug.Int()
		r := New(testTime, level.Debug, "m", ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Info", func(t *testing.T) {
		wants := level.Info.Int()
		r := New(testTime, level.Info, "m", ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Warn", func(t *testing.T) {
		wants := level.Warn.Int()
		r := New(testTime, level.Warn, "m", ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Error", func(t *testing.T) {
		wants := level.Error.Int()
		r := New(testTime, level.Error, "m", ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
	t.Run("Fatal", func(t *testing.T) {
		wants := level.Fatal.Int()
		r := New(testTime, level.Fatal, testMsg, ta1, ta2, ta3, ta4)
		a := r.Level().Int()

		if a != wants {
			t.Errorf("unexpected output error: wanted %v ; got %v", wants, a)
		}
	})
}
