package main

import (
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func Test_getDoc(t *testing.T) {
	type args struct {
		file string
	}
	path, _ := os.Getwd()
	handler := http.FileServer(http.Dir("./"))
	server := httptest.NewServer(handler)
	defer server.Close()
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{"Test without argument", args{""}, ErrNoInput},
		{"Test valid http url", args{server.URL}, nil},
		{"Test valid relative file url", args{"file://./note.xml"}, nil},
		{"Test valid relative non existing file url", args{"file://./doesnotexist.xml"}, os.ErrNotExist},
		{"Test valid absolute file url", args{"file://" + path + "/note.xml"}, nil},
		{"Test non existing absolute file url", args{"file://doesnotexist.xml"}, os.ErrNotExist},
		{"Test invalid ftp url", args{"ftp://"}, ErrUnsupportedURL},
		{"Test valid relative file path", args{"./note.xml"}, nil},
		{"Test valid absolute file path", args{path + "/note.xml"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getReader(tt.args.file)
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("getReader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if got == nil && tt.wantErr == nil {
				t.Errorf("getReader() = %v, want %v", got, "!= <nil>")
			}
		})
	}

}

func Test_parseArguments(t *testing.T) {
	type args struct {
		x     string
		xpath string
		f     string
		file  string
		input string
	}
	tests := []struct {
		name string
		args
		wantFile  string
		wantXpath string
		wantStdin bool
		wantErr   error
	}{
		{"no arguments and no input",
			args{},
			"", "", false, ErrMissingFile},
		{"only x argument",
			args{x: "abc"},
			"", "", false, ErrMissingFile},
		{"only xpath argument",
			args{xpath: "abc"},
			"", "", false, ErrMissingFile},
		{"no arguments but stdin input",
			args{input: "./note.xml"},
			"", "", false, ErrMissingXpath},
		{"only f argument",
			args{f: "./note.xml"},
			"./note.xml", "", false, ErrMissingXpath},
		{"only file argument",
			args{file: "./note.xml"},
			"./note.xml", "", false, ErrMissingXpath},
		{"f,x argument",
			args{x: "xxx", f: "./note.xml"},
			"./note.xml", "xxx", false, nil},
		{"f,xpath argument",
			args{xpath: "xxx", f: "./note.xml"},
			"./note.xml", "xxx", false, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var stdin *os.File
			if tt.args.input == "" {
				stdin = os.NewFile(0, "stdin")
			} else {
				stdin, _ = os.Open(tt.args.input)
			}
			gotFile, gotXpath, gotStdin, err := parseArguments(tt.args.file, tt.args.f, tt.args.xpath, tt.args.x, *stdin)
			if gotFile != tt.wantFile {
				t.Errorf("parseArguments() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotXpath != tt.wantXpath {
				t.Errorf("parseArguments() gotXpath = %v, want %v", gotXpath, tt.wantXpath)
			}
			if gotXpath != tt.wantXpath {
				t.Errorf("parseArguments() gotStdin = %v, want %v", gotStdin, tt.wantStdin)
			}
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("parseArguments() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

type stringReader struct {
	io.ReadCloser
	sr *strings.Reader
}

func (s *stringReader) Close() error {
	return nil
}

func (s *stringReader) Read(b []byte) (int, error) {
	return s.sr.Read(b)
}

func newStringReader(s string) *stringReader {
	return &stringReader{sr: strings.NewReader(s)}
}

func Test_query(t *testing.T) {
	type args struct {
		reader io.ReadCloser
		xpath  string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    error
	}{
		{"invalid xml syntax (unexpected EOF)", args{newStringReader("<"), "/"}, "", ErrXMLParse},
		{"invalid xml syntax (expected element name after)", args{newStringReader("<>"), "/"}, "", ErrXMLParse},
		{"empty expression", args{newStringReader(""), ""}, "", ErrXMLQuery},
		{"invalid expression", args{newStringReader("<abc></abc>"), "-/"}, "", ErrXMLQuery},
		{"test root element specifically queried by name", args{newStringReader("<abc></abc>"), "/abc"}, "<abc></abc>", nil},
		{"test root element specifically queried", args{newStringReader("<abc></abc>"), "/"}, "<?xml?><abc></abc>", nil},
		{"test root element queried", args{newStringReader("<abc></abc>"), "/*"}, "<abc></abc>", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			rescueStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := query(tt.args.reader, tt.args.xpath)

			w.Close()
			out, _ := ioutil.ReadAll(r)
			os.Stdout = rescueStdout

			gotOutput := strings.TrimSuffix(string(out), "\n")
			if gotOutput != tt.wantOutput {
				t.Errorf("query() gotOutput = %v, want %v", gotOutput, tt.wantOutput)
			}
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}

}
