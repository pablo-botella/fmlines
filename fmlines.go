// Package fmlines is a front matter line extractor — not a parser: it
// tokenizes a header line by line and classifies each one; what the lines
// MEAN (the dialect: which keys, which sections) is the consumer's
// business. Every line read becomes a token, the unrecognized ones too —
// raw and numbered, never lost.
//
// It reads through a github.com/pablo-botella/linereader LineReader. The family:
// ReadLine takes ONE line into the list and tells its type, ReadHeader
// takes a fenced header (.md) and stops right after the closing --- — the
// body stays in the reader, keep reading it there — and ReadAll takes
// everything (.fm: everything is header).
package fmlines

// The docs are composed from _mkskill/ — edit the sources there, then:
//go:generate go run github.com/pablo-botella/mkskill/cmd/mkskill@latest build

import (
	"fmt"
	"io"
	"strings"

	"github.com/pablo-botella/linereader"
)

type FmLineType int

const (
	FmLineInvalid    FmLineType = 0
	FmLineStartBlock FmLineType = 0x0001
	FmLineEndBlock   FmLineType = 0x0002
	FmLineKeyValue   FmLineType = 0x0010
	FmLineSection    FmLineType = 0x0020
	FmLineEmpty      FmLineType = 0x0100
)

type FmLine struct {
	Type     FmLineType
	LineNum  int
	RawLine  string
	KvSplit  int
	Left     int
	Parent   *FmLine
	Level    int
	Children *FmLines
}
type FmLines []*FmLine

type FmLineReader struct {
	Lines  *FmLines
	Reader linereader.IoLineReader
}

func (l *FmLine) Value() string {
	if l.Type != FmLineKeyValue {
		return ""
	}
	return l.RawLine[l.KvSplit+1:]
}

func (l *FmLine) Name() string {
	if l.Type != FmLineKeyValue && l.Type != FmLineSection {
		return ""
	}
	return l.RawLine[:l.KvSplit]
}

// isFence tells whether a line is a --- fence.
func isFence(text string) bool {
	return strings.TrimRight(text, " \t") == "---"
}

// parent finds who a line at this indentation hangs from: walking the list
// backwards, the first line further left decides — a section adopts it,
// anything else means no parent. Empty lines and fences don't count.
func (lines FmLines) parent(left int) *FmLine {
	for i := len(lines) - 1; i >= 0; i-- {
		previous := lines[i]
		switch previous.Type {
		case FmLineEmpty, FmLineStartBlock, FmLineEndBlock:
			continue
		}
		if previous.Left < left {
			if previous.Type == FmLineSection {
				return previous
			}
			return nil
		}
	}
	return nil
}

// add tokenizes one line and puts it into the list — the list itself is the
// context: Parent and Level come from what is already in it.
func (lines *FmLines) add(text string) *FmLine {
	line := &FmLine{LineNum: len(*lines) + 1, RawLine: text}
	context := *lines
	*lines = append(*lines, line)

	trimmed := strings.TrimSpace(text)
	line.Left = len(text) - len(strings.TrimLeft(text, " \t"))

	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		line.Type = FmLineEmpty
		return line
	}

	if parent := context.parent(line.Left); parent != nil {
		line.Parent = parent
		line.Level = parent.Level + 1
		if parent.Children == nil {
			parent.Children = &FmLines{}
		}
		*parent.Children = append(*parent.Children, line)
	}

	colon := strings.Index(text[line.Left:], ":")
	if colon < 0 {
		line.Type = FmLineInvalid
		return line
	}
	line.KvSplit = line.Left + colon
	if strings.TrimSpace(text[line.KvSplit+1:]) == "" {
		line.Type = FmLineSection
		return line
	}
	line.Type = FmLineKeyValue
	return line
}

// ReadLine reads ONE line from the reader, puts it into the list and tells
// what it was. Fences are not its business (that is ReadHeader's): a ---
// here is just an invalid line. io.EOF when the reader is done.
func (lines *FmLines) ReadLine(lr linereader.IoLineReader) (FmLineType, error) {
	text, err := lr.ReadLine()
	if err != nil {
		return FmLineInvalid, err
	}
	return lines.add(string(text)).Type, nil
}

// ReadHeader tokenizes a fenced header (.md) into the list and STOPS right after the
// closing ---: the body stays in the reader — keep reading it there, by
// lines, by bytes, or hand the reader to any io.Reader consumer. The first
// line must be the opening fence; when it is not, there is no front matter:
// that single line comes back as the only token (read is read, nothing gets
// lost) and the rest of the document was never touched. An unclosed fence
// is an error.
func (lines *FmLines) ReadHeader(lr linereader.IoLineReader) error {
	first, err := lr.ReadLine()
	if err == io.EOF {
		return nil // empty document: no front matter, no body
	}
	if err != nil {
		return err
	}

	if !isFence(string(first)) {
		// no front matter: the consumed line, kept as the only token
		lines.add(string(first))
		return nil
	}
	*lines = append(*lines, &FmLine{Type: FmLineStartBlock, LineNum: len(*lines) + 1, RawLine: string(first)})

	for {
		text, err := lr.ReadLine()
		if err == io.EOF {
			return fmt.Errorf("unclosed front matter: missing closing ---")
		}
		if err != nil {
			return err
		}
		if isFence(string(text)) {
			*lines = append(*lines, &FmLine{Type: FmLineEndBlock, LineNum: len(*lines) + 1, RawLine: string(text)})
			return nil // the body stays in lr, right where it starts
		}
		lines.add(string(text))
	}
}

// ReadAll tokenizes a bare header (.fm) into the list: every line to EOF is
// front matter — no fences, no body.
func (lines *FmLines) ReadAll(lr linereader.IoLineReader) error {
	for {
		if _, err := lines.ReadLine(lr); err == io.EOF {
			return nil
		} else if err != nil {
			return err
		}
	}
}

func NewFmLineReader(reader io.Reader, lines *FmLines) *FmLineReader {
	if lines == nil {
		lines = &FmLines{}
	}
	return &FmLineReader{
		Reader: linereader.NewLineReader(reader, 0, 0),
		Lines:  lines,
	}
}

func (flr *FmLineReader) ReadLine() (FmLineType, error) {
	if flr.Lines == nil {
		return FmLineInvalid, fmt.Errorf("field Lines not pointing to a valid FmLines")
	}
	return flr.Lines.ReadLine(flr.Reader)
}

func (flr *FmLineReader) ReadHeader() error {
	if flr.Lines == nil {
		return fmt.Errorf("field Lines not pointing to a valid FmLines")
	}
	return flr.Lines.ReadHeader(flr.Reader)
}

func (flr *FmLineReader) ReadAll() error {
	if flr.Lines == nil {
		return fmt.Errorf("field Lines not pointing to a valid FmLines")
	}
	return flr.Lines.ReadAll(flr.Reader)
}
