package fmlines_test

import (
	"io"
	"strings"
	"testing"

	"github.com/pablo-botella/fmlines"
	"github.com/pablo-botella/linereader"
)

func reader(content string) *linereader.LineReader {
	return linereader.NewLineReader(strings.NewReader(content), 0, 0)
}

// kinds flattens the token types for a quick shape check.
func kinds(lines fmlines.FmLines) []fmlines.FmLineType {
	out := make([]fmlines.FmLineType, len(lines))
	for i, line := range lines {
		out[i] = line.Type
	}
	return out
}

func TestReadHeaderFenced(t *testing.T) {
	lr := reader("---\nname: demo\n\n# a comment\nmkskill:\n  pos: 20\n  in: \"*\"\n---\nbody line 1\nbody line 2\n")
	var lines fmlines.FmLines
	if err := lines.ReadHeader(lr); err != nil {
		t.Fatal(err)
	}

	want := []fmlines.FmLineType{
		fmlines.FmLineStartBlock,
		fmlines.FmLineKeyValue, // name: demo
		fmlines.FmLineEmpty,    // blank
		fmlines.FmLineEmpty,    // # a comment
		fmlines.FmLineSection,  // mkskill:
		fmlines.FmLineKeyValue, //   pos: 20
		fmlines.FmLineKeyValue, //   in: "*"
		fmlines.FmLineEndBlock,
	}
	got := kinds(lines)
	if len(got) != len(want) {
		t.Fatalf("token shape = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("token %d = %#x, want %#x (shape %v)", i, int(got[i]), int(want[i]), got)
		}
	}

	// raw views: Name/Value are slices of the raw line, untrimmed
	if name := lines[1].Name(); name != "name" {
		t.Errorf("Name() = %q, want name", name)
	}
	if value := lines[1].Value(); value != " demo" {
		t.Errorf("Value() = %q, want %q (raw, untrimmed)", value, " demo")
	}

	// the section tree: pos and in hang from mkskill, one level down
	section := lines[4]
	for _, child := range lines[5:7] {
		if child.Parent != section || child.Level != 1 {
			t.Errorf("%q: Parent/Level = %v/%d, want mkskill/1", child.RawLine, child.Parent, child.Level)
		}
	}

	// the body stayed in the reader, right where it starts
	body, err := lr.ReadLine()
	if err != nil || string(body) != "body line 1" {
		t.Fatalf("body after header = %q, %v; want body line 1", body, err)
	}
}

// TestReadHeaderNoFence: without an opening fence there is no front matter —
// the one consumed line comes back as the only token, the rest is untouched.
func TestReadHeaderNoFence(t *testing.T) {
	lr := reader("# just a markdown\nmore\n")
	var lines fmlines.FmLines
	if err := lines.ReadHeader(lr); err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 || lines[0].RawLine != "# just a markdown" {
		t.Fatalf("want the consumed line as the only token, got %v", kinds(lines))
	}
	rest, err := lr.ReadLine()
	if err != nil || string(rest) != "more" {
		t.Fatalf("rest = %q, %v; want more", rest, err)
	}
}

func TestReadHeaderUnclosed(t *testing.T) {
	var lines fmlines.FmLines
	if err := lines.ReadHeader(reader("---\nname: x\n")); err == nil {
		t.Fatal("want an unclosed-fence error, got none")
	}
}

func TestReadHeaderEmptyDocument(t *testing.T) {
	var lines fmlines.FmLines
	if err := lines.ReadHeader(reader("")); err != nil || len(lines) != 0 {
		t.Fatalf("empty document = %v, %v; want empty, nil", lines, err)
	}
}

// TestReadAll: a bare .fm — everything is header, no fences involved.
func TestReadAll(t *testing.T) {
	var lines fmlines.FmLines
	if err := lines.ReadAll(reader("mkskill:\n  pos: 30\n")); err != nil {
		t.Fatal(err)
	}
	got := kinds(lines)
	if len(got) != 2 || got[0] != fmlines.FmLineSection || got[1] != fmlines.FmLineKeyValue {
		t.Fatalf("token shape = %v", got)
	}
	if lines[1].Parent != lines[0] {
		t.Error("pos should hang from mkskill")
	}
}

// TestHugoListsSurvive: full-YAML lists are outside the dialect — they fall
// as Invalid, but raw, numbered and adopted by their section: nothing lost.
func TestHugoListsSurvive(t *testing.T) {
	content := "categories:\n" +
		"  - [Sports, Baseball]\n" +
		"  - Rivalries\n"
	var lines fmlines.FmLines
	if err := lines.ReadAll(reader(content)); err != nil {
		t.Fatal(err)
	}
	if lines[0].Type != fmlines.FmLineSection {
		t.Fatalf("categories should be a section, got %#x", int(lines[0].Type))
	}
	for _, item := range lines[1:] {
		if item.Type != fmlines.FmLineInvalid {
			t.Errorf("%q: type %#x, want Invalid", item.RawLine, int(item.Type))
		}
		if item.Parent != lines[0] {
			t.Errorf("%q: should hang from categories", item.RawLine)
		}
		if item.RawLine == "" {
			t.Error("raw line lost")
		}
	}
}

