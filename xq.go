package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/antchfx/xmlquery"
)

// set during built time
var (
	name      string = "xq"
	source    string = "unknown"
	version   string = "unknown"
	commit    string = "unknown"
	platform  string = "unknown"
	buildTime string = "unknown"
)

var (
	ErrUnsupportedURL = errors.New("unsupported url")
	ErrNoInput        = errors.New("no input")
	ErrMissingFile    = errors.New("missing file argument")
	ErrMissingXpath   = errors.New("missing xpath expression argument")
	ErrXMLParse       = errors.New("xmlparse error")
	ErrXMLQuery       = errors.New("xmlquery error")
)

func getReader(file string) (io.ReadCloser, error) {
	url, err := url.ParseRequestURI(file)
	if err == nil {
		if url.Scheme == "file" {
			filePath := fmt.Sprintf("%s%s", url.Host, url.Path)
			if _, err := os.Stat(filePath); err != nil {
				return nil, err
			}
			f, err := os.Open(filePath)
			if err != nil {
				return nil, err
			}
			return f, nil
		}
		if url.Scheme == "http" || url.Scheme == "https" {
			// nolint:gosec,bodyclose,noctx
			resp, err := http.Get(file)
			if err != nil {
				return nil, err
			}
			return resp.Body, nil
		}
		if url.Scheme != "" {
			return nil, fmt.Errorf("%w: %s", ErrUnsupportedURL, file)
		}
	}
	if file != "" {
		f, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	return nil, ErrNoInput
}

func printVersion() {
	fmt.Fprintln(os.Stderr, name)
	fmt.Fprintf(os.Stderr, "Source: %s\n", source)
	fmt.Fprintf(os.Stderr, "Version: %s\n", version)
	fmt.Fprintf(os.Stderr, "Commit: %s\n", commit)
	fmt.Fprintf(os.Stderr, "Platform: %s\n", platform)
	fmt.Fprintf(os.Stderr, "Build Time: %s\n", buildTime)
}

func printUsage() {
	printVersion()
	fmt.Fprintln(os.Stderr, "usage:")
	flag.PrintDefaults()
}

func parseArguments(_file string, _f string, _xpath string, _x string, s os.File) (file string, xpath string, stdin bool, err error) {
	stat, _ := s.Stat()
	if stdin = (stat.Mode() & os.ModeCharDevice) == 0; !stdin {
		if _f != "" {
			file = _f
		}
		if _file != "" {
			file = _file
		}
		if file == "" {
			err = ErrMissingFile
			return
		}
		if file == "-" {
			stdin = true
		}
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

func query(reader io.ReadCloser, xpath string) error {
	defer reader.Close()
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
			fmt.Printf("%s\n", elem.InnerText())
		} else {
			// prevent output of pseudo tags if root element was selected
			isRoot := elem.Parent != nil
			fmt.Printf("%s\n", elem.OutputXML(isRoot))
		}
	}
	return nil
}

func main() {
	_file := flag.String("file", "", "file or URL to the xml")
	_f := flag.String("f", "", "file or URL to the xml")
	_xpath := flag.String("xpath", "", "the xpath expression")
	_x := flag.String("x", "", "the xpath expression")
	_version := flag.Bool("version", false, "print the version")
	_v := flag.Bool("v", false, "print the version")
	_help := flag.Bool("help", false, "print usage")
	_h := flag.Bool("h", false, "print usage")
	flag.Parse()

	if *_h || *_help {
		printUsage()
		os.Exit(0)
	}

	if *_v || *_version {
		printVersion()
		os.Exit(0)
	}

	file, xpath, stdin, err := parseArguments(*_file, *_f, *_xpath, *_x, *os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		printUsage()
		os.Exit(1)
	}

	var reader io.ReadCloser
	if stdin {
		reader = os.Stdin
	} else {
		var err error
		if reader, err = getReader(file); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			os.Exit(1)
		}
	}
	if err := query(reader, xpath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
