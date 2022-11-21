package texth

import (
	"github.com/zalgonoise/x/log/handlers"
)

// WithWrapper creates a copy the Handler `h`, with the left and right
// wrapper runes `leftWrapper` and `rightWrapper`. Returns nil if the Handler
// is not a textHandler
func WithWrapper(h handlers.Handler, leftWrapper, rightWrapper rune) handlers.Handler {
	textH, ok := (h).(textHandler)
	if !ok {
		return nil
	}

	return textHandler{
		w:         textH.w,
		addSource: textH.addSource,
		levelRef:  textH.levelRef,
		replFn:    textH.replFn,
		attrs:     textH.attrs,
		conf: textHandlerConfig{
			wrapperL:   leftWrapper,
			wrapperR:   rightWrapper,
			sepKV:      textH.conf.sepKV,
			sepAttr:    textH.conf.sepAttr,
			whitespace: textH.conf.whitespace,
			timeFmt:    textH.conf.timeFmt,
		},
	}
}

// WithWrapper creates a copy the Handler `h`, with the key-value separator
// string `kvSeparator`. Returns nil if the Handler is not a textHandler
func WithKVSeparator(h handlers.Handler, kvSeparator string) handlers.Handler {
	textH, ok := (h).(textHandler)
	if !ok {
		return nil
	}

	if kvSeparator == "" {
		kvSeparator = sepKV
	}

	return textHandler{
		w:         textH.w,
		addSource: textH.addSource,
		levelRef:  textH.levelRef,
		replFn:    textH.replFn,
		attrs:     textH.attrs,
		conf: textHandlerConfig{
			wrapperL:   textH.conf.wrapperL,
			wrapperR:   textH.conf.wrapperR,
			sepKV:      kvSeparator,
			sepAttr:    textH.conf.sepAttr,
			whitespace: textH.conf.whitespace,
			timeFmt:    textH.conf.timeFmt,
		},
	}
}

// WithWrapper creates a copy the Handler `h`, with the attribute key-value separator
// string `attrSeparator`. Returns nil if the Handler is not a textHandler
func WithAttrSeparator(h handlers.Handler, attrSeparator rune) handlers.Handler {
	textH, ok := (h).(textHandler)
	if !ok {
		return nil
	}

	return textHandler{
		w:         textH.w,
		addSource: textH.addSource,
		levelRef:  textH.levelRef,
		replFn:    textH.replFn,
		attrs:     textH.attrs,
		conf: textHandlerConfig{
			wrapperL:   textH.conf.wrapperL,
			wrapperR:   textH.conf.wrapperR,
			sepKV:      textH.conf.sepKV,
			sepAttr:    attrSeparator,
			whitespace: textH.conf.whitespace,
			timeFmt:    textH.conf.timeFmt,
		},
	}
}

// WithWrapper creates a copy the Handler `h`, with the time format string
// `timeFmt`. Returns nil if the Handler is not a textHandler
func WithTimeFormat(h handlers.Handler, timeFmt string) handlers.Handler {
	textH, ok := (h).(textHandler)
	if !ok {
		return nil
	}
	if timeFmt == "" {
		timeFmt = tFmt
	}

	return textHandler{
		w:         textH.w,
		addSource: textH.addSource,
		levelRef:  textH.levelRef,
		replFn:    textH.replFn,
		attrs:     textH.attrs,
		conf: textHandlerConfig{
			wrapperL:   textH.conf.wrapperL,
			wrapperR:   textH.conf.wrapperR,
			sepKV:      textH.conf.sepKV,
			sepAttr:    textH.conf.sepAttr,
			whitespace: textH.conf.whitespace,
			timeFmt:    timeFmt,
		},
	}
}

// WithWhitespace creates a copy the Handler `h`, with the whitespace rune
// `whitespace`. Returns nil if the Handler is not a textHandler
func WithWhitespace(h handlers.Handler, whitespace rune) handlers.Handler {
	textH, ok := (h).(textHandler)
	if !ok {
		return nil
	}

	return textHandler{
		w:         textH.w,
		addSource: textH.addSource,
		levelRef:  textH.levelRef,
		replFn:    textH.replFn,
		attrs:     textH.attrs,
		conf: textHandlerConfig{
			wrapperL:   textH.conf.wrapperL,
			wrapperR:   textH.conf.wrapperR,
			sepKV:      textH.conf.sepKV,
			sepAttr:    textH.conf.sepAttr,
			whitespace: whitespace,
			timeFmt:    textH.conf.timeFmt,
		},
	}
}
