package texth

import (
	"bytes"
	"testing"
	"time"
)

func TestWithWrapper(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Success", func(t *testing.T) {
		wantL := '{'
		wantR := '}'
		wh := WithWrapper(h, wantL, wantR)

		if wh.(textHandler).conf.wrapperL != wantL {
			t.Errorf("unexpected left wrapper: %s", string(wh.(textHandler).conf.wrapperL))
		}
		if wh.(textHandler).conf.wrapperR != wantR {
			t.Errorf("unexpected right wrapper: %s", string(wh.(textHandler).conf.wrapperL))
		}
	})
	t.Run("Fail", func(t *testing.T) {
		wh := WithWrapper(nil, '{', '}')
		if wh != nil {
			t.Errorf("expected output to be nil")
		}
	})
}

func TestWithKVSeparator(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Success", func(t *testing.T) {
		wants := " = "
		wh := WithKVSeparator(h, wants)

		if wh.(textHandler).conf.sepKV != wants {
			t.Errorf("unexpected key-value separator: %s", wh.(textHandler).conf.sepKV)
		}
	})
	t.Run("EmptyString", func(t *testing.T) {
		wants := ": "
		wh := WithKVSeparator(h, "")

		if wh.(textHandler).conf.sepKV != wants {
			t.Errorf("unexpected key-value separator: %s", wh.(textHandler).conf.sepKV)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		wh := WithKVSeparator(nil, "")
		if wh != nil {
			t.Errorf("expected output to be nil")
		}
	})
}

func TestWithAttrSeparator(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Success", func(t *testing.T) {
		wants := ','
		wh := WithAttrSeparator(h, wants)

		if wh.(textHandler).conf.sepAttr != wants {
			t.Errorf("unexpected key-value separator: %s", string(wh.(textHandler).conf.sepAttr))
		}
	})
	t.Run("Fail", func(t *testing.T) {
		wh := WithAttrSeparator(nil, ';')
		if wh != nil {
			t.Errorf("expected output to be nil")
		}
	})
}

func TestWithTimeFormat(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Success", func(t *testing.T) {
		wants := time.RubyDate
		wh := WithTimeFormat(h, wants)
		if wh.(textHandler).conf.timeFmt != wants {
			t.Errorf("unexpected key-value separator: %s", wh.(textHandler).conf.timeFmt)
		}
	})
	t.Run("EmptyString", func(t *testing.T) {
		wants := tFmt
		wh := WithTimeFormat(h, "")
		if wh.(textHandler).conf.timeFmt != wants {
			t.Errorf("unexpected key-value separator: %s", wh.(textHandler).conf.timeFmt)
		}
	})
	t.Run("Fail", func(t *testing.T) {
		wh := WithTimeFormat(nil, tFmt)
		if wh != nil {
			t.Errorf("expected output to be nil")
		}
	})
}

func TestWithWhitespace(t *testing.T) {
	b := &bytes.Buffer{}
	h := New(b)

	t.Run("Success", func(t *testing.T) {
		wants := '\t'
		wh := WithWhitespace(h, wants)
		if wh.(textHandler).conf.whitespace != wants {
			t.Errorf("unexpected key-value separator: %s", string(wh.(textHandler).conf.whitespace))
		}
	})
	t.Run("Fail", func(t *testing.T) {
		wh := WithWhitespace(nil, ' ')
		if wh != nil {
			t.Errorf("expected output to be nil")
		}
	})
}
