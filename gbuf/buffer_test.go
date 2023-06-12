package gbuf_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"testing"

	"github.com/zalgonoise/x/gbuf"
)

const N = 10000       // make this bigger for a larger (and slower) test
var testString string // test data for write tests
var testBytes []byte  // test data; same as testString but as a slice.

type negativeReader struct{}

func (r *negativeReader) Read([]byte) (int, error) { return -1, nil }

func init() {
	testBytes = make([]byte, N)
	for i := 0; i < N; i++ {
		testBytes[i] = 'a' + byte(i%26)
	}
	testString = string(testBytes)
}

// Verify that contents of buf match the string s.
func check(t *testing.T, name string, buf *gbuf.Buffer[byte], s string) {
	byteValue := buf.Value()
	str := string(byteValue)
	if buf.Len() != len(byteValue) {
		t.Errorf("%s: buf.Len() == %d, len(buf.Value()) == %d", name, buf.Len(), len(byteValue))
	}

	if buf.Len() != len(str) {
		t.Errorf("%s: buf.Len() == %d, len(buf.String()) == %d", name, buf.Len(), len(str))
	}

	if buf.Len() != len(s) {
		t.Errorf("%s: buf.Len() == %d, len(s) == %d", name, buf.Len(), len(s))
	}

	if string(byteValue) != s {
		t.Errorf("%s: string(buf.Value()) == %q, s == %q", name, string(byteValue), s)
	}
}

// Fill buf through n writes of string fus.
// The initial contents of buf corresponds to the string s;
// the result is the final contents of buf returned as a string.
func fillString(t *testing.T, name string, buf *gbuf.Buffer[byte], s string, n int, fus string) string {
	check(t, name+" (fill 1)", buf, s)
	for ; n > 0; n-- {
		m, err := buf.Write([]byte(fus))
		if m != len(fus) {
			t.Errorf(name+" (fill 2): m == %d, expected %d", m, len(fus))
		}
		if err != nil {
			t.Errorf(name+" (fill 3): err should always be nil, found err == %s", err)
		}
		s += fus
		check(t, name+" (fill 4)", buf, s)
	}
	return s
}

// Fill buf through n writes of byte slice fub.
// The initial contents of buf corresponds to the string s;
// the result is the final contents of buf returned as a string.
func fillBytes(t *testing.T, testname string, buf *gbuf.Buffer[byte], s string, n int, fub []byte) string {
	check(t, testname+" (fill 1)", buf, s)
	for ; n > 0; n-- {
		m, err := buf.Write(fub)
		if m != len(fub) {
			t.Errorf(testname+" (fill 2): m == %d, expected %d", m, len(fub))
		}
		if err != nil {
			t.Errorf(testname+" (fill 3): err should always be nil, found err == %s", err)
		}
		s += string(fub)
		check(t, testname+" (fill 4)", buf, s)
	}
	return s
}

func TestNewBuffer(t *testing.T) {
	buf := gbuf.NewBuffer(testBytes)
	check(t, "NewBuffer", buf, testString)
}

// Empty buf through repeated reads into fub.
// The initial contents of buf corresponds to the string s.
func empty(t *testing.T, testname string, buf *gbuf.Buffer[byte], s string, fub []byte) {
	check(t, testname+" (empty 1)", buf, s)

	for {
		n, err := buf.Read(fub)
		if n == 0 {
			break
		}
		if err != nil {
			t.Errorf(testname+" (empty 2): err should always be nil, found err == %s", err)
		}
		s = s[n:]
		check(t, testname+" (empty 3)", buf, s)
	}

	check(t, testname+" (empty 4)", buf, "")
}

