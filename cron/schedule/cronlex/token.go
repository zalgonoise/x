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

var tokenStrings = [...]string{
	"TokenEOF",
	"TokenError",
	"TokenAlphaNum",
	"TokenStar",
	"TokenComma",
	"TokenDash",
	"TokenSlash",
	"TokenAt",
	"TokenSpace",
}

func (t Token) String() string {
	return tokenStrings[t]
}
