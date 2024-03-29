package main_test

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yuin/goldmark/ast"
	exast "github.com/yuin/goldmark/extension/ast"
	goldRend "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"golang.org/x/xerrors"

	markdown "github.com/rolflewis/goldmark-down"
)

var caller testRenderCaller

func init() {
	caller = testRenderCaller{
		funcs: make(map[ast.NodeKind]goldRend.NodeRendererFunc),
	}
	markdown.NewMarkdown().RegisterFuncs(caller)
}

func runTestCase(t *testing.T, doc *ast.Document, source string) {
	sourceBytes := []byte(source)
	got, err := caller.Render(sourceBytes, doc)
	if err != nil {
		t.Fatal(err)
	}

	if got != source {
		doc.Dump(sourceBytes, 4)
		t.Fatal(cmp.Diff(source, got))
	}
}

func TestRenderer_SimpleHeaderWithText(t *testing.T) {
	source := "# heading\n\nthis is text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newHeader(true, source, 1, "heading"))
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{newText(source, "this is text", false)}))
	runTestCase(t, doc, source)
}

func TestRenderer_MultipleHeaders(t *testing.T) {
	source := "# heading 1\n## heading 2\n# heading 3\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newHeader(true, source, 1, "heading 1"))
	doc.AppendChild(doc, newHeader(false, source, 2, "heading 2"))
	doc.AppendChild(doc, newHeader(false, source, 1, "heading 3"))
	runTestCase(t, doc, source)
}

func TestRenderer_Emphasis(t *testing.T) {
	source := "this is a *line* of text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "this is a ", false),
		newEmphasis(source, "line", 1),
		newText(source, " of text", false),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_CodeSpan(t *testing.T) {
	source := "this is a `code span` in text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "this is a ", false),
		newCodeSpan([]ast.Node{
			newText(source, "code span", false),
		}),
		newText(source, " in text", false),
	}))
	runTestCase(t, doc, source)
}

func newCodeSpan(parts []ast.Node) *ast.CodeSpan {
	span := ast.NewCodeSpan()
	for _, p := range parts {
		span.AppendChild(span, p)
	}
	return span
}

