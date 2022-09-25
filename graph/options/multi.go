package options

type multiOption struct {
	opts []Setting
}

// Apply sets the slice of options in the input pointer to a GraphConfig
func (o *multiOption) Apply(c *GraphConfig) {
	for _, opt := range o.opts {
		opt.Apply(c)
	}
}

// MultiOption function will return a single Setting wrapping the input
// Setting provided
func MultiOption(opts ...Setting) Setting {
	if len(opts) == 0 {
		return nil
	}
	if len(opts) == 1 {
		if opts[0] == nil {
			return nil
		}
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

	if len(multi) == 0 {
		return nil
	}

	return &multiOption{
		opts: multi,
	}
}
