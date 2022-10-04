package scan

import (
	"fmt"
	"go/scanner"
	"go/token"
	"os"
)

type GoFileScanner interface {
	Skip() *SkipExtractor
	Package() *PackageExtractor
	Import() *ImportExtractor
	Type() *TypeExtractor
	// Element() *ElementsExtractor
	Generics() *GenericsExtractor
}

// Skip returns a SkipExtractor
func (f *GoFile) Skip() *SkipExtractor {
	return &SkipExtractor{}
}

// Package returns a PackageExtractor
func (f *GoFile) Package() *PackageExtractor {
	return &PackageExtractor{
		f: f,
	}
}

// Import returns an ImportExtractor
func (f *GoFile) Import() *ImportExtractor {
	return &ImportExtractor{
		f: f,
	}
}

// Type returns a TypeExtractor for LogicBlock on index idx
func (f *GoFile) Type(idx int) *TypeExtractor {
	return &TypeExtractor{
		f:   f,
		idx: idx,
	}
}

// // Element returns an ElementsExtractor for LogicBlock on index idx, and for element index depth
// func (f *GoFile) Element(idx, depth int) *ElementsExtractor {
// 	return &ElementsExtractor{
// 		f:   f,
// 		idx: idx,
// 		lvl: depth,
// 	}
// }

// Struct returns a StructExtractor for LogicBlock on index idx, and for element index depth
func (f *GoFile) Struct(parent Extractor, idx, depth int) *StructExtractor {
	return &StructExtractor{
		f:      f,
		idx:    idx,
		lvl:    depth,
		parent: parent,
	}
}

// Interface returns a InterfaceExtractor for LogicBlock on index idx, and for element index depth
func (f *GoFile) Interface(parent Extractor, idx, depth int) *InterfaceExtractor {
	return &InterfaceExtractor{
		f:      f,
		idx:    idx,
		lvl:    depth,
		parent: parent,
	}
}

// Generics returns a GenericsExtractor for LogicBlock on index idx
func (f *GoFile) Generics(idx int) *GenericsExtractor {
	return &GenericsExtractor{
		f:   f,
		idx: idx,
	}
}

type Target int

const (
	TargetReceiver Target = iota + 1
	TargetInput
	TargetReturn
)

func (f *GoFile) BlockParam(parent Extractor, idx, depth int, target Target) *ParamExtractor {
	return &ParamExtractor{
		f:      f,
		idx:    idx,
		iidx:   depth,
		target: target,
		parent: parent,
	}
}

type Filter struct {
	key string
	idx int
}

func NewFilter(key string, idx int) Filter {
	switch key {
	case "logicBlock":
		return Filter{key: key, idx: idx}
	case "input":
		return Filter{key: key, idx: idx}
	case "return":
		return Filter{key: key, idx: idx}
	case "block":
		return Filter{key: key, idx: idx}
	case "receiver":
		return Filter{key: key, idx: idx}
	default:
		return Filter{}
	}
}

func applyFilters(f *GoFile, filters ...Filter) *LogicBlock {
	var lb *LogicBlock
	for _, filter := range filters {
		filter := filter
		switch filter.key {
		case "logicBlock":
			lb = f.GetLogicBlock(filter.idx)

		case "input":
			lb = lb.InputParam(filter.idx)
		case "return":
			lb = lb.ReturnParam(filter.idx)
		case "block":
			lb = lb.BlockParam(filter.idx)
		case "receiver":
			lb = lb.Receiver()
		}
	}
	return lb
}

func (f *GoFile) LBlock(e Extractor, filters ...Filter) *LogicBlockExtractor {
	lb := applyFilters(f, filters...)
	return &LogicBlockExtractor{
		parent: e,
		lb:     lb,
	}
}

func New(path string) (*GoFile, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return new(path, b), nil
}

func new(path string, b []byte) *GoFile {
	return &GoFile{
		Path:        path,
		bytes:       b,
		Imports:     []*Import{},
		LogicBlocks: []*LogicBlock{},
	}
}

type ParseFunc func(pos token.Pos, tok token.Token, lit string)

func (f *GoFile) Parse(fns ...ParseFunc) error {
	var (
		fs   = token.NewFileSet()
		sc   = scanner.Scanner{}
		err  error
		extr Extractor = f.Skip()
	)

	file := fs.AddFile(f.Path, fs.Base(), len(f.bytes))

	sc.Init(file, f.bytes, func(pos token.Position, msg string) {
		if err == nil {
			err = fmt.Errorf("error in %s: %s", pos.String(), msg)
			return
		}
		err = fmt.Errorf("error in %s: %s ; %w", pos.String(), msg, err)
	}, scanner.Mode(1))

	if err != nil {
		return err
	}

	for {
		pos, tok, lit := sc.Scan()
		if tok == token.EOF {
			break // end of GoFile
		}

		if extr.Done() {
			extr = f.match(tok)
		}
		extr = extr.Do(pos, tok, lit)

		// execute optional functions
		for _, fn := range fns {
			fn(pos, tok, lit)
		}
	}
	return nil
}

func (f *GoFile) match(tok token.Token) Extractor {
	switch tok {
	case token.PACKAGE:
		return f.Package()
	case token.IMPORT:
		return f.Import()
	case token.TYPE:
		f.typeCount += 1
		return f.Type(f.typeCount - 1)
	case token.SEMICOLON:
		return f.Skip()
	default:
		return f.Skip()
	}
}
