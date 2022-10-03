package scan

func (f *GoFile) SetMain(b bool) {
	f.IsMain = b
}

func (f *GoFile) SetPkgName(s string) {
	f.PackageName = s
}

func (f *GoFile) AddImport(i *Import) {
	f.Imports = append(f.Imports, i)
}

func (f *GoFile) AddLogicBlock(lb *LogicBlock) {
	f.LogicBlocks = append(f.LogicBlocks, lb)
}

func (f *GoFile) initLB(idx int) {
	if len(f.LogicBlocks) == idx {
		f.LogicBlocks = append(f.LogicBlocks, &LogicBlock{
			Generics:     []*LogicBlock{},
			InputParams:  []*LogicBlock{},
			ReturnParams: []*LogicBlock{},
			BlockParams:  []*LogicBlock{},
		})
	}
}

func (f *GoFile) SetLBName(idx int, s string) {
	f.initLB(idx)
	if f.LogicBlocks[idx].Name == "" {
		f.LogicBlocks[idx].Name = s
	}
}
func (f *GoFile) SetLBKind(idx int, t BlockType) {
	f.initLB(idx)
	if f.LogicBlocks[idx].Kind == 0 {
		f.LogicBlocks[idx].Kind = t
	}
}

func (f *GoFile) initITF(idx, iidx int) {
	if len(f.LogicBlocks[idx].BlockParams) == iidx {
		f.LogicBlocks[idx].BlockParams = append(f.LogicBlocks[idx].BlockParams, &LogicBlock{})
	}
}

func (f *GoFile) SetITFType(idx, iidx int) {
	f.initITF(idx, iidx)
	if f.LogicBlocks[idx].BlockParams[iidx].Type == "" {
		f.LogicBlocks[idx].BlockParams[iidx].Type = "method"
	}
}

func (f *GoFile) SetITFName(idx, iidx int, s string) {
	f.initITF(idx, iidx)
	if f.LogicBlocks[idx].BlockParams[iidx].Name == "" {
		f.LogicBlocks[idx].BlockParams[iidx].Name = s
	}
}

// TODO: builder pattern for LogicBlock? because of input / return elements, and so on

func (f *GoFile) SetPath(s string) {
	f.Path = s
}

func NewImport(pkg, uri string) *Import {
	return &Import{
		Package: pkg,
		URI:     uri,
	}
}
