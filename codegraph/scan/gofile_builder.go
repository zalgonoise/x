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

func (f *GoFile) GetLogicBlock(idx int) *LogicBlock {
	if idx == len(f.LogicBlocks) {
		f.LogicBlocks = append(f.LogicBlocks, &LogicBlock{})
	}
	return f.LogicBlocks[idx]
}

func (f *GoFile) SetPath(s string) {
	f.Path = s
}

func NewImport(pkg, uri string) *Import {
	return &Import{
		Package: pkg,
		URI:     uri,
	}
}

// TODO: builder pattern for LogicBlock? because of input / return elements, and so on
type LogicBlockBuilder interface {
	SetName(string)
	SetType(string)
	SetKind(BlockType)
	GenericParam(idx int) *LogicBlock
	InputParam(idx int) *LogicBlock
	ReturnParam(idx int) *LogicBlock
	BlockParam(idx int) *LogicBlock
	GenericLen() int
	InputLen() int
	ReturnLen() int
	BlockLen() int
	Receiver() *LogicBlock
	AddCall(string)
	SetPackage(string)
	IsFunc()
}

func (l *LogicBlock) SetName(s string) {
	l.Name = s
}

func (l *LogicBlock) SetType(s string) {
	if s == "*" {
		l.Type = "*" + l.Type
		return
	}
	if l.Type != "" {
		l.Name = l.Type
		l.Type = s
		return
	}
	l.Type = s
}

func (l *LogicBlock) SetKind(k BlockType) {
	l.Kind = k
}

func (l *LogicBlock) GenericParam(idx int) *LogicBlock {
	if idx == len(l.Generics) {
		l.Generics = append(l.Generics, &LogicBlock{})
	}
	return l.Generics[idx]
}

func (l *LogicBlock) InputParam(idx int) *LogicBlock {
	if idx == len(l.InputParams) {
		l.InputParams = append(l.InputParams, &LogicBlock{})
	}
	return l.InputParams[idx]
}

func (l *LogicBlock) ReturnParam(idx int) *LogicBlock {
	if idx == len(l.ReturnParams) {
		l.ReturnParams = append(l.ReturnParams, &LogicBlock{})
	}
	return l.ReturnParams[idx]
}

func (l *LogicBlock) BlockParam(idx int) *LogicBlock {
	if idx == len(l.BlockParams) {
		l.BlockParams = append(l.BlockParams, &LogicBlock{})
	}
	return l.BlockParams[idx]
}

func (l *LogicBlock) Receiver() *LogicBlock {
	if l.Receivers == nil {
		l.Receivers = &LogicBlock{}
	}
	return l.Receivers
}

func (l *LogicBlock) AddCall(s string) {
	l.Calls = append(l.Calls, s)
}

func (l *LogicBlock) SetPackage(s string) {
	l.Package = s
}

func (l *LogicBlock) GenericLen() int {
	return len(l.Generics)
}

func (l *LogicBlock) InputLen() int {
	return len(l.InputParams)
}

func (l *LogicBlock) ReturnLen() int {
	return len(l.ReturnParams)
}

func (l *LogicBlock) BlockLen() int {
	return len(l.BlockParams)
}

func (l *LogicBlock) IsFunc() {
	l.Kind = TypeFuncParam
	l.Name = "func"
	l.isFunc = true
}
