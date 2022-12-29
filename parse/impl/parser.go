package impl

import "github.com/zalgonoise/x/parse"

func Run[C TextToken, T rune, R string](s []T) (R, error) {
	return parse.Run(
		s,
		initState[C, T],
		initParse[C, T],
		processFn[C, T, R],
	)
}
