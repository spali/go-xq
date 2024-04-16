//go:build !test

package main

import (
	"flag"
	"fmt"
	"os"
)

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

	file, xpath, err := parseArguments(*_file, *_f, *_xpath, *_x)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		printUsage()
		os.Exit(1)
	}

	if err := xq(file, xpath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}

}
