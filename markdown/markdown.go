package markdown

import (
	"fmt"

	markdowntable "github.com/fbiville/markdown-table-formatter/pkg/markdown"
)

// Markdown accumulates markdown text in memory. The zero value is an empty
// document ready to use. It is not safe for concurrent use.
type Markdown struct {
	value string
}

// Println appends the operands formatted as by [fmt.Sprintln].
func (s *Markdown) Println(a ...any) {
	s.value += fmt.Sprintln(a...)
}

// Printf appends the operands formatted as by [fmt.Sprintf], followed by a
// newline.
func (s *Markdown) Printf(format string, a ...any) {
	s.Println(fmt.Sprintf(format, a...))
}

// Print appends the operands formatted as by [fmt.Sprint].
func (s *Markdown) Print(a ...any) {
	s.value += fmt.Sprint(a...)
}

// String returns the accumulated markdown.
func (s *Markdown) String() string {
	return s.value
}

// Table appends a pretty-printed markdown table with the given headers and
// rows, sorted alphabetically in ascending order. It panics if the table
// cannot be formatted.
func (s *Markdown) Table(headers []string, rows [][]string) {
	table, err := markdowntable.NewTableFormatterBuilder().
		WithAlphabeticalSortIn(markdowntable.ASCENDING_ORDER).
		WithPrettyPrint().
		Build(headers...).
		Format(rows)
	if err != nil {
		panic(err)
	}

	s.Print(table)
}
