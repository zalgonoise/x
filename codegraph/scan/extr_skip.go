package scan

import "go/token"

type SkipExtractor struct{}

func (e *SkipExtractor) Do(tok token.Token, lit string) Extractor { return e }
func (e *SkipExtractor) Done() bool                               { return true }