func TestBasicOperations(t *testing.T) {
	var buf gbuf.Buffer[byte]

	for i := 0; i < 5; i++ {
		check(t, "TestBasicOperations (1)", &buf, "")

		buf.Reset()
		check(t, "TestBasicOperations (2)", &buf, "")

		buf.Truncate(0)
		check(t, "TestBasicOperations (3)", &buf, "")

		n, err := buf.Write(testBytes[0:1])
		if want := 1; err != nil || n != want {
			t.Errorf("Write: got (%d, %v), want (%d, %v)", n, err, want, nil)
		}
		check(t, "TestBasicOperations (4)", &buf, "a")

		_ = buf.WriteItem(testString[1])
		check(t, "TestBasicOperations (5)", &buf, "ab")

		n, err = buf.Write(testBytes[2:26])
		if want := 24; err != nil || n != want {
			t.Errorf("Write: got (%d, %v), want (%d, %v)", n, err, want, nil)
		}
		check(t, "TestBasicOperations (6)", &buf, testString[0:26])

		buf.Truncate(26)
		check(t, "TestBasicOperations (7)", &buf, testString[0:26])

		buf.Truncate(20)
		check(t, "TestBasicOperations (8)", &buf, testString[0:20])

		empty(t, "TestBasicOperations (9)", &buf, testString[0:20], make([]byte, 5))
		empty(t, "TestBasicOperations (10)", &buf, "", make([]byte, 100))

		_ = buf.WriteItem(testString[1])
		c, err := buf.ReadItem()
		if want := testString[1]; err != nil || c != want {
			t.Errorf("ReadItem: got (%q, %v), want (%q, %v)", c, err, want, nil)
		}
		c, err = buf.ReadItem()
		if err != io.EOF {
			t.Errorf("ReadItem: got (%q, %v), want (%q, %v)", c, err, byte(0), io.EOF)
		}
	}
}

func TestLargeStringWrites(t *testing.T) {
	var buf gbuf.Buffer[byte]
	limit := 30
	if testing.Short() {
		limit = 9
	}
	for i := 3; i < limit; i += 3 {
		s := fillString(t, "TestLargeWrites (1)", &buf, "", 5, testString)
		empty(t, "TestLargeStringWrites (2)", &buf, s, make([]byte, len(testString)/i))
	}
	check(t, "TestLargeStringWrites (3)", &buf, "")
}

func TestLargeByteWrites(t *testing.T) {
	var buf gbuf.Buffer[byte]
	limit := 30
	if testing.Short() {
		limit = 9
	}
	for i := 3; i < limit; i += 3 {
		s := fillBytes(t, "TestLargeWrites (1)", &buf, "", 5, testBytes)
		empty(t, "TestLargeByteWrites (2)", &buf, s, make([]byte, len(testString)/i))
	}
	check(t, "TestLargeByteWrites (3)", &buf, "")
}

func TestLargeStringReads(t *testing.T) {
	var buf gbuf.Buffer[byte]
	for i := 3; i < 30; i += 3 {
		s := fillString(t, "TestLargeReads (1)", &buf, "", 5, testString[0:len(testString)/i])
		empty(t, "TestLargeReads (2)", &buf, s, make([]byte, len(testString)))
	}
	check(t, "TestLargeStringReads (3)", &buf, "")
}

func TestLargeByteReads(t *testing.T) {
	var buf gbuf.Buffer[byte]
	for i := 3; i < 30; i += 3 {
		s := fillBytes(t, "TestLargeReads (1)", &buf, "", 5, testBytes[0:len(testBytes)/i])
		empty(t, "TestLargeReads (2)", &buf, s, make([]byte, len(testString)))
	}
	check(t, "TestLargeByteReads (3)", &buf, "")
}

func TestMixedReadsAndWrites(t *testing.T) {
	var buf gbuf.Buffer[byte]
	s := ""
	for i := 0; i < 50; i++ {
		wlen := rand.Intn(len(testString))
		if i%2 == 0 {
			s = fillString(t, "TestMixedReadsAndWrites (1)", &buf, s, 1, testString[0:wlen])
		} else {
			s = fillBytes(t, "TestMixedReadsAndWrites (1)", &buf, s, 1, testBytes[0:wlen])
		}

		rlen := rand.Intn(len(testString))
		fub := make([]byte, rlen)
		n, _ := buf.Read(fub)
		s = s[n:]
	}
	empty(t, "TestMixedReadsAndWrites (2)", &buf, s, make([]byte, buf.Len()))
}

