package gio_test

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/zalgonoise/x/gio"
)

func ExampleCopy() {
	r := strings.NewReader("some io.Reader stream to be read\n")

	if _, err := gio.Copy[byte](os.Stdout, r); err != nil {
		log.Fatal(err)
	}

	// Output:
	// some io.Reader stream to be read
}

func ExampleCopyBuffer() {
	r1 := strings.NewReader("first reader\n")
	r2 := strings.NewReader("second reader\n")
	buf := make([]byte, 8)

	// buf is used here...
	if _, err := gio.CopyBuffer[byte](os.Stdout, r1, buf); err != nil {
		log.Fatal(err)
	}

	// ... reused here also. No need to allocate an extra buffer.
	if _, err := gio.CopyBuffer[byte](os.Stdout, r2, buf); err != nil {
		log.Fatal(err)
	}

	// Output:
	// first reader
	// second reader
}

func ExampleCopyN() {
	r := strings.NewReader("some io.Reader stream to be read")

	if _, err := gio.CopyN[byte](os.Stdout, r, 4); err != nil {
		log.Fatal(err)
	}

	// Output:
	// some
}

func ExampleReadAtLeast() {
	r := strings.NewReader("some io.Reader stream to be read\n")

	buf := make([]byte, 14)
	if _, err := gio.ReadAtLeast[byte](r, buf, 4); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf)

	// buffer smaller than minimal read size.
	shortBuf := make([]byte, 3)
	if _, err := gio.ReadAtLeast[byte](r, shortBuf, 4); err != nil {
		fmt.Println("error:", err)
	}

	// minimal read size bigger than io.Reader stream
	longBuf := make([]byte, 64)
	if _, err := gio.ReadAtLeast[byte](r, longBuf, 64); err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// some io.Reader
	// error: short buffer
	// error: unexpected EOF
}

func ExampleReadFull() {
	r := strings.NewReader("some io.Reader stream to be read\n")

	buf := make([]byte, 4)
	if _, err := gio.ReadFull[byte](r, buf); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s\n", buf)

	// minimal read size bigger than io.Reader stream
	longBuf := make([]byte, 64)
	if _, err := gio.ReadFull[byte](r, longBuf); err != nil {
		fmt.Println("error:", err)
	}

	// Output:
	// some
	// error: unexpected EOF
}

func ExampleLimitReader() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	lr := gio.LimitReader[byte](r, 4)

	if _, err := gio.Copy[byte](os.Stdout, lr); err != nil {
		log.Fatal(err)
	}

	// Output:
	// some
}

func ExampleMultiReader() {
	r1 := strings.NewReader("first reader ")
	r2 := strings.NewReader("second reader ")
	r3 := strings.NewReader("third reader\n")
	r := gio.MultiReader[byte](r1, r2, r3)

	if _, err := gio.Copy[byte](os.Stdout, r); err != nil {
		log.Fatal(err)
	}

	// Output:
	// first reader second reader third reader
}

func ExampleTeeReader() {
	var r io.Reader = strings.NewReader("some io.Reader stream to be read\n")

	r = gio.TeeReader[byte](r, os.Stdout)

	// Everything read from r will be copied to stdout.
	if _, err := gio.ReadAll[byte](r); err != nil {
		log.Fatal(err)
	}

	// Output:
	// some io.Reader stream to be read
}

func ExampleSectionReader() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	s := gio.NewSectionReader[byte](r, 5, 17)

	if _, err := gio.Copy[byte](os.Stdout, s); err != nil {
		log.Fatal(err)
	}

	// Output:
	// io.Reader stream
}

func ExampleSectionReader_Read() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	s := gio.NewSectionReader[byte](r, 5, 17)

	buf := make([]byte, 9)
	if _, err := s.Read(buf); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", buf)

	// Output:
	// io.Reader
}

func ExampleSectionReader_ReadAt() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	s := gio.NewSectionReader[byte](r, 5, 17)

	buf := make([]byte, 6)
	if _, err := s.ReadAt(buf, 10); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s\n", buf)

	// Output:
	// stream
}

func ExampleSectionReader_Seek() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	s := gio.NewSectionReader[byte](r, 5, 17)

	if _, err := s.Seek(10, io.SeekStart); err != nil {
		log.Fatal(err)
	}

	if _, err := gio.Copy[byte](os.Stdout, s); err != nil {
		log.Fatal(err)
	}

	// Output:
	// stream
}

func ExampleSectionReader_Size() {
	r := strings.NewReader("some io.Reader stream to be read\n")
	s := gio.NewSectionReader[byte](r, 5, 17)

	fmt.Println(s.Size())

	// Output:
	// 17
}

func ExampleSeeker_Seek() {
	r := strings.NewReader("some io.Reader stream to be read\n")

	r.Seek(5, gio.SeekStart) // move to the 5th char from the start
	if _, err := gio.Copy[byte](os.Stdout, r); err != nil {
		log.Fatal(err)
	}

	r.Seek(-5, gio.SeekEnd)
	if _, err := gio.Copy[byte](os.Stdout, r); err != nil {
		log.Fatal(err)
	}

	// Output:
	// io.Reader stream to be read
	// read
}

func ExampleMultiWriter() {
	r := strings.NewReader("some io.Reader stream to be read\n")

	var buf1, buf2 bytes.Buffer
	w := gio.MultiWriter[byte](&buf1, &buf2)

	if _, err := gio.Copy[byte](w, r); err != nil {
		log.Fatal(err)
	}

	fmt.Print(buf1.String())
	fmt.Print(buf2.String())

	// Output:
	// some io.Reader stream to be read
	// some io.Reader stream to be read
}

func ExamplePipe() {
	r, w := gio.Pipe[byte]()

	go func() {
		fmt.Fprint(w, "some io.Reader stream to be read\n")
		w.Close()
	}()

	if _, err := gio.Copy[byte](os.Stdout, r); err != nil {
		log.Fatal(err)
	}

	// Output:
	// some io.Reader stream to be read
}

func ExampleReadAll() {
	r := strings.NewReader("Go is a general-purpose language designed with systems programming in mind.")

	b, err := gio.ReadAll[byte](r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s", b)

	// Output:
	// Go is a general-purpose language designed with systems programming in mind.
}
