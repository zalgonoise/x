package scan

import (
	"encoding/json"
	"go/token"
)

type Project struct {
	Imports  []string  `json:"imports,omitempty"`
	Packages []Package `json:"packages,omitempty"`
}

type Package struct {
	Name  string   `json:"name,omitempty"`
	Files []GoFile `json:"goFiles,omitempty"`
}

type GoFile struct {
	IsMain      bool          `json:"isMain,omitempty"`
	PackageName string        `json:"package,omitempty"`
	Imports     []*Import     `json:"imports,omitempty"`
	LogicBlocks []*LogicBlock `json:"logicBlocks,omitempty"`
	Path        string        `json:"path,omitempty"`
	bytes       []byte
	typeCount   int
}

type Import struct {
	Package string `json:"package,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type LogicBlock struct {
	Name         string    `json:"name,omitempty"`
	Type         BlockType `json:"type,omitempty"`
	Generics     []*Param  `json:"typeConstraints,omitempty"`
	InputParams  []*Param  `json:"inputs,omitempty"`
	ReturnParams []*Param  `json:"returns,omitempty"`
	BlockParams  []*Param  `json:"elements,omitempty"`
	Calls        []string  `json:"calls,omitempty"`
	Package      string    `json:"pacakage,omitempty"`
	// level        int
}

type Param struct {
	Name    string `json:"name,omitempty"`
	Type    string `json:"type,omitempty"`
	Package string `json:"package,omitempty"`
}

type BlockType uint
type ParamType uint

const (
	TypeUndefined BlockType = iota
	TypeFunction
	TypeMethod
	TypeStruct
	TypeInterface
	TypeFuncParam
	TypeFuncReturn
	TypeDeferFunc
	TypeGoFunc
	TypeVariableDecl
	TypeConstantDecl
)

func (f *GoFile) String() string {
	b, _ := json.Marshal(f)
	return string(b)
}

type Extractor interface {
	Do(pos token.Pos, tok token.Token, lit string) Extractor
	Done() bool
}