func TestRenderer_HeavyEmphasis(t *testing.T) {
	source := "this is a **line** of text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "this is a ", false),
		newEmphasis(source, "line", 2),
		newText(source, " of text", false),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_BasicOrderedList(t *testing.T) {
	source := "1. element 1\n2. element 2\n3. element 3\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newList(true, true, 1, '.', []ast.Node{
		newListItem(true, 3, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 1", false)})}),
		newListItem(false, 3, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 2", false)})}),
		newListItem(false, 3, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 3", false)})}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_NestedOrderedList_Tabs(t *testing.T) {
	source := "1. element 1\n   1. element 2\n      1. element 3\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newList(true, true, 1, '.', []ast.Node{
		newListItem(true, 3, []ast.Node{
			newTextBlock(false, []ast.Node{newText(source, "element 1", false)}),
			newList(false, true, 1, '.', []ast.Node{
				newListItem(false, 3, []ast.Node{
					newTextBlock(false, []ast.Node{newText(source, "element 2", false)}),
					newList(false, true, 1, '.', []ast.Node{
						newListItem(false, 3, []ast.Node{
							newTextBlock(false, []ast.Node{newText(source, "element 3", false)}),
						}),
					}),
				}),
			}),
		}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_UnorderedList(t *testing.T) {
	source := "- element 1\n- element 2\n- element 3\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newList(true, true, 0, '-', []ast.Node{
		newListItem(true, 2, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 1", false)})}),
		newListItem(false, 2, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 2", false)})}),
		newListItem(false, 2, []ast.Node{newTextBlock(false, []ast.Node{newText(source, "element 3", false)})}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_BlockQuotes(t *testing.T) {
	source := "> quote 1\nquote 2\nquote 3\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newBlockquote(true, []ast.Node{
		newParagraph(true, source, []ast.Node{
			newText(source, "quote 1", true),
			newText(source, "quote 2", true),
			newText(source, "quote 3", false),
		}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_Link(t *testing.T) {
	source := "[link label](guide1)\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newLink("", "guide1", []ast.Node{
			newText(source, "link label", false),
		}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_ImageLink(t *testing.T) {
	source := "![link label](link_destination.png)\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		ast.NewImage(newLink("", "link_destination.png", []ast.Node{
			newText(source, "link label", false),
		})),
	}))
	runTestCase(t, doc, source)
}

func newTableCell(prevBlank bool, parts []ast.Node) *exast.TableCell {
	cell := exast.NewTableCell()
	cell.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		cell.AppendChild(cell, p)
	}
	return cell
}

func newTableRow(prevBlank bool, parts []ast.Node) *exast.TableRow {
	row := exast.NewTableRow(nil)
	row.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		row.AppendChild(row, p)
	}
	return row
}

func newTableHeader(prevBlank bool, row *exast.TableRow) *exast.TableHeader {
	head := exast.NewTableHeader(row)
	head.SetBlankPreviousLines(prevBlank)
	return head
}

func newTable(prevBlank bool, parts []ast.Node) *exast.Table {
	table := exast.NewTable()
	table.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		table.AppendChild(table, p)
	}
	return table
}

func TestRenderer_Table(t *testing.T) {
	source := "| foo | bar |\n| --- | --- |\n| baz | bim |\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newTable(true, []ast.Node{
		newTableHeader(false, newTableRow(false, []ast.Node{
			newTableCell(false, []ast.Node{
				newText(source, "foo", false),
			}),
			newTableCell(false, []ast.Node{
				newText(source, "bar", false),
			}),
		})),
		newTableRow(false, []ast.Node{
			newTableCell(false, []ast.Node{
				newText(source, "baz", false),
			}),
			newTableCell(false, []ast.Node{
				newText(source, "bim", false),
			}),
		}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_Strikethrough(t *testing.T) {
	source := "this is ~~strikethrough-ed~~ text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "this is ", false),
		newStrikethrough([]ast.Node{
			newText(source, "strikethrough-ed", false),
		}),
		newText(source, " text", false),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_Strikethrough_ContainingInline(t *testing.T) {
	source := "this is ~~strikethrough-ed *emphasized*~~ text\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "this is ", false),
		newStrikethrough([]ast.Node{
			newText(source, "strikethrough-ed ", false),
			newEmphasis(source, "emphasized", 1),
		}),
		newText(source, " text", false),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_Task_Unchecked(t *testing.T) {
	source := "- [ ] task 1\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newList(true, true, 0, '-', []ast.Node{
		newListItem(true, 2, []ast.Node{newTextBlock(false, []ast.Node{
			exast.NewTaskCheckBox(false),
			newText(source, "task 1", false),
		})}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_Task_Checked(t *testing.T) {
	source := "- [x] task 1\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newList(true, true, 0, '-', []ast.Node{
		newListItem(true, 2, []ast.Node{newTextBlock(false, []ast.Node{
			exast.NewTaskCheckBox(true),
			newText(source, "task 1", false),
		})}),
	}))
	runTestCase(t, doc, source)
}

func TestRenderer_FencedCode(t *testing.T) {
	source := "```json\n{\"att\":1}\n```\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, ast.NewFencedCodeBlock(
		newText(source, "json\n{\"att\":1}", false),
	))
	runTestCase(t, doc, source)
}

// func TestRenderer_UnfencedCode(t *testing.T) {
// 	source := "\t{\n\t\t\"att\":1\n\t}\n"
// 	doc := ast.NewDocument()
// 	doc.AppendChild(doc, newCodeBlock([]ast.Node{
// 		newParagraph(false, source, []ast.Node{
// 			newText(source, "{\n\t\"att\":1\n\t}", false),
// 		}),
// 	}))
// 	runTestCase(t, doc, source)
// }

// func newCodeBlock(parts []ast.Node) *ast.CodeBlock {
// 	code := ast.NewCodeBlock()
// 	for _, p := range parts {
// 		code.AppendChild(code, p)
// 	}
// 	return code
// }

func TestRenderer_ThematicBreak(t *testing.T) {
	source := "text1\n--------\n\ntext2\n"
	doc := ast.NewDocument()
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "text1", false),
	}))
	doc.AppendChild(doc, ast.NewThematicBreak())
	doc.AppendChild(doc, newParagraph(true, source, []ast.Node{
		newText(source, "text2", false),
	}))
	runTestCase(t, doc, source)
}

