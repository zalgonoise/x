package texth

import (
	"github.com/zalgonoise/x/log/handlers"
)

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
