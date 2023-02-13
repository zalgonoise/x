package generic

import (
	"strings"
	"testing"

	pbgen "github.com/zalgonoise/x/encoding/proto/protofile/testdata/google_protobuf/generic"
	"google.golang.org/protobuf/proto"
)

func TestGoogleProtobuf(t *testing.T) {
	status := pbgen.Status_ok
	gen := pbgen.Generic{
		BoolField:   true,
		Unsigned_32: 12,
		Unsigned_64: 32,
		Signed_32:   -12,
		Signed_64:   -32,
		Int_32:      12546734,
		Int_64:      -15675732,
		Fixed_32:    45445645,
		Fixed_64:    112315435323,
		Sfixed_32:   -12454,
		Sfixed_64:   -12434324,
		Float_32:    1.5,
		Float_64:    1.6546456,
		Varchar:     "something",
		ByteSlice:   []byte("yep"),
		IntSlice:    []uint64{1, 2, 3},
		EnumField:   &status,
		InnerStruct: []*pbgen.Generic_Short{{Ok: true}},
	}

	b, err := proto.Marshal(&gen)
	if err != nil {
		t.Error(err)
	}
	var sb = new(strings.Builder)
	var args = make([]any, 0, len(b))
	for idx, byt := range b {
		if idx%4 == 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("%08b\t")
		args = append(args, byt)
	}
	t.Logf(sb.String(), args...)
	t.Error()
}

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
	var sb = new(strings.Builder)
	var args = make([]any, 0, len(buf))
	for idx, byt := range buf {
		if idx%4 == 0 {
			sb.WriteString("\n")
		}
		sb.WriteString("%08b\t")
		args = append(args, byt)
	}
	t.Logf(sb.String(), args...)

	gen2, err := ToGeneric(buf)
	if err != nil {
		t.Error(err)
	}
	t.Log(string(buf))
	t.Log(gen2)
	t.Error()

}

func TestBinaryOutput(t *testing.T) {
	t.Run("Boolean", func(t *testing.T) {
		protobufGen := pbgen.Generic{BoolField: true}
		gen := Generic{BoolField: true}

		b, err := proto.Marshal(&protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		for idx, pbbyte := range b {
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch: wanted %08b ; got %08b", pbbyte, buf[idx])
			}
		}
	})
	t.Run("Uint32", func(t *testing.T) {
		protobufGen := &pbgen.Generic{Unsigned_32: 12}
		gen := &Generic{Unsigned32: 12}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 16
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
			}
		}

		for idx, pbbyte := range b {
			if pbbyte != buf[idx+i] {
				t.Errorf("byte output mismatch: wanted %08b ; got %08b", pbbyte, buf[idx+i])
			}
		}
	})

	t.Run("Uint64", func(t *testing.T) {
		protobufGen := &pbgen.Generic{Unsigned_64: 32}
		gen := &Generic{Unsigned64: 32}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 24
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
			}
		}

		for idx, pbbyte := range b {
			if pbbyte != buf[idx+i] {
				t.Errorf("byte output mismatch: wanted %08b ; got %08b", pbbyte, buf[idx+i])
			}
		}
	})

	t.Run("Sint32", func(t *testing.T) {
		protobufGen := &pbgen.Generic{Signed_32: -12}
		gen := &Generic{Signed32: -12}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 32
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
			}
		}

		for idx, pbbyte := range b {
			if pbbyte != buf[idx+i] {
				t.Errorf("byte output mismatch: wanted %08b ; got %08b", pbbyte, buf[idx+i])
			}
		}
	})

	t.Run("Sint64", func(t *testing.T) {
		protobufGen := &pbgen.Generic{Signed_64: -32}
		gen := &Generic{Signed64: -32}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 40
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
			}
		}

		for idx, pbbyte := range b {
			if pbbyte != buf[idx+i] {
				t.Errorf("byte output mismatch: wanted %08b ; got %08b", pbbyte, buf[idx+i])
			}
		}
	})
}
