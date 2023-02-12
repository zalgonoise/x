package protofile

type GoField struct {
	IsRepeated bool   `json:"is_repeated,omitempty"`
	IsOptional bool   `json:"is_optional,omitempty"`
	IsStruct   bool   `json:"is_struct,omitempty"`
	GoType     string `json:"go_type,omitempty"`
	GoName     string `json:"go_name,omitempty"`
	ProtoType  string `json:"proto_type,omitempty"`
	ProtoIndex int    `json:"proto_index"`
	ProtoName  string `json:"proto_name,omitempty"`
}

type EnumField struct {
	Index     int    `json:"index"`
	GoName    string `json:"go_name,omitempty"`
	ProtoName string `json:"proto_name,omitempty"`
}

type GoType struct {
	Name        string    `json:"name,omitempty"`
	Fields      []GoField `json:"fields,omitempty"`
	uniqueNames map[string]struct{}
	uniqueIDs   map[int]struct{}
}

type GoEnum struct {
	ProtoName   string      `json:"proto_name,omitempty"`
	GoName      string      `json:"go_name,omitempty"`
	Fields      []EnumField `json:"fields,omitempty"`
	uniqueNames map[string]struct{}
	uniqueIDs   map[int]struct{}
}

type GoFile struct {
	Path        string              `json:"path,omitempty"`
	Package     string              `json:"package,omitempty"`
	Types       []*GoType           `json:"types,omitempty"`
	Enums       []*GoEnum           `json:"enums,omitempty"`
	UniqueTypes map[string]struct{} `json:"unique_types"`
}

var goTypes = map[string]string{
	"bool":     "bool",
	"uint32":   "uint32",
	"uint64":   "uint64",
	"sint32":   "int32",
	"sint64":   "int64",
	"int32":    "int32",
	"int64":    "int64",
	"fixed32":  "uint32",
	"fixed64":  "uint64",
	"sfixed32": "int32",
	"sfixed64": "int64",
	"double":   "float32",
	"float":    "float64",
	"string":   "string",
	"bytes":    "[]byte",
}
