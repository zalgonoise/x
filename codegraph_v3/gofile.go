package codegraph

type GoFile struct {
	IsMain      *bool         `json:"isMain,omitempty"`
	PackageName string        `json:"package,omitempty"`
	Imports     []*Import     `json:"imports,omitempty"`
	LogicBlocks []*LogicBlock `json:"logicBlocks,omitempty"`
	Path        string        `json:"path,omitempty"`
}

type Import struct {
	Name    *string `json:"name,omitempty"`
	Package string  `json:"package,omitempty"`
}

func NewImport() *Import {
	return &Import{}
}

type LogicBlock struct {
	IsPointer    *bool          `json:"is_pointer,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Type         *string        `json:"type,omitempty"`
	Kind         LogicBlockKind `json:"kind,omitempty"`
	Generics     []*Identifier  `json:"generic_types,omitempty"`
	InputParams  []*LogicBlock  `json:"inputs,omitempty"`
	ReturnParams []*LogicBlock  `json:"returns,omitempty"`
	Receiver     *Identifier    `json:"receiver,omitempty"`
	Package      string         `json:"package,omitempty"`
	parent       *GoFile
}

func NewLogicBlock(kind LogicBlockKind) *LogicBlock {
	return &LogicBlock{
		Kind: kind,
	}
}

type Identifier struct {
	IsPointer    *bool          `json:"is_pointer,omitempty"`
	Package      string         `json:"package,omitempty"`
	Name         *string        `json:"name,omitempty"`
	Type         string         `json:"type,omitempty"`
	GenericTypes []*Identifier  `json:"generic_types,omitempty"`
	Kind         LogicBlockKind `json:"kind,omitempty"`
}

func NewIdentifier() *Identifier {
	return &Identifier{}
}

type LogicBlockKind uint

const (
	TypeUndefined LogicBlockKind = iota
	TypeFunction
	TypeMethod
	TypeReceiver
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

var kindMap = map[LogicBlockKind][]byte{
	TypeUndefined:    []byte(`"undefined"`),
	TypeFunction:     []byte(`"function"`),
	TypeMethod:       []byte(`"method"`),
	TypeReceiver:     []byte(`"receiver"`),
	TypeStruct:       []byte(`"struct"`),
	TypeInterface:    []byte(`"interface"`),
	TypeFuncParam:    []byte(`"func_parameter"`),
	TypeFuncReturn:   []byte(`"func_return"`),
	TypeDeferFunc:    []byte(`"defer_func"`),
	TypeGoFunc:       []byte(`"go_func"`),
	TypeVariableDecl: []byte(`"var_declr"`),
	TypeConstantDecl: []byte(`"const_declr"`),
	TypeGenericParam: []byte(`"generic_type"`),
}

func (k LogicBlockKind) String() string {
	return string(kindMap[k])
}

func (k LogicBlockKind) MarshalJSON() ([]byte, error) {
	return kindMap[k], nil
}