func TestCapWithPreallocatedSlice(t *testing.T) {
	buf := gbuf.NewBuffer(make([]byte, 10))
	n := buf.Cap()
	if n != 10 {
		t.Errorf("expected 10, got %d", n)
	}
}

func TestCapWithSliceAndWrittenData(t *testing.T) {
	buf := gbuf.NewBuffer(make([]byte, 0, 10))
	_, _ = buf.Write([]byte("test"))
	n := buf.Cap()
	if n != 10 {
		t.Errorf("expected 10, got %d", n)
	}
}

func TestReadFrom(t *testing.T) {
	var buf gbuf.Buffer[byte]
	for i := 3; i < 30; i += 3 {
		s := fillBytes(t, "TestReadFrom (1)", &buf, "", 5, testBytes[0:len(testBytes)/i])
		var b gbuf.Buffer[byte]
		_, _ = b.ReadFrom(&buf)
		empty(t, "TestReadFrom (2)", &b, s, make([]byte, len(testString)))
	}
}

type panicReader struct{ panic bool }

func (r panicReader) Read(_ []byte) (int, error) {
	if r.panic {
		panic(nil)
	}
	return 0, io.EOF
}

// Make sure that an empty Buffer remains empty when
// it is "grown" before a Read that panics
func TestReadFromPanicReader(t *testing.T) {
	// First verify non-panic behaviour
	var buf gbuf.Buffer[byte]
	i, err := buf.ReadFrom(panicReader{})
	if err != nil {
		t.Fatal(err)
	}
	if i != 0 {
		t.Fatalf("unexpected return from gbuf.ReadFrom (1): got: %d, want %d", i, 0)
	}
	check(t, "TestReadFromPanicReader (1)", &buf, "")

	// Confirm that when Reader panics, the empty buffer remains empty
	var buf2 gbuf.Buffer[byte]
	defer func() {
		recover()
		check(t, "TestReadFromPanicReader (2)", &buf2, "")
	}()
	_, _ = buf2.ReadFrom(panicReader{panic: true})
}

func TestReadFromNegativeReader(t *testing.T) {
	var b gbuf.Buffer[byte]
	defer func() {
		switch err := recover().(type) {
		case nil:
			t.Fatal("gbuf.Buffer.ReadFrom didn't panic")
		case error:
			// this is the error string of errNegativeRead
			wantError := "gbuf.Buffer: reader returned negative count from Read"
			if err.Error() != wantError {
				t.Fatalf("recovered panic: got %v, want %v", err.Error(), wantError)
			}
		default:
			t.Fatalf("unexpected panic value: %#v", err)
		}
	}()

	_, _ = b.ReadFrom(new(negativeReader))
}

func TestWriteTo(t *testing.T) {
	var buf gbuf.Buffer[byte]
	for i := 3; i < 30; i += 3 {
		s := fillBytes(t, "TestWriteTo (1)", &buf, "", 5, testBytes[0:len(testBytes)/i])
		var b gbuf.Buffer[byte]
		_, _ = buf.WriteTo(&b)
		empty(t, "TestWriteTo (2)", &b, s, make([]byte, len(testString)))
	}
}

