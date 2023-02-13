package generic

import (
	"bytes"
	"reflect"
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
	gen2, err := ToGeneric(buf)
	buf2 := gen2.Bytes()
	if err != nil {
		t.Error(err)
	}
	if !reflect.DeepEqual(gen, gen2) {
		t.Errorf("output mismatch error: %v, %v", gen, gen2)
	}

	if len(buf) != len(buf2) {
		t.Errorf("buffer length mismatch on double conversion: wanted %d ; got %d", len(buf), len(buf2))
	}
	for i := 0; i < len(buf); i++ {
		if buf[i] != buf2[i] {
			t.Errorf("byte output mismatch on double conversion, on index %d: wanted %d, got %d", i, buf[i], buf2[i])
		}
	}

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
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
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
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
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
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
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
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
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
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Int32", func(t *testing.T) {
		var wants int32 = 12546734
		protobufGen := &pbgen.Generic{Int_32: wants}
		gen := &Generic{Int32: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 48
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if int32((v>>1)^-(v&1)) != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int32((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Int64", func(t *testing.T) {
		var wants int64 = -15675732
		protobufGen := &pbgen.Generic{Int_64: wants}
		gen := &Generic{Int64: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 56
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if int64((v>>1)^-(v&1)) != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int64((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Fixed32", func(t *testing.T) {
		var wants uint32 = 45445645
		protobufGen := &pbgen.Generic{Fixed_32: wants}
		gen := &Generic{Fixed32: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 64
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if uint32(v) != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int32((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Fixed64", func(t *testing.T) {
		var wants uint64 = 112315435323
		protobufGen := &pbgen.Generic{Fixed_64: wants}
		gen := &Generic{Fixed64: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 72
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if v != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int64((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Sfixed32", func(t *testing.T) {
		var wants int32 = -12454
		protobufGen := &pbgen.Generic{Sfixed_32: wants}
		gen := &Generic{Sfixed32: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 80
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if int32((v>>1)^-(v&1)) != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int32((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Sfixed64", func(t *testing.T) {
		var wants int64 = -12434324
		protobufGen := &pbgen.Generic{Sfixed_64: wants}
		gen := &Generic{Sfixed64: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 88
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Logf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		bbuf := bytes.NewBuffer(buf[1:len(buf)])
		v, err := decodeVarint(bbuf)
		if err != nil {
			t.Error(err)
		}
		if int64((v>>1)^-(v&1)) != wants {
			t.Errorf("output mismatch error: wanted %v ; got %v", wants, int64((v>>1)^-(v&1)))
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Float32", func(t *testing.T) {
		var wants float32 = 1.5
		protobufGen := &pbgen.Generic{Float_32: wants}
		gen := &Generic{Float32: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 101
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Float64", func(t *testing.T) {
		var wants float64 = 1.6546456
		protobufGen := &pbgen.Generic{Float_64: wants}
		gen := &Generic{Float64: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 105
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Varchar", func(t *testing.T) {
		var wants string = "something"
		protobufGen := &pbgen.Generic{Varchar: wants}
		gen := &Generic{Varchar: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 114
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("Varchar", func(t *testing.T) {
		var wants []byte = []byte("yep")
		protobufGen := &pbgen.Generic{ByteSlice: wants}
		gen := &Generic{ByteSlice: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 122
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("IntSlice", func(t *testing.T) {
		wants := []uint64{1, 2, 3}
		protobufGen := &pbgen.Generic{IntSlice: wants}
		gen := &Generic{IntSlice: wants}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 128
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}

		for i, j := 0, 1; i < len(wants); i, j = i+1, j+2 {
			if uint64(buf[j]) != wants[i] {
				t.Errorf("output mismatch error: wanted %v ; got %v", wants[i], buf[j])
			}
		}
	})

	t.Run("EnumField", func(t *testing.T) {
		status := Ok
		protobufGen := &pbgen.Generic{EnumField: pbgen.Status_ok.Enum()}
		gen := &Generic{EnumField: &status}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 136
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx+i])
			}
		}
		if buf[0] != header {
			t.Errorf("invalid header: wanted %d, got %d", header, buf[0])
		}
		if buf[1] != 1 {
			// Ok = 1
			t.Errorf("invalid value: wanted %d, got %d", 1, buf[1])
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

	t.Run("InnerStruct", func(t *testing.T) {
		protobufGen := &pbgen.Generic{InnerStruct: []*pbgen.Generic_Short{{Ok: true}}}
		gen := &Generic{InnerStruct: []Short{{Ok: true}}}

		b, err := proto.Marshal(protobufGen)
		if err != nil {
			t.Error(err)
		}
		buf := gen.Bytes()

		// seek to pos
		const header = 146
		var i int
		for idx, byt := range buf {
			if byt == header {
				i = idx
				break
			}
		}
		buf = buf[i:]
		var bi int
		for idx, byt := range b {
			if byt == header {
				bi = idx
				break
			}
		}
		b = b[bi:]

		for idx, pbbyte := range b {
			if idx >= len(buf) {
				break
			}
			if pbbyte != buf[idx] {
				t.Errorf("byte output mismatch on index %d: wanted %08b ; got %08b", idx, pbbyte, buf[idx])
			}
		}
		t.Errorf("\n%v\n%v\n", buf, b)
	})

}
