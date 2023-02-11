package protofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/zalgonoise/parse"
)

const supportedSyntax = "proto3"
const supportedOption = "go_package"

var (
	ErrInvalidSyntax     = fmt.Errorf("invalid syntax (only %s is supported)", supportedSyntax)
	ErrInvalidEdgeAmount = errors.New("invalid amount of edges")
	ErrMissingPackage    = errors.New("missing package name")
	ErrInvalidOption     = errors.New(`invalid option; include 'option go_package = "./my_package"'`)
	ErrInvalidTokenType  = errors.New("invalid token type")
	ErrAlreadyExistsName = errors.New("name already exists")
	ErrAlreadyExistsID   = errors.New("ID already exists")
)

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

func processFn[C ProtoToken, T byte, R string](t *parse.Tree[C, T]) (R, error) {
	var goFile = new(GoFile)
	goFile.UniqueTypes = make(map[string]struct{})
	var sb = new(bytes.Buffer)
	enc := json.NewEncoder(sb)

	for _, n := range t.List() {
		switch n.Type {
		case C(TokenSYNTAX):
			err := processSyntax(goFile, n)
			if err != nil {
				return "", err
			}
		case C(TokenPACKAGE):
			err := processPackage(goFile, n)
			if err != nil {
				return "", err
			}
		case C(TokenOPTION):
			err := processOption(goFile, n)
			if err != nil {
				return "", err
			}
		case C(TokenENUM):
			err := processEnum(goFile, n)
			if err != nil {
				return "", err
			}
		case C(TokenMESSAGE):
			err := processMessage(goFile, n)
			if err != nil {
				return "", err
			}
		default:
			invalidErr := fmt.Errorf("invalid top-level token: %d -- %s", n.Type, toString(n.Value))
			err := enc.Encode(goFile)
			if err != nil {
				return (R)(sb.String()), errors.Join(invalidErr, err)
			}
			return (R)(sb.String()), invalidErr
		}
	}
	err := enc.Encode(goFile)
	if err != nil {
		return (R)(sb.String()), err
	}

	return (R)(sb.String()), nil
}

func processSyntax[C ProtoToken, T byte](goFile *GoFile, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if n.Edges[0].Type != C(TokenVALUE) || toString(n.Edges[0].Value) != supportedSyntax {
		return ErrInvalidSyntax
	}
	return nil
}

func processPackage[C ProtoToken, T byte](goFile *GoFile, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if n.Edges[0].Type != C(TokenVALUE) {
		return ErrMissingPackage
	}
	goFile.Package = toString(n.Edges[0].Value)
	return nil
}

func processOption[C ProtoToken, T byte](goFile *GoFile, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}
	if n.Edges[0].Type != (C)(TokenTYPE) ||
		toString(n.Edges[0].Value) != supportedOption {
		return ErrInvalidOption
	}

	if len(n.Edges[0].Edges) != 1 ||
		n.Edges[0].Edges[0].Type != (C)(TokenVALUE) {
		return ErrInvalidTokenType
	}
	goFile.Path = toString(n.Edges[0].Edges[0].Value)
	return nil

}

func processEnum[C ProtoToken, T byte](goFile *GoFile, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 && len(n.Edges[0].Edges) == 0 {
		return ErrInvalidEdgeAmount
	}
	if n.Edges[0].Type != (C)(TokenTYPE) {
		return ErrInvalidTokenType
	}

	var jerr []error
	enum := new(GoEnum)
	enum.uniqueIDs = make(map[int]struct{})
	enum.uniqueNames = make(map[string]struct{})

	name := toString(n.Edges[0].Value)
	if _, ok := goFile.UniqueTypes[name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
	}
	goFile.UniqueTypes[name] = struct{}{}
	enum.ProtoName = name

	for _, e := range n.Edges[0].Edges {
		err := processEnumFields(enum, e)
		if err != nil {
			jerr = append(jerr, err)
		}
	}

	if len(jerr) > 0 {
		return errors.Join(jerr...)
	}
	goFile.Enums = append(goFile.Enums, enum)
	return nil
}
func processEnumFields[C ProtoToken, T byte](enum *GoEnum, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if n.Type != (C)(TokenVALUE) {
		return ErrInvalidTokenType
	}

	idx, err := strconv.Atoi(toString(n.Value))
	if err != nil {
		return err
	}
	name := toString(n.Edges[0].Value)

	if _, ok := enum.uniqueIDs[idx]; ok {
		return fmt.Errorf("%w: %d", ErrAlreadyExistsID, idx)
	}
	enum.uniqueIDs[idx] = struct{}{}
	if _, ok := enum.uniqueNames[name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
	}
	enum.uniqueNames[name] = struct{}{}

	f := EnumField{
		Index:     idx,
		ProtoName: name,
	}
	enum.Fields = append(enum.Fields, f)
	return nil
}

func processMessage[C ProtoToken, T byte](goFile *GoFile, n *parse.Node[C, T]) error {
	if len(n.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	var jerr []error
	goType := new(GoType)
	goType.uniqueIDs = make(map[int]struct{})
	goType.uniqueNames = make(map[string]struct{})

	goType.Name = toString(n.Edges[0].Value)
	if _, ok := goFile.UniqueTypes[goType.Name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, goType.Name)
	}
	goFile.UniqueTypes[goType.Name] = struct{}{}

	for _, e := range n.Edges[0].Edges {
		err := processMessageFields(goType, goFile, e)
		if err != nil {
			jerr = append(jerr, err)
		}
	}
	if len(jerr) > 0 {
		return errors.Join(jerr...)
	}
	goFile.Types = append(goFile.Types, goType)
	return nil
}

func processMessageFields[C ProtoToken, T byte](goType *GoType, goFile *GoFile, n *parse.Node[C, T]) error {
	field := new(GoField)

	switch n.Type {
	case (C)(TokenVALUE):
		idx, err := strconv.Atoi(toString(n.Value))
		if err != nil {
			return err
		}
		if _, ok := goType.uniqueIDs[idx]; ok {
			return fmt.Errorf("%w: %d", ErrAlreadyExistsID, idx)
		}
		goType.uniqueIDs[idx] = struct{}{}
		field.ProtoIndex = idx
		for _, e := range n.Edges {
			switch e.Type {
			case (C)(TokenMESSAGE):
				return processMessage(goFile, e)
			case C(TokenIDENT):
				name := toString(e.Value)
				if _, ok := goType.uniqueNames[name]; ok {
					return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
				}
				goType.uniqueNames[name] = struct{}{}
				field.ProtoName = name
			case C(TokenTYPE):
				field.ProtoType = toString(e.Value)
			case C(TokenREPEATED):
				field.IsRepeated = true
			case C(TokenOPTIONAL):
				field.IsOptional = true
			}
		}
	case (C)(TokenMESSAGE):
		return processMessage(goFile, n)
	}
	goType.Fields = append(goType.Fields, *field)
	return nil

}