func TestNext(t *testing.T) {
	b := []byte{0, 1, 2, 3, 4}
	tmp := make([]byte, 5)
	for i := 0; i <= 5; i++ {
		for j := i; j <= 5; j++ {
			for k := 0; k <= 6; k++ {
				// 0 <= i <= j <= 5; 0 <= k <= 6
				// Check that if we start with a buffer
				// of length j at offset i and ask for
				// Next(k), we get the right bytes.
				buf := gbuf.NewBuffer(b[0:j])
				n, _ := buf.Read(tmp[0:i])
				if n != i {
					t.Fatalf("Read %d returned %d", i, n)
				}
				bb := buf.Next(k)
				want := k
				if want > j-i {
					want = j - i
				}
				if len(bb) != want {
					t.Fatalf("in %d,%d: len(Next(%d)) == %d", i, j, k, len(bb))
				}
				for l, v := range bb {
					if v != byte(l+i) {
						t.Fatalf("in %d,%d: Next(%d)[%d] = %d, want %d", i, j, k, l, v, l+i)
					}
				}
			}
		}
	}
}

var readBytesTests = []struct {
	buffer   string
	delim    byte
	expected []string
	err      error
}{
	{"", 0, []string{""}, io.EOF},
	{"a\x00", 0, []string{"a\x00"}, nil},
	{"abbbaaaba", 'b', []string{"ab", "b", "b", "aaab"}, nil},
	{"hello\x01world", 1, []string{"hello\x01"}, nil},
	{"foo\nbar", 0, []string{"foo\nbar"}, io.EOF},
	{"alpha\nbeta\ngamma\n", '\n', []string{"alpha\n", "beta\n", "gamma\n"}, nil},
	{"alpha\nbeta\ngamma", '\n', []string{"alpha\n", "beta\n", "gamma"}, io.EOF},
}

func TestReadItems(t *testing.T) {
	for _, test := range readBytesTests {
		buf := gbuf.NewBuffer([]byte(test.buffer))
		var err error
		for _, expected := range test.expected {
			var byteValue []byte
			byteValue, err = buf.ReadItems(func(b byte) bool {
				return b == test.delim
			})
			if string(byteValue) != expected {
				t.Errorf("expected %q, got %q", expected, byteValue)
			}
			if err != nil {
				break
			}
		}
		if err != test.err {
			t.Errorf("expected error %v, got %v", test.err, err)
		}
	}
}

func TestGrow(t *testing.T) {
	x := []byte{'x'}
	y := []byte{'y'}
	tmp := make([]byte, 72)
	for _, growLen := range []int{0, 100, 1000, 10000, 100000} {
		for _, startLen := range []int{0, 100, 1000, 10000, 100000} {
			xBytes := gbuf.Repeat(x, startLen)

			buf := gbuf.NewBuffer(xBytes)
			// If we read, this affects buf.off, which is good to test.
			readBytes, _ := buf.Read(tmp)
			yBytes := gbuf.Repeat(y, growLen)
			allocs := testing.AllocsPerRun(100, func() {
				buf.Grow(growLen)
				_, _ = buf.Write(yBytes)
			})
			// Check no allocation occurs in write, as long as we're single-threaded.
			if allocs != 0 {
				t.Errorf("allocation occurred during write")
			}
			// Check that buffer has correct data.
			if !bytes.Equal(buf.Value()[0:startLen-readBytes], xBytes[readBytes:]) {
				t.Errorf("bad initial data at %d %d", startLen, growLen)
			}
			if !bytes.Equal(buf.Value()[startLen-readBytes:startLen-readBytes+growLen], yBytes) {
				t.Errorf("bad written data at %d %d", startLen, growLen)
			}
		}
	}
}

func TestGrowOverflow(t *testing.T) {
	defer func() {
		if err := recover(); err != gbuf.ErrTooLarge {
			t.Errorf("after too-large Grow, recover() = %v; want %v", err, gbuf.ErrTooLarge)
		}
	}()

	buf := gbuf.NewBuffer(make([]byte, 1))
	const maxInt = int(^uint(0) >> 1)
	buf.Grow(maxInt)
}

// Was a bug: used to give EOF reading empty slice at EOF.
func TestReadEmptyAtEOF(t *testing.T) {
	b := new(gbuf.Buffer[byte])
	slice := make([]byte, 0)
	n, err := b.Read(slice)
	if err != nil {
		t.Errorf("read error: %v", err)
	}
	if n != 0 {
		t.Errorf("wrong count; got %d want 0", n)
	}
}

