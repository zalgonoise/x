package errs

var (
	spaceSeparator = []rune{' '}
	colonSeparator = []rune{':', ' '}
)

type Domain string

func (e Domain) Error() string { return (string)(e) }

type Kind string

func (e Kind) Error() string { return (string)(e) }

type Entity string

func (e Entity) Error() string { return (string)(e) }

func Sentinel(kind Kind, entity Entity) error {
	return newSentinel("", kind, entity)
}

func WithDomain(domain Domain, kind Kind, entity Entity) error {
	return newSentinel(domain, kind, entity)
}

func withSpace[K, E ~string](first K, last E) string {
	s := make([]rune, len(first)+len(spaceSeparator)+len(last))

	n := copy(s, []rune(first))
	n += copy(s[n:], spaceSeparator)
	copy(s[n:], []rune(last))

	return string(s)
}

func withColon[K, E ~string](first K, last E) string {
	s := make([]rune, len(first)+len(colonSeparator)+len(last))

	n := copy(s, []rune(first))
	n += copy(s[n:], colonSeparator)
	n += copy(s[n:], []rune(last))

	return string(s)
}
