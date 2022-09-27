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
	Element() *ElementsExtractor
	Generics() *GenericsExtractor
}

func (f *GoFile) Skip() *SkipExtractor {
	return &SkipExtractor{}
}

func (f *GoFile) Package() *PackageExtractor {
	return &PackageExtractor{
		f: f,
	}
}

func (f *GoFile) Import() *ImportExtractor {
	return &ImportExtractor{
		f: f,
	}
}

func (f *GoFile) Type() *TypeExtractor {
	return &TypeExtractor{
		f: f,
	}
}

func (f *GoFile) Element() *ElementsExtractor {
	return &ElementsExtractor{
		f: f,
	}
}

func (f *GoFile) Generics() *GenericsExtractor {
	return &GenericsExtractor{
		f: f,
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

func (f *GoFile) Parse() error {
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
		_, tok, lit := sc.Scan()
		if tok == token.EOF {
			break // end of GoFile
		}

		if extr.Done() {
			extr = f.match(tok)
		}
		extr.Do(tok, lit)

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
		return f.Type()
	default:
		return f.Skip()
	}
}
