package dns

type RecordType uint16

const (
	TypeNone  RecordType = 0
	TypeA     RecordType = 1
	TypeCNAME RecordType = 5
	TypeAAAA  RecordType = 28
)

var (
	RecordTypeKeys = map[RecordType]uint16{
		TypeNone:  0,
		TypeA:     1,
		TypeCNAME: 5,
		TypeAAAA:  28,
	}
	RecordTypeVals = map[uint16]RecordType{
		0:  TypeNone,
		1:  TypeA,
		5:  TypeCNAME,
		28: TypeAAAA,
	}
	RecordTypeStrings = map[RecordType]string{
		TypeNone:  "",
		TypeA:     "A",
		TypeCNAME: "CNAME",
		TypeAAAA:  "AAAA",
	}
)

func (t RecordType) String() string {
	return RecordTypeStrings[t]
}
