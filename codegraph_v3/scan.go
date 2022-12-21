package codegraph

import (
	"fmt"
	"go/scanner"
	"go/token"
	"os"
)

type GoToken struct {
	Pos token.Pos
	Tok token.Token
	Lit string
}

func Extract(path string) ([]GoToken, error) {
	var (
		fs  = token.NewFileSet()
		sc  = scanner.Scanner{}
		err error
	)

	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	file := fs.AddFile(path, fs.Base(), len(b))

	sc.Init(file, b, func(pos token.Position, msg string) {
		if err == nil {
			err = fmt.Errorf("error in %s: %s", pos.String(), msg)
			return
		}
		err = fmt.Errorf("error in %s: %s ; %w", pos.String(), msg, err)
	}, scanner.Mode(1))

	if err != nil {
		return nil, err
	}

	var tokens = make([]GoToken, 0, 1024)

	for {
		pos, tok, lit := sc.Scan()
		tokens = append(tokens, GoToken{
			Pos: pos,
			Tok: tok,
			Lit: lit,
		})
		if tok == token.EOF {
			break // end of GoFile
		}
	}

	return tokens, nil
}

func Explore(goCode []byte) ([]GoToken, error) {
	var (
		fs  = token.NewFileSet()
		sc  = scanner.Scanner{}
		err error
	)

	file := fs.AddFile("", fs.Base(), len(goCode))

	sc.Init(file, goCode, func(pos token.Position, msg string) {
		if err == nil {
			err = fmt.Errorf("error in %s: %s", pos.String(), msg)
			return
		}
		err = fmt.Errorf("error in %s: %s ; %w", pos.String(), msg, err)
	}, scanner.Mode(1))

	if err != nil {
		return nil, err
	}

	var tokens = make([]GoToken, 0, 1024)

	for {
		pos, tok, lit := sc.Scan()
		tokens = append(tokens, GoToken{
			Pos: pos,
			Tok: tok,
			Lit: lit,
		})
		if tok == token.EOF {
			break // end of GoFile
		}
	}

	return tokens, nil
}
