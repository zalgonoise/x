package scan

// SetMain sets whether the file is a package main file
func (f *GoFile) SetMain(b bool) {
	f.IsMain = b
}

// SetPkgName sets the package name
func (f *GoFile) SetPkgName(s string) {
	f.PackageName = s
}

// AddImport appends the input Import pointer to the imports slice
func (f *GoFile) AddImport(i *Import) {
	f.Imports = append(f.Imports, i)
}

// AddLogicBlock appends the LogicBlock pointer to the LogicBlocks slice
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

// SetLBName sets the LogicBlock with index idx's name to s
func (f *GoFile) SetLBName(idx int, s string) {
	f.initLB(idx)
	if f.LogicBlocks[idx].Name == "" {
		f.LogicBlocks[idx].Name = s
	}
}

// SetLBKind sets the LogicBlock with index idx's kind to t
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

// SetITFType sets the LogicBlock with index idx and entry index iidx's type to method
func (f *GoFile) SetITFType(idx, iidx int) {
	f.initITF(idx, iidx)
	if f.LogicBlocks[idx].BlockParams[iidx].Type == "" {
		f.LogicBlocks[idx].BlockParams[iidx].Type = "method"
	}
}

// SetITFName sets the LogicBlock with index idx and entry index iidx 's name to s
func (f *GoFile) SetITFName(idx, iidx int, s string) {
	f.initITF(idx, iidx)
	if f.LogicBlocks[idx].BlockParams[iidx].Name == "" {
		f.LogicBlocks[idx].BlockParams[iidx].Name = s
	}
}

// GetLogicBlock will retrieve the LogicBlock on index idx
func (f *GoFile) GetLogicBlock(idx int) *LogicBlock {
	if idx == len(f.LogicBlocks) {
		f.LogicBlocks = append(f.LogicBlocks, &LogicBlock{})
	}
	return f.LogicBlocks[idx]
}

// SetPath sets the FS path to s
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
	if s == "." {
		l.Type += "."
		return
	}
	if len(l.Type) > 0 && l.Type[len(l.Type)-1] == '.' {
		l.Type = l.Type + s
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
