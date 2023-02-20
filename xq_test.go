package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rhysd/go-fakeio"
)

func Test_getReader(t *testing.T) {
	type args struct {
		file string
	}
	wd, _ := os.Getwd()
	path := filepath.ToSlash(wd)
	handler := http.FileServer(http.Dir("./"))
	server := httptest.NewServer(handler)
	defer server.Close()

	testFileContent := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" +
		"<note>\n" +
		"  <to>Tove</to>\n" +
		"  <from>Jani</from>\n" +
		"  <heading>Reminder</heading>\n" +
		"  <body>Don't forget me this weekend!</body>\n" +
		"</note>"

	tests := []struct {
		name      string
		args      args
		wantInput string
		wantErr   error
	}{
		// valid file paths
		{"valid relative file path", args{"./note.xml"}, testFileContent, nil},
		{"valid absolute file path", args{path + "/note.xml"}, testFileContent, nil},
		{"valid relative file url", args{"file://./note.xml"}, testFileContent, nil},
		{"valid absolute file url", args{"file://" + path + "/note.xml"}, testFileContent, nil},
		{"valid http url", args{server.URL + "/note.xml"}, testFileContent, nil},
		// non existing paths
		{"non existing http url", args{server.URL + "/doesnotexist.xml"}, "", ErrInvalidFile},
		{"invalid http url", args{"http://127.0.0.1:99999"}, "", ErrInvalidFile},
		{"non existing file url", args{"file://./doesnotexist.xml"}, "", ErrInvalidFile},
		{"non existing absolute file url", args{"file://doesnotexist.xml"}, "", ErrInvalidFile},
		{"invalid ftp url", args{"ftp://"}, "", ErrInvalidFile},
		{"without argument", args{""}, "", ErrNoInput},
		// stdin
		{"without stdin argument", args{"-"}, testFileContent, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if tt.args.file == "-" {
				// fake stdin from file
				if f, err := os.Open("./note.xml"); err != nil {
					panic(err)
				} else {
					if b, err := io.ReadAll(f); err != nil {
						panic(err)
					} else {
						fake := fakeio.StdinBytes(b).CloseStdin()
						defer fake.Restore()
					}
				}
			}

			var gotInput string
			reader, err := getReader(tt.args.file)

			if err == nil {
				bytes, err := io.ReadAll(reader)
				if err != nil {
					panic(err)
				}
				gotInput = string(bytes)
			}
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("getReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotInput != tt.wantInput {
				t.Errorf("getReader() gotInput = %v, want %v", gotInput, tt.wantInput)
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
		wantErr   error
	}{
		{"xpath and explicit stdin argument",
			args{xpath: "xxx", f: "-", input: "./note.xml"},
			"-", "xxx", nil},
		{"no arguments and no input",
			args{},
			"-", "", ErrMissingXpath},
		{"only x argument",
			args{x: "abc"},
			"-", "abc", nil},
		{"only xpath argument",
			args{xpath: "abc"},
			"-", "abc", nil},
		{"no arguments but stdin input",
			args{input: "./note.xml"},
			"-", "", ErrMissingXpath},
		{"only f argument",
			args{f: "./note.xml"},
			"./note.xml", "", ErrMissingXpath},
		{"only file argument",
			args{file: "./note.xml"},
			"./note.xml", "", ErrMissingXpath},
		{"f,x argument",
			args{x: "xxx", f: "./note.xml"},
			"./note.xml", "xxx", nil},
		{"f,xpath argument",
			args{xpath: "xxx", f: "./note.xml"},
			"./note.xml", "xxx", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotXpath, err := parseArguments(tt.args.file, tt.args.f, tt.args.xpath, tt.args.x)
			if gotFile != tt.wantFile {
				t.Errorf("parseArguments() gotFile = %v, want %v", gotFile, tt.wantFile)
			}
			if gotXpath != tt.wantXpath {
				t.Errorf("parseArguments() gotXpath = %v, want %v", gotXpath, tt.wantXpath)
			}
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("parseArguments() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_query(t *testing.T) {
	type args struct {
		reader io.Reader
		xpath  string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    error
	}{
		{"invalid xml syntax (unexpected EOF)", args{strings.NewReader("<"), "/"}, "", ErrXMLParse},
		{"invalid xml syntax (expected element name after)", args{strings.NewReader("<>"), "/"}, "", ErrXMLParse},
		{"empty expression", args{strings.NewReader(""), ""}, "", ErrXMLQuery},
		{"invalid expression", args{strings.NewReader("<abc></abc>"), "-*/"}, "", ErrXMLQuery},
		{"test root element specifically queried by name", args{strings.NewReader("<abc></abc>"), "/abc"}, "<abc></abc>\n", nil},
		{"test root element specifically queried", args{strings.NewReader("<abc></abc>"), "/"}, "<?xml version=\"1.0\"?><abc></abc>\n", nil},
		{"test root element queried", args{strings.NewReader("<abc></abc>"), "/*"}, "<abc></abc>\n", nil},
		{"test attribute node", args{strings.NewReader("<abc id=\"test\"></abc>"), "/abc/@id"}, "test\n", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := fakeio.Stdout()
			defer fake.Restore()

			var gotOutput string
			err := query(tt.args.reader, tt.args.xpath)
			if err == nil {
				gotOutput, err = fake.String()
				if err != nil {
					panic(err)
				}
			}

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

func Test_xq(t *testing.T) {
	type args struct {
		file  string
		xpath string
		stdin string
	}
	tests := []struct {
		name       string
		args       args
		wantOutput string
		wantErr    error
	}{
		{"valid file path", args{"./note.xml", "/note/to/text()", ""}, "Tove\n", nil},
		{"invalid file path", args{"./doesnotexist.xml", "/note/to/text()", ""}, "", ErrInvalidFile},
		{"invalid query", args{"./note.xml", "", ""}, "", ErrXMLQuery},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fake := fakeio.Stdout()
			defer fake.Restore()

			var gotOutput string
			err := xq(tt.args.file, tt.args.xpath)
			if err == nil {
				gotOutput, err = fake.String()
				if err != nil {
					panic(err)
				}
			}

			if gotOutput != tt.wantOutput {
				t.Errorf("xq() gotOutput = %v, want %v", gotOutput, tt.wantOutput)
			}
			if (err != nil || tt.wantErr != nil) && !errors.Is(err, tt.wantErr) {
				t.Errorf("query() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
