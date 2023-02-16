package protofile

import (
	"errors"
	"fmt"
	"os"
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

func processFn[C ProtoToken, T byte, R int](tree *parse.Tree[C, T]) (R, error) {
	var (
		goFile   = NewGoFile()
		n      R = -1
	)

	for _, node := range tree.List() {
		switch node.Type {
		case C(TokenSYNTAX):
			err := processSyntax(goFile, node)
			if err != nil {
				return n, err
			}
		case C(TokenPACKAGE):
			err := processPackage(goFile, node)
			if err != nil {
				return n, err
			}
		case C(TokenOPTION):
			err := processOption(goFile, node)
			if err != nil {
				return n, err
			}
		case C(TokenENUM):
			err := processEnum(goFile, node)
			if err != nil {
				return n, err
			}
		case C(TokenMESSAGE):
			err := processMessage(goFile, node)
			if err != nil {
				return n, err
			}
		default:
			return n, fmt.Errorf("invalid top-level token: %d -- %s", node.Type, toString(node.Value))
		}
	}

	return createPbGo[R](goFile)
}

func createPbGo[T int](goFile *GoFile) (T, error) {
	err := os.MkdirAll(goFile.Path, os.ModeDir|(OS_USER_RWX|OS_ALL_R))
	if err != nil {
		return -1, fmt.Errorf("failed to create folder in %s: %w", goFile.Path, err)
	}
	f, err := os.Create(goFile.Path + "/" + goFile.Package + ".pb.go")
	if err != nil {
		return -1, fmt.Errorf("failed to create .pb.go file in %s: %w", goFile.Path, err)
	}

	n, err := f.Write(goFile.GoBytes())
	if err != nil {
		return -1, fmt.Errorf("failed to write data to .pb.go file in %s: %w", goFile.Path, err)
	}
	return (T)(n), nil
}

func processSyntax[C ProtoToken, T byte](goFile *GoFile, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if node.Edges[0].Type != C(TokenVALUE) || toString(node.Edges[0].Value) != supportedSyntax {
		return ErrInvalidSyntax
	}
	return nil
}

func processPackage[C ProtoToken, T byte](goFile *GoFile, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if node.Edges[0].Type != C(TokenVALUE) {
		return ErrMissingPackage
	}
	goFile.Package = toString(node.Edges[0].Value)
	if goFile.Package == "" {
		return ErrEmptyPackage
	}
	return nil
}

func processOption[C ProtoToken, T byte](goFile *GoFile, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}
	if node.Edges[0].Type != (C)(TokenTYPE) ||
		toString(node.Edges[0].Value) != supportedOption {
		return ErrInvalidOption
	}

	if len(node.Edges[0].Edges) != 1 ||
		node.Edges[0].Edges[0].Type != (C)(TokenVALUE) {
		return ErrInvalidTokenType
	}
	goFile.Path = toString(node.Edges[0].Edges[0].Value)
	if goFile.Path == "" {
		return ErrEmptyPath
	}
	return nil

}

func processEnum[C ProtoToken, T byte](goFile *GoFile, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 && len(node.Edges[0].Edges) == 0 {
		return ErrInvalidEdgeAmount
	}
	if node.Edges[0].Type != (C)(TokenTYPE) {
		return ErrInvalidTokenType
	}

	var (
		jerr []error
		enum = NewGoEnum()
	)

	name := toString(node.Edges[0].Value)
	if name == "" {
		return ErrEmptyName
	}
	if _, ok := goFile.UniqueTypes[name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, name)
	}

	goFile.UniqueTypes[name] = true
	enum.ProtoName = name
	enum.GoName = fmtPascal(name)

	for _, e := range node.Edges[0].Edges {
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

func processEnumFields[C ProtoToken, T byte](enum *GoEnum, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	if node.Type != (C)(TokenVALUE) {
		return ErrInvalidTokenType
	}

	idx, err := strconv.Atoi(toString(node.Value))
	if err != nil {
		return err
	}
	name := toString(node.Edges[0].Value)
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

func processMessage[C ProtoToken, T byte](goFile *GoFile, node *parse.Node[C, T]) error {
	if len(node.Edges) != 1 {
		return ErrInvalidEdgeAmount
	}

	var (
		jerr   []error
		goType = NewGoType()
	)

	goType.Name = toString(node.Edges[0].Value)
	if goType.Name == "" {
		return ErrEmptyName
	}
	if _, ok := goFile.UniqueTypes[goType.Name]; ok {
		return fmt.Errorf("%w: %s", ErrAlreadyExistsName, goType.Name)
	}
	goFile.UniqueTypes[goType.Name] = false

	for _, e := range node.Edges[0].Edges {
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

func processMessageFields[C ProtoToken, T byte](goType *GoType, goFile *GoFile, node *parse.Node[C, T]) error {
	field := new(GoField)

	switch node.Type {
	case (C)(TokenVALUE):
		idx, err := strconv.Atoi(toString(node.Value))
		if err != nil {
			return err
		}
		if _, ok := goType.uniqueIDs[idx]; ok {
			return fmt.Errorf("%w: %d", ErrAlreadyExistsID, idx)
		}
		goType.uniqueIDs[idx] = struct{}{}
		field.ProtoIndex = idx
		for _, e := range node.Edges {
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
		return processMessage(goFile, node)
	}

	var wireType int
	if t, ok := goFile.concreteTypes[field.GoType]; ok {
		wireType = t.WireType()
	} else if isEnum, ok := goFile.UniqueTypes[field.ProtoType]; ok {
		if isEnum {
			wireType = 0
		} else {
			wireType = 2
		}
	}

	field.idAndWire = IDAndWire{
		ID:   field.ProtoIndex,
		Wire: wireType,
		Name: field.GoName,
	}

	if header := field.idAndWire.Header(); header == 0 {
		return fmt.Errorf("field %s generates a header larger than 255: %d -- please generate a shorter message", field.GoName, (field.idAndWire.ID<<3)|field.idAndWire.Wire)
	}

	goType.Fields = append(goType.Fields, *field)
	return nil

}
