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
			whiteSpace: textH.conf.whiteSpace,
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
			whiteSpace: textH.conf.whiteSpace,
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
			whiteSpace: textH.conf.whiteSpace,
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
			whiteSpace: textH.conf.whiteSpace,
			timeFmt:    timeFmt,
		},
	}
}

// WithWhiteSpace creates a copy the Handler `h`, with the whitespace rune
// `whiteSpace`. Returns nil if the Handler is not a textHandler
func WithWhiteSpace(h handlers.Handler, whiteSpace rune) handlers.Handler {
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
			whiteSpace: whiteSpace,
			timeFmt:    textH.conf.timeFmt,
		},
	}
}
