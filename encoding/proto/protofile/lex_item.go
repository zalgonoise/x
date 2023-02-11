package protofile

type ProtoToken int

const (
	TokenEOF ProtoToken = iota
	TokenIDENT
	TokenTYPE
	TokenVALUE
	TokenEQUAL
	TokenDQUOTE
	TokenSEMICOL
	TokenLBRACE
	TokenRBRACE
	TokenSYNTAX
	TokenPACKAGE
	TokenOPTION
	TokenMESSAGE
	TokenENUM
	TokenREPEATED
	TokenOPTIONAL
)

var keywords = map[string]ProtoToken{
	"syntax":   TokenSYNTAX,
	"package":  TokenPACKAGE,
	"option":   TokenOPTION,
	"message":  TokenMESSAGE,
	"enum":     TokenENUM,
	"repeated": TokenREPEATED,
	"optional": TokenOPTIONAL,
}

var types = map[string]struct{}{
	"bool":     {},
	"uint32":   {},
	"uint64":   {},
	"sint32":   {},
	"sint64":   {},
	"int32":    {},
	"int64":    {},
	"fixed32":  {},
	"fixed64":  {},
	"sfixed32": {},
	"sfixed64": {},
	"double":   {},
	"float":    {},
	"string":   {},
	"bytes":    {},
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
