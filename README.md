# xq

[![Build Status](https://github.com/spali/go-xq/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/spali/go-xq/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://codecov.io/gh/spali/go-xq/branch/main/graph/badge.svg)](https://codecov.io/gh/spali/go-xq)
[![CodeQL](https://github.com/spali/go-xq/workflows/CodeQL/badge.svg)](https://github.com/spali/go-xq/actions?query=workflow%3A%22CodeQL%22)
[![Go Report Card](https://goreportcard.com/badge/github.com/spali/go-xq)](https://goreportcard.com/report/github.com/spali/go-xq)
[![GoDoc](https://godoc.org/github.com/spali/go-xq?status.svg)](https://pkg.go.dev/github.com/spali/go-xq)
[![GitHub license](https://img.shields.io/github/license/spali/go-xq)](https://github.com/spali/go-xq/blob/main/LICENSE)

XML query utility to extract XML from a file, an input pipe or a URL with XPath.
Backed by the excellent library [antchfx/xmlquery](https://github.com/antchfx/xmlquery).

`xq` reads one XML document, evaluates an XPath expression against it and prints
every matching node on its own line. It is a single static binary without
dependencies, handy in shell scripts and CI pipelines.

## Installation

Download the archive for your platform from the
[latest release](https://github.com/spali/go-xq/releases/latest)
(Linux, macOS and Windows; amd64, arm64, arm and 386), extract it and put `xq`
somewhere in your `PATH`.

Or build it with Go:

```sh
go install github.com/spali/go-xq@latest
```

Note: `go install` names the binary after the module, so it is installed as `go-xq`.

## Usage

```text
xq [-f <file|url|->] -x <xpath>
```

| Flag              | Description                                                                      |
| ----------------- | -------------------------------------------------------------------------------- |
| `-x`, `-xpath`    | XPath expression to evaluate (required)                                           |
| `-f`, `-file`     | Input: a file path, a `file://` URL, an `http(s)://` URL or `-` for stdin (default) |
| `-v`, `-version`  | Print version information                                                         |
| `-h`, `-help`     | Print usage                                                                       |

Output rules:

- Element nodes are printed as XML, including their children.
- `text()` nodes are printed as (XML-escaped) text, attribute nodes as their plain value.
- Each match is printed on its own line, in the order returned by the query.
- No match prints nothing and still exits with `0`.

Exit codes: `0` on success (also when nothing matched), `1` on a missing argument,
an unreadable input, invalid XML or an invalid XPath expression. Errors are written to stderr.

## Examples

The examples use [`note.xml`](note.xml) from this repository:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<note>
  <to>Tove</to>
  <from>Jani</from>
  <heading>Reminder</heading>
  <body>Don't forget me this weekend!</body>
</note>
```

### Select an element

```sh
xq -f note.xml -x '/note/to'
```

```xml
<to>Tove</to>
```

### Get only the text

```sh
xq -f note.xml -x '/note/to/text()'
```

```text
Tove
```

### Several nodes at once

```sh
xq -f note.xml -x '/note/*[self::to or self::from]/text()'
```

```text
Tove
Jani
```

### Read from a pipe

Without `-f` (or with `-f -`) the document is read from stdin:

```sh
curl -s https://example.com/feed.xml | xq -x '//item/title/text()'
echo '<a><b>1</b><b>2</b></a>' | xq -x '//b'
```

```xml
<b>1</b>
<b>2</b>
```

### Read from a URL

`http://` and `https://` URLs are fetched directly; any status other than `200 OK` is an error:

```sh
xq -f https://example.com/feed.xml -x '//item[1]/link/text()'
```

### Attributes and predicates

Given `catalog.xml`:

```xml
<catalog>
  <book id="bk101" lang="en">
    <title>XML Developer's Guide</title>
    <price>44.95</price>
  </book>
  <book id="bk102" lang="de">
    <title>Midnight Rain</title>
    <price>5.95</price>
  </book>
</catalog>
```

| Query                                  | Output                                         |
| -------------------------------------- | ---------------------------------------------- |
| `//book/@id`                           | `bk101`<br>`bk102`                             |
| `//book[@lang="de"]/title/text()`      | `Midnight Rain`                                |
| `//book[price > 10]/title`             | `<title>XML Developer&#39;s Guide</title>`     |
| `//book[contains(title, "Rain")]/@id`  | `bk102`                                        |
| `//book[last()]/price/text()`          | `5.95`                                         |

### Use it in a script

```sh
version=$(xq -f pom.xml -x '/*[local-name()="project"]/*[local-name()="version"]/text()')
echo "building version $version"
```

## Limitations

- The expression must select nodes. Expressions returning a number, string or
  boolean, like `count(//book)`, print nothing.
- Namespace prefixes in the query match the prefixes used in the document
  (`//a:x` finds `<a:x>`). For a default namespace (`xmlns="…"`), plain names
  like `//x` work, or use `local-name()` to be independent of prefixes.
- Special characters in element and `text()` output are XML-escaped
  (`'` becomes `&#39;`). Attribute values are printed unescaped.
