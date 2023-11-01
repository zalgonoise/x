package cronlex

type Token uint8

const (
	TokenEOF Token = iota
	TokenError
	TokenAlphaNum
	TokenStar
	TokenComma
	TokenDash
	TokenSlash
	TokenAt
	TokenSpace
)
