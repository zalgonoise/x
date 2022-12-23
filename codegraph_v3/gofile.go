package codegraph

type GoFile struct {
	IsMain      *bool     `json:"isMain,omitempty"`
	PackageName string    `json:"package,omitempty"`
	Imports     []*Import `json:"imports,omitempty"`
	LogicBlocks []*Type   `json:"logic_blocks,omitempty"`
	Path        string    `json:"path,omitempty"`
}

type Import struct {
	Name    *string `json:"name,omitempty"`
	Package string  `json:"package,omitempty"`
}

type Type struct {
	IsPointer *bool          `json:"is_pointer,omitempty"`
	Name      string         `json:"name,omitempty"`
	Type      string         `json:"type,omitempty"`
	Package   *string        `json:"package,omitempty"`
	Kind      LogicBlockKind `json:"kind,omitempty"`
	Slice     *RSlice        `json:"slice,omitempty"`
	Map       *RMap          `json:"map,omitempty"`
	Generics  []*Type        `json:"type_params,omitempty"`
	Func      *RFunc         `json:"func,omitempty"`
	Struct    *RStruct       `json:"struct,omitempty"`
	Interface *RInterface    `json:"interface,omitempty"`
}

type RSlice struct {
	IsSlice    *bool   `json:"is_slice,omitempty"`
	IsPointer  *bool   `json:"is_pointer,omitempty"`
	IsVariadic *bool   `json:"is_variadic,omitempty"`
	Len        *int    `json:"len,omitempty"`
	LenName    *string `json:"len_var_name,omitempty"`
}

type RMap struct {
	IsMap     *bool `json:"is_map,omitempty"`
	IsPointer *bool `json:"is_pointer,omitempty"`
	Key       Type  `json:"key,omitempty"`
	Value     Type  `json:"value,omitempty"`
}

type RFunc struct {
	IsFunc      *bool   `json:"is_func,omitempty"`
	Receiver    *Type   `json:"receiver,omitempty"`
	InputParams []*Type `json:"input_params,omitempty"`
	Returns     []*Type `json:"returns,omitempty"`
}

type RStruct struct {
	IsStruct *bool   `json:"is_struct,omitempty"`
	Elems    []*Type `json:"elements"`
}

type RInterface struct {
	IsInterface *bool   `json:"is_interface,omitempty"`
	Methods     []*Type `json:"methods"`
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
