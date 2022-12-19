package codegraph

type GoFile struct {
	IsMain      bool          `json:"isMain,omitempty"`
	PackageName string        `json:"package,omitempty"`
	Imports     []*Import     `json:"imports,omitempty"`
	LogicBlocks []*LogicBlock `json:"logicBlocks,omitempty"`
	Path        string        `json:"path,omitempty"`
}

type Import struct {
	Package string `json:"package,omitempty"`
	URI     string `json:"uri,omitempty"`
}

type LogicBlock struct {
	Name         string         `json:"name,omitempty"`
	Type         string         `json:"type,omitempty"`
	Kind         LogicBlockKind `json:"kind,omitempty"`
	Generics     []*LogicBlock  `json:"typeConstraints,omitempty"`
	InputParams  []*Identifier  `json:"inputs,omitempty"`
	ReturnParams []*Identifier  `json:"returns,omitempty"`
	Receiver     *Identifier    `json:"receiver,omitempty"`
	parent       *GoFile
}

type Identifier struct {
	Package string
	Name    string
	Type    string
}

type LogicBlockKind uint

const (
	TypeUndefined LogicBlockKind = iota
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
