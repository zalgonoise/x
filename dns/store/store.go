package store

// Record defines the elements of a DNS Record
//
// TODO: add the elements necessary to comprehend the most common
// DNS records' elements
type Record struct {
	Type string
	Name string
	Addr string
}

type RecordBuilder struct {
	t    string
	name string
	addr string
}

func New() *RecordBuilder {
	return &RecordBuilder{}
}

func (b *RecordBuilder) Type(s string) *RecordBuilder {
	b.t = s
	return b
}

func (b *RecordBuilder) Name(s string) *RecordBuilder {
	b.name = s
	return b
}

func (b *RecordBuilder) Addr(s string) *RecordBuilder {
	b.addr = s
	return b
}

func (b *RecordBuilder) Build() *Record {
	return &Record{
		Name: b.name,
		Type: b.t,
		Addr: b.addr,
	}
}
