package scan

import "go/token"

type SkipExtractor struct{}

func (e *SkipExtractor) Do(pos token.Pos, tok token.Token, lit string) Extractor { return e }
func (e *SkipExtractor) Done() bool                                              { return true }
