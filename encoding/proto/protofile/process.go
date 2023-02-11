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
)

type GoField struct {
	IsRepeated bool   `json:"is_repeated,omitempty"`
	IsOptional bool   `json:"is_optional,omitempty"`
	IsStruct   bool   `json:"is_struct,omitempty"`
	GoType     string `json:"go_type,omitempty"`
	GoName     string `json:"go_name,omitempty"`
	ProtoType  string `json:"proto_type,omitempty"`
	ProtoIndex int    `json:"proto_index,omitempty"`
	ProtoName  string `json:"proto_name,omitempty"`
}

type EnumField struct {
	Index     int    `json:"index,omitempty"`
	GoName    string `json:"go_name,omitempty"`
	ProtoName string `json:"proto_name,omitempty"`
}

type GoType struct {
	Name   string    `json:"name,omitempty"`
	Fields []GoField `json:"fields,omitempty"`
}

type GoEnum struct {
	ProtoName string      `json:"proto_name,omitempty"`
	GoName    string      `json:"go_name,omitempty"`
	Fields    []EnumField `json:"fields,omitempty"`
}

type GoFile struct {
	Path        string    `json:"path,omitempty"`
	Package     string    `json:"package,omitempty"`
	Types       []*GoType `json:"types,omitempty"`
	Enums       []*GoEnum `json:"enums,omitempty"`
	CustomTypes []string  `json:"custom_types,omitempty"`
}

func processFn[C ProtoToken, T byte, R string](t *parse.Tree[C, T]) (R, error) {
	var goFile = new(GoFile)
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

	enum.ProtoName = toString(n.Edges[0].Value)

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
	f := EnumField{
		Index:     idx,
		ProtoName: toString(n.Edges[0].Value),
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
	goType.Name = toString(n.Edges[0].Value)
	goFile.CustomTypes = append(goFile.CustomTypes, goType.Name)

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
		field.ProtoIndex = idx
		for _, e := range n.Edges {
			switch e.Type {
			case (C)(TokenMESSAGE):
				return processMessage(goFile, e)
			case C(TokenIDENT):
				field.ProtoName = toString(e.Value)
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
