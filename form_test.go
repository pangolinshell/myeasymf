package form

import (
	"errors"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"strings"
	"testing"
)

// --- Mock helpers ---

func newMultipartRequest() (*http.Request, *multipart.Writer) {
	body := &strings.Builder{}
	writer := multipart.NewWriter(body)
	req := &http.Request{
		Method: "POST",
		Header: make(http.Header),
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = nopCloser{strings.NewReader(body.String())}
	return req, writer
}

type nopCloser struct {
	*strings.Reader
}

func (nopCloser) Close() error { return nil }

// Mock struct for decoding
type TestForm struct {
	Name  string                  `form:"name"`
	Age   int                     `form:"age"`
	File  *multipart.FileHeader   `form:"file"`
	Files []*multipart.FileHeader `form:"files"`
}

// --- Tests ---

func TestDecode_SimpleValues(t *testing.T) {
	req := &http.Request{
		MultipartForm: &multipart.Form{
			Value: map[string][]string{
				"name": {"Alice"},
				"age":  {"30"},
			},
		},
	}
	d := NewDecoder(req)

	var dst TestForm
	if err := d.Decode(&dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if dst.Name != "Alice" {
		t.Errorf("expected Name=Alice, got %q", dst.Name)
	}
	if dst.Age != 30 {
		t.Errorf("expected Age=30, got %d", dst.Age)
	}
}

func TestDecode_FileSlice(t *testing.T) {
	fh := &multipart.FileHeader{
		Filename: "test.txt",
		Header:   make(textproto.MIMEHeader),
	}
	req := &http.Request{
		MultipartForm: &multipart.Form{
			File: map[string][]*multipart.FileHeader{
				"files": {fh},
			},
		},
	}
	d := NewDecoder(req)
	var dst TestForm
	if err := d.Decode(&dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(dst.Files) != 1 || dst.Files[0].Filename != "test.txt" {
		t.Errorf("expected 1 file named test.txt, got %+v", dst.Files)
	}
}

func TestDecode_SingleFile(t *testing.T) {
	fh := &multipart.FileHeader{
		Filename: "photo.png",
		Header:   make(textproto.MIMEHeader),
	}
	req := &http.Request{
		MultipartForm: &multipart.Form{
			File: map[string][]*multipart.FileHeader{
				"file": {fh},
			},
		},
	}
	d := NewDecoder(req)
	var dst TestForm
	if err := d.Decode(&dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.File == nil || dst.File.Filename != "photo.png" {
		t.Errorf("expected File=photo.png, got %+v", dst.File)
	}
}

func TestDecode_SingleFile_MultipleFilesError(t *testing.T) {
	fh1 := &multipart.FileHeader{Filename: "a.txt"}
	fh2 := &multipart.FileHeader{Filename: "b.txt"}
	req := &http.Request{
		MultipartForm: &multipart.Form{
			File: map[string][]*multipart.FileHeader{
				"file": {fh1, fh2}, // too many
			},
		},
	}
	d := NewDecoder(req)
	var dst TestForm
	err := d.Decode(&dst)
	if !errors.Is(err, ErrMultipleFilesForSingleField) {
		t.Fatalf("expected ErrMultipleFilesForSingleField, got %v", err)
	}
}

func TestDecode_InvalidDestination(t *testing.T) {
	d := NewDecoder(&http.Request{})
	var notStruct int
	err := d.Decode(notStruct)
	if !errors.Is(err, ErrNotPtrToStruct) {
		t.Fatalf("expected ErrNotPtrToStruct, got %v", err)
	}
}

func TestDecode_UnsupportedFileType(t *testing.T) {
	type BadStruct struct {
		BadField string `form:"file"`
	}
	req := &http.Request{
		MultipartForm: &multipart.Form{
			File: map[string][]*multipart.FileHeader{
				"file": {{Filename: "data.bin"}},
			},
		},
	}
	d := NewDecoder(req)
	var dst BadStruct
	err := d.Decode(&dst)
	if !errors.Is(err, ErrUnsupportedFileFieldType) {
		t.Fatalf("expected ErrUnsupportedFileFieldType, got %v", err)
	}
}

// --- Sanity check for reflection safety ---
func TestDecode_ShouldNotPanicOnUnsettable(t *testing.T) {
	typ := reflect.TypeOf(struct{ x string }{})
	v := reflect.New(typ)
	req := &http.Request{MultipartForm: &multipart.Form{Value: map[string][]string{"x": {"y"}}}}
	d := NewDecoder(req)
	_ = d.Decode(v.Interface()) // Should not panic
}
