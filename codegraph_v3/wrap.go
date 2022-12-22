package codegraph

func WrapType(ts []GoToken, kind LogicBlockKind) []GoToken {
	filter := filterMap[kind]
	if filter == nil {
		return ts
	}

	tokens := make([]GoToken, len(ts)+2, len(ts)+2)
	tokens[0] = GoToken{Tok: filter.init}
	for idx, t := range ts {
		tokens[idx+1] = t
	}
	tokens[len(tokens)-1] = GoToken{Tok: filter.closer}
	return tokens
}
