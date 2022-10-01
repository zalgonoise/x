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
	Name         string        `json:"name,omitempty"`
	Type         string        `json:"type,omitempty"`
	Kind         BlockType     `json:"kind,omitempty"`
	Generics     []*LogicBlock `json:"typeConstraints,omitempty"`
	InputParams  []*LogicBlock `json:"inputs,omitempty"`
	ReturnParams []*LogicBlock `json:"returns,omitempty"`
	BlockParams  []*LogicBlock `json:"elements,omitempty"`
	Calls        []string      `json:"calls,omitempty"`
	Package      string        `json:"pacakage,omitempty"`
}

type BlockType uint

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
	TypeGenericParam
)

var (
	blockTypeVals = map[BlockType]string{
		0:  "",
		1:  token.FUNC.String(),
		2:  "method",
		3:  token.STRUCT.String(),
		4:  token.INTERFACE.String(),
		5:  "funcParam",
		6:  "funcReturn",
		7:  token.DEFER.String(),
		8:  token.GO.String(),
		9:  token.VAR.String(),
		10: token.CONST.String(),
		11: "genericParam",
	}
	blockTypeKeys = map[string]BlockType{
		"":                       0,
		token.FUNC.String():      1,
		"method":                 2,
		token.STRUCT.String():    3,
		token.INTERFACE.String(): 4,
		"funcParam":              5,
		"funcReturn":             6,
		token.DEFER.String():     7,
		token.GO.String():        8,
		token.VAR.String():       9,
		token.CONST.String():     10,
		"genericParam":           11,
	}
)

func (b BlockType) String() string {
	return blockTypeVals[b]
}

func (f *GoFile) String() string {
	b, _ := json.Marshal(f)
	return string(b)
}

type Extractor interface {
	Do(pos token.Pos, tok token.Token, lit string) Extractor
	Done() bool
}
