package main

import (
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"

	"github.com/antchfx/xmlquery"
	"github.com/antchfx/xpath"
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
		//nolint:gosec,bodyclose,noctx
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

func query(reader io.Reader, expression string) error {
	if readCloser, ok := reader.(io.ReadCloser); ok {
		defer readCloser.Close()
	}
	doc, err := xmlquery.Parse(reader)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrXMLParse, err)
	}
	expr, err := xpath.Compile(expression)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrXMLQuery, err)
	}
	switch result := expr.Evaluate(xmlquery.CreateXPathNavigator(doc)).(type) {
	case *xpath.NodeIterator:
		// let xmlquery map the result to nodes (it builds the attribute nodes)
		for _, node := range xmlquery.QuerySelectorAll(doc, expr) {
			printNode(node)
		}
	case float64:
		fmt.Fprintln(os.Stdout, formatNumber(result))
	default: // string or bool
		fmt.Fprintln(os.Stdout, result)
	}
	return nil
}

func printNode(node *xmlquery.Node) {
	if node.Type == xmlquery.AttributeNode || node.Type == xmlquery.TextNode || node.Type == xmlquery.CharDataNode {
		// print the value itself, not nested in a tag and not XML-escaped
		fmt.Fprintln(os.Stdout, node.InnerText())
		return
	}
	// prevent output of pseudo tags if root element was selected
	isRoot := node.Parent != nil
	fmt.Fprintln(os.Stdout, node.OutputXML(isRoot))
}

// formatNumber converts a number to a string like the XPath 1.0 string()
// function (REC 4.2): no exponent, "NaN", "Infinity", "-Infinity" and "0" for -0.
func formatNumber(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0:
		return "0"
	default:
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
}

func xq(file string, xpath string) error {
	if reader, err := getReader(file); err != nil {
		return err
	} else if err = query(reader, xpath); err != nil {
		return err
	}
	return nil
}
