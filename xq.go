package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/antchfx/xmlquery"
)

// set during built time
var (
	name      = "xq"
	source    = "unknown"
	version   = "unknown"
	commit    = "unknown"
	platform  = "unknown"
	buildTime = "unknown"
)

var (
	ErrInvalidFile  = errors.New("invalid file")
	ErrNoInput      = errors.New("no input")
	ErrMissingXpath = errors.New("missing xpath expression argument")
	ErrXMLParse     = errors.New("xmlparse error")
	ErrXMLQuery     = errors.New("xmlquery error")
)

func getReader(file string) (io.Reader, error) {
	if file == "-" {
		return os.Stdin, nil
	}
	if u, err := url.ParseRequestURI(file); err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		// nolint:gosec,bodyclose,noctx
		resp, err := http.Get(file)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFile, err)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFile, resp.Status)
		}
		return resp.Body, nil
	} else if err == nil && u.Scheme == "file" {
		file = filepath.FromSlash(fmt.Sprintf("%s%s", u.Host, u.Path))
	}
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("%w: %s", ErrInvalidFile, err)
		}
		return f, nil
	}
	return nil, ErrNoInput
}

func parseArguments(_file string, _f string, _xpath string, _x string) (file string, xpath string, err error) {
	if _f != "" {
		file = _f
	}
	if _file != "" {
		file = _file
	}
	if file == "-" || file == "" {
		file = "-"
	}
	if _x != "" {
		xpath = _x
	}
	if _xpath != "" {
		xpath = _xpath
	}
	if xpath == "" {
		err = ErrMissingXpath
		return
	}
	return
}

func query(reader io.Reader, xpath string) error {
	if readCloser, ok := reader.(io.ReadCloser); ok {
		defer readCloser.Close()
	}
	doc, err := xmlquery.Parse(reader)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrXMLParse, err)
	}
	list, err := xmlquery.QueryAll(doc, xpath)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrXMLQuery, err)
	}
	for _, elem := range list {
		if elem.Type == xmlquery.AttributeNode {
			// do not nest attribute node in a tag
			fmt.Fprintf(os.Stdout, "%s\n", elem.InnerText())
		} else {
			// prevent output of pseudo tags if root element was selected
			isRoot := elem.Parent != nil
			fmt.Fprintf(os.Stdout, "%s\n", elem.OutputXML(isRoot))
		}
	}
	return nil
}

func xq(file string, xpath string) error {
	if reader, err := getReader(file); err != nil {
		return err
	} else if err = query(reader, xpath); err != nil {
		return err
	}
	return nil
}
