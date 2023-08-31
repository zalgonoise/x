package schedule

type token uint8

const (
	tokenEOF token = iota
	tokenError
	tokenAlphanum
	tokenStar
	tokenComma
	tokenDash
	tokenSlash
	tokenAt
	_
	_
	tokenSpace
)