// TestChildrenField: the tree is doubly linked at birth — a section's
// Children holds what hangs from it, a leaf carries nil.
func TestChildrenField(t *testing.T) {
	var lines fmlines.FmLines
	if err := lines.ReadAll(reader("name: demo\nmkskill:\n  pos: 20\n  in: \"*\"\nother: x\n")); err != nil {
		t.Fatal(err)
	}
	section := lines[1] // mkskill:
	if section.Children == nil || len(*section.Children) != 2 {
		t.Fatalf("mkskill.Children = %v, want pos and in", section.Children)
	}
	for _, child := range *section.Children {
		if child.Parent != section {
			t.Errorf("%q: in Children but Parent does not point back", child.RawLine)
		}
	}
	if lines[0].Children != nil {
		t.Errorf("a leaf should carry nil Children, got %v", lines[0].Children)
	}
}

// TestDeepNesting: sections inside sections — Level counts, both links hold.
func TestDeepNesting(t *testing.T) {
	var lines fmlines.FmLines
	if err := lines.ReadAll(reader("a:\n  b:\n    c: 1\n  d: 2\n")); err != nil {
		t.Fatal(err)
	}
	a, b, c, d := lines[0], lines[1], lines[2], lines[3]
	if b.Parent != a || b.Level != 1 || c.Parent != b || c.Level != 2 {
		t.Errorf("nesting broken: b(%v,%d) c(%v,%d)", b.Parent, b.Level, c.Parent, c.Level)
	}
	if d.Parent != a || d.Level != 1 {
		t.Errorf("d should climb back to a: %v/%d", d.Parent, d.Level)
	}
	if len(*a.Children) != 2 || len(*b.Children) != 1 {
		t.Errorf("children counts: a=%d want 2, b=%d want 1", len(*a.Children), len(*b.Children))
	}
}

// TestFmLineReader: the bundle — reader plus list, methods without params;
// the body still waits in the reader after ReadHeader.
func TestFmLineReader(t *testing.T) {
	flr := fmlines.NewFmLineReader(strings.NewReader("---\nname: demo\n---\nbody\n"), nil)
	if err := flr.ReadHeader(); err != nil {
		t.Fatal(err)
	}
	if flr.Lines == nil || len(*flr.Lines) != 3 {
		t.Fatalf("Lines = %v, want fence+kv+fence", flr.Lines)
	}
	body, err := flr.Reader.ReadLine()
	if err != nil || string(body) != "body" {
		t.Fatalf("body = %q, %v; want body", body, err)
	}

	// an injected list is used as is
	var mine fmlines.FmLines
	flr = fmlines.NewFmLineReader(strings.NewReader("k: v\n"), &mine)
	if _, err := flr.ReadLine(); err != nil {
		t.Fatal(err)
	}
	if len(mine) != 1 || mine[0].Type != fmlines.FmLineKeyValue {
		t.Fatalf("injected list not used: %v", kinds(mine))
	}

	// a hand-built FmLineReader without Lines refuses politely
	broken := &fmlines.FmLineReader{Reader: reader("x: y\n")}
	if _, err := broken.ReadLine(); err == nil {
		t.Error("ReadLine with nil Lines should fail")
	}
	if err := broken.ReadHeader(); err == nil {
		t.Error("ReadHeader with nil Lines should fail")
	}
	if err := broken.ReadAll(); err == nil {
		t.Error("ReadAll with nil Lines should fail")
	}
}

// TestReadLineOneAtATime: the smallest step — one line in, its type out.
func TestReadLineOneAtATime(t *testing.T) {
	lr := reader("name: demo\nmkskill:\n")
	var lines fmlines.FmLines

	kind, err := lines.ReadLine(lr)
	if err != nil || kind != fmlines.FmLineKeyValue {
		t.Fatalf("first = %#x, %v; want KeyValue", int(kind), err)
	}
	kind, err = lines.ReadLine(lr)
	if err != nil || kind != fmlines.FmLineSection {
		t.Fatalf("second = %#x, %v; want Section", int(kind), err)
	}
	if _, err = lines.ReadLine(lr); err != io.EOF {
		t.Fatalf("end = %v; want io.EOF", err)
	}
	if len(lines) != 2 {
		t.Fatalf("list holds %d lines, want 2", len(lines))
	}
}