// AST Builder Helpers

func newHeader(prevBlank bool, source string, level int, contents string) *ast.Heading {
	txt := newText(source, contents, false)
	head := ast.NewHeading(level)
	head.SetBlankPreviousLines(prevBlank)
	head.AppendChild(head, txt)
	return head
}

func newParagraph(prevBlank bool, source string, parts []ast.Node) *ast.Paragraph {
	para := ast.NewParagraph()
	para.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		para.AppendChild(para, p)
	}
	return para
}

func newText(source, contents string, soft bool) *ast.Text {
	index := strings.Index(source, contents)
	seg := text.NewSegment(index, index+len(contents))
	txt := ast.NewTextSegment(seg)
	txt.SetSoftLineBreak(soft)
	return txt
}

func newEmphasis(source, contents string, level int) *ast.Emphasis {
	emp := ast.NewEmphasis(level)
	emp.AppendChild(emp, newText(source, contents, false))
	return emp
}

func newList(prevBlank, tight bool, start int, marker byte, parts []ast.Node) *ast.List {
	list := ast.NewList(marker)
	list.IsTight = tight
	list.Start = start
	list.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		list.AppendChild(list, p)
	}
	return list
}

func newListItem(prevBlank bool, offset int, parts []ast.Node) *ast.ListItem {
	item := ast.NewListItem(offset)
	item.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		item.AppendChild(item, p)
	}
	return item
}

func newTextBlock(prevBlank bool, parts []ast.Node) *ast.TextBlock {
	block := ast.NewTextBlock()
	block.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		block.AppendChild(block, p)
	}
	return block
}

func newBlockquote(prevBlank bool, parts []ast.Node) *ast.Blockquote {
	quote := ast.NewBlockquote()
	quote.SetBlankPreviousLines(prevBlank)
	for _, p := range parts {
		quote.AppendChild(quote, p)
	}
	return quote
}

func newLink(title, dest string, parts []ast.Node) *ast.Link {
	link := ast.NewLink()
	link.Title = []byte(title)
	link.Destination = []byte(dest)
	for _, p := range parts {
		link.AppendChild(link, p)
	}
	return link
}

func newStrikethrough(parts []ast.Node) *exast.Strikethrough {
	strike := exast.NewStrikethrough()
	for _, p := range parts {
		strike.AppendChild(strike, p)
	}
	return strike
}

// Struct Helpers

type testRenderCaller struct {
	funcs map[ast.NodeKind]goldRend.NodeRendererFunc
}

func (rc testRenderCaller) Register(k ast.NodeKind, f goldRend.NodeRendererFunc) {
	rc.funcs[k] = f
}

func (rc testRenderCaller) Render(source []byte, n ast.Node) (string, error) {
	out := bytes.Buffer{}
	buf := bufio.NewWriter(&out)

	if err := ast.Walk(n, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		var err error
		s := ast.WalkStatus(ast.WalkContinue)
		if f := rc.funcs[n.Kind()]; f != nil {
			s, err = f(buf, source, n, entering)
			if err != nil {
				return s, xerrors.Errorf(": %w", err)
			}
		}
		return s, nil
	}); err != nil {
		return "", xerrors.Errorf(": %w", err)
	}

	buf.Flush()
	res, err := io.ReadAll(&out)
	if err != nil {
		return "", xerrors.Errorf(": %w", err)
	}

	return string(res), nil
}
