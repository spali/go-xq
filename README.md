# xq

[![Build Status](https://github.com/spali/go-xq/actions/workflows/build.yml/badge.svg?branch=main)](https://github.com/spali/go-xq/actions/workflows/build.yml?query=branch%3Amain)
[![Coverage Status](https://codecov.io/gh/spali/go-xq/branch/main/graph/badge.svg)](https://codecov.io/gh/spali/go-xq)
[![CodeQL](https://github.com/spali/go-xq/workflows/CodeQL/badge.svg)](https://github.com/spali/go-xq/actions?query=workflow%3A%22CodeQL%22)
[![govulncheck](https://github.com/spali/go-xq/actions/workflows/govulncheck.yml/badge.svg?branch=main)](https://github.com/spali/go-xq/actions/workflows/govulncheck.yml?query=branch%3Amain)
[![GoDoc](https://godoc.org/github.com/spali/go-xq?status.svg)](https://pkg.go.dev/github.com/spali/go-xq)
[![GitHub license](https://img.shields.io/github/license/spali/go-xq)](https://github.com/spali/go-xq/blob/main/LICENSE)
[![Release](https://img.shields.io/github/v/release/spali/go-xq.svg)](https://github.com/spali/go-xq/releases/latest)

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
- Text, CDATA and attribute nodes are printed as their plain value (not XML-escaped).
- Each match is printed on its own line, in the order returned by the query.
- No match prints nothing and still exits with `0`.
- Expressions returning a number, string or boolean print that value on one line.
  Numbers are formatted like the XPath `string()` function: no exponent,
  `NaN`, `Infinity` and `-Infinity` for special values.

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
| `//book[1]/title/text()`               | `XML Developer's Guide`                        |

### Numbers, strings and booleans

Functions and comparisons print their value, using `catalog.xml` from above:

| Query                         | Output               |
| ----------------------------- | -------------------- |
| `count(//book)`               | `2`                  |
| `sum(//price)`                | `50.900000000000006` |
| `round(sum(//price))`         | `51`                 |
| `string(//book[2]/@id)`       | `bk102`              |
| `concat(//book[1]/@id, "-x")` | `bk101-x`            |
| `//price > 40`                | `true`               |
| `boolean(//book[@lang="fr"])` | `false`              |

Numbers keep all digits needed to represent the floating point result exactly,
like XPath's `string()` does, so `sum(//price)` above is not `50.9`. Use
`round()`, `floor()` or `ceiling()` for whole numbers, or
`round(sum(//price) * 100) div 100` for two decimals.

### Text with special characters

Text and attribute values are printed as plain text, elements stay valid XML:

```sh
echo '<a note="Tom &amp; Jerry"><b>Rain &lt;2&gt; &amp; Sun</b></a>' > special.xml
```

| Query          | Output                            |
| -------------- | --------------------------------- |
| `/a/b/text()`  | `Rain <2> & Sun`                  |
| `/a/@note`     | `Tom & Jerry`                     |
| `/a/b`         | `<b>Rain &lt;2&gt; &amp; Sun</b>` |

### Use it in a script

```sh
version=$(xq -f pom.xml -x '/*[local-name()="project"]/*[local-name()="version"]/text()')
echo "building version $version"

if [ "$(xq -f report.xml -x 'count(//error)')" -gt 0 ]; then
  echo "report contains errors" >&2
  exit 1
fi

if [ "$(xq -f config.xml -x 'boolean(//feature[@name="beta"])')" = true ]; then
  echo "beta enabled"
fi
```

## Limitations

- Namespace prefixes in the query match the prefixes used in the document
  (`//a:x` finds `<a:x>`). For a default namespace (`xmlns="…"`), plain names
  like `//x` work, or use `local-name()` to be independent of prefixes.
- Special characters in element output are XML-escaped (`'` becomes `&#39;`),
  so the output stays valid XML. Select `text()` to get the plain text.
