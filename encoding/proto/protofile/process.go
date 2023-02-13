package protofile

import (
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
	ErrEmptyName         = errors.New("name cannot be empty")
	ErrEmptyPath         = errors.New("go path cannot be empty")
	ErrEmptyPackage      = errors.New("package name cannot be empty")
	ErrInvalidType       = errors.New("invalid, undeclared or unsupported type")
)

func processFn[C ProtoToken, T byte, R *GoFile](t *parse.Tree[C, T]) (R, error) {
	var goFile = new(GoFile)
	goFile.UniqueTypes = make(map[string]bool)
	goFile.concreteTypes = make(map[string]ConcreteType)
	goFile.importsList = make(map[string]struct{})
	goFile.importsList["bytes"] = struct{}{}
	goFile.importsList["errors"] = struct{}{}
	goFile.importsList["io"] = struct{}{}

	for _, n := range t.List() {
		switch n.Type {
		case C(TokenSYNTAX):
			err := processSyntax(goFile, n)
			if err != nil {
				return (R)(goFile), err
			}
		case C(TokenPACKAGE):
			err := processPackage(goFile, n)
			if err != nil {
				return (R)(goFile), err
			}
		case C(TokenOPTION):
			err := processOption(goFile, n)
			if err != nil {
				return (R)(goFile), err
			}
		case C(TokenENUM):
			err := processEnum(goFile, n)
			if err != nil {
				return (R)(goFile), err
			}
		case C(TokenMESSAGE):
			err := processMessage(goFile, n)
			if err != nil {
				return (R)(goFile), err
			}
		default:
			return (R)(goFile), fmt.Errorf("invalid top-level token: %d -- %s", n.Type, toString(n.Value))
		}
	}
	return (R)(goFile), nil
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
	if goFile.Package == "" {
		return ErrEmptyPackage
	}
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
	if goFile.Path == "" {
		return ErrEmptyPath
	}
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
	if name == "" {
		return ErrEmptyName
	}
	if _, ok := goFile.UniqueTypes[name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
	}
	goFile.UniqueTypes[name] = true
	enum.ProtoName = name
	enum.GoName = fmtPascal(name)

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
	if name == "" {
		return ErrEmptyName
	}

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
		GoName:    fmtPascal(name),
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
	if goType.Name == "" {
		return ErrEmptyName
	}
	if _, ok := goFile.UniqueTypes[goType.Name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, goType.Name)
	}
	goFile.UniqueTypes[goType.Name] = false

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
				if name == "" {
					return ErrEmptyName
				}
				if _, ok := goType.uniqueNames[name]; ok {
					return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
				}
				goType.uniqueNames[name] = struct{}{}
				field.ProtoName = name
				field.GoName = fmtPascal(name)
			case C(TokenTYPE):
				field.ProtoType = toString(e.Value)
				if goType, ok := protoTypes[field.ProtoType]; ok {
					field.GoType = goType.GoType()
					if _, ok := goFile.concreteTypes[goType.GoType()]; !ok {
						goFile.concreteTypes[goType.GoType()] = goType
					}
					if goType == Float32 || goType == Float64 {
						goFile.importsList["math"] = struct{}{}
						goFile.importsList["encoding/binary"] = struct{}{}
					}
					if n := allocMap[goType]; n > 0 {
						goFile.minAlloc += n
					}
					continue
				}
				if _, ok := goFile.UniqueTypes[field.ProtoType]; ok {
					field.GoType = field.ProtoType
					continue
				}
				return ErrInvalidType
			case C(TokenREPEATED):
				field.IsRepeated = true
			case C(TokenOPTIONAL):
				field.IsOptional = true
			}
		}
	case (C)(TokenMESSAGE):
		return processMessage(goFile, n)
	}

	var wireType int
	if t, ok := goFile.concreteTypes[field.GoType]; ok {
		wireType = t.WireType()
	} else {
		if v, ok := goFile.UniqueTypes[field.GoName]; ok {
			if v {
				wireType = 0
			} else {
				wireType = 2
			}
		} else {
			wireType = 2
		}
	}

	field.idAndWire = IDAndWire{
		ID:   field.ProtoIndex,
		Wire: wireType,
		Name: field.GoName,
	}

	goType.Fields = append(goType.Fields, *field)
	return nil

}