func TestUnreadByte(t *testing.T) {
	b := new(gbuf.Buffer[byte])

	// check at EOF
	if err := b.UnreadItem(); err == nil {
		t.Fatal("UnreadItem at EOF: got no error")
	}
	if _, err := b.ReadItem(); err == nil {
		t.Fatal("ReadItem at EOF: got no error")
	}
	if err := b.UnreadItem(); err == nil {
		t.Fatal("UnreadItem after ReadItem at EOF: got no error")
	}

	// check not at EOF
	_, _ = b.Write([]byte("abcdefghijklmnopqrstuvwxyz"))

	// after unsuccessful read
	if n, err := b.Read(nil); n != 0 || err != nil {
		t.Fatalf("Read(nil) = %d,%v; want 0,nil", n, err)
	}
	if err := b.UnreadItem(); err == nil {
		t.Fatal("UnreadItem after Read(nil): got no error")
	}

	// after successful read
	if _, err := b.ReadItems(func(b byte) bool {
		return b == 'm'
	}); err != nil {
		t.Fatalf("ReadItems: %v", err)
	}
	if err := b.UnreadItem(); err != nil {
		t.Fatalf("UnreadItem: %v", err)
	}
	c, err := b.ReadItem()
	if err != nil {
		t.Fatalf("ReadItem: %v", err)
	}
	if c != 'm' {
		t.Errorf("ReadItem = %q; want %q", c, 'm')
	}
}

// Tests that we occasionally compact. Issue 5154.
func TestBufferGrowth(t *testing.T) {
	var b gbuf.Buffer[byte]
	buf := make([]byte, 1024)
	_, _ = b.Write(buf[0:1])
	var cap0 int
	for i := 0; i < 5<<10; i++ {
		_, _ = b.Write(buf)
		_, _ = b.Read(buf)
		if i == 0 {
			cap0 = b.Cap()
		}
	}
	cap1 := b.Cap()
	// (*Buffer).grow allows for 2x capacity slop before sliding,
	// so set our error threshold at 3x.
	if cap1 > cap0*3 {
		t.Errorf("buffer cap = %d; too big (grew from %d)", cap1, cap0)
	}
}

func BenchmarkWriteByte(b *testing.B) {
	const n = 4 << 10
	b.SetBytes(n)
	buf := gbuf.NewBuffer(make([]byte, n))
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for i := 0; i < n; i++ {
			_ = buf.WriteItem('x')
		}
	}
}

// From Issue 5154.
func BenchmarkBufferNotEmptyWriteRead(b *testing.B) {
	buf := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		var bb gbuf.Buffer[byte]
		_, _ = bb.Write(buf[0:1])
		for i := 0; i < 5<<10; i++ {
			_, _ = bb.Write(buf)
			_, _ = bb.Read(buf)
		}
	}
}

// Check that we don't compact too often. From Issue 5154.
func BenchmarkBufferFullSmallReads(b *testing.B) {
	buf := make([]byte, 1024)
	for i := 0; i < b.N; i++ {
		var bb gbuf.Buffer[byte]
		_, _ = bb.Write(buf)
		for bb.Len()+20 < bb.Cap() {
			_, _ = bb.Write(buf[:10])
		}
		for idx := 0; idx < 5<<10; idx++ {
			_, _ = bb.Read(buf[:1])
			_, _ = bb.Write(buf[:1])
		}
	}
}

func BenchmarkBufferWriteBlock(b *testing.B) {
	block := make([]byte, 1024)
	for _, n := range []int{1 << 12, 1 << 16, 1 << 20} {
		b.Run(fmt.Sprintf("N%d", n), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				var bb gbuf.Buffer[byte]
				for bb.Len() < n {
					_, _ = bb.Write(block)
				}
			}
		})
	}
}
