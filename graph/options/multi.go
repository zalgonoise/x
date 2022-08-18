package options

type multiOption struct {
	opts []Setting
}

func (o *multiOption) Apply(c *GraphConfig) {
	for _, opt := range o.opts {
		opt.Apply(c)
	}
}

func MultiOption(opts ...Setting) Setting {
	if len(opts) == 0 {
		return nil
	}
	if len(opts) == 1 {
		return opts[0]
	}

	multi := []Setting{}

	for _, opt := range opts {
		if opt == nil {
			continue
		}
		if mo, ok := opt.(*multiOption); ok {
			multi = append(multi, mo.opts...)
			continue
		}
		multi = append(multi, opt)
	}
	return &multiOption{
		opts: multi,
	}
}
