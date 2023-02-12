package generic

import "testing"

func TestGeneric(t *testing.T) {
	status := Ok
	gen := Generic{
		BoolField:   true,
		Unsigned32:  12,
		Unsigned64:  32,
		Signed32:    -12,
		Signed64:    -32,
		Int32:       12546734,
		Int64:       -15675732,
		Fixed32:     45445645,
		Fixed64:     112315435323,
		Sfixed32:    -12454,
		Sfixed64:    -12434324,
		Float32:     1.5,
		Float64:     1.6546456,
		Varchar:     "something",
		ByteSlice:   []byte("yep"),
		IntSlice:    []uint64{1, 2, 3},
		EnumField:   &status,
		InnerStruct: []Short{{Ok: true}},
	}

	buf := gen.Bytes()
	t.Log(string(buf))
	t.Error()

}
