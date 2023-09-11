package markdown

import (
	"fmt"

	markdowntable "github.com/fbiville/markdown-table-formatter/pkg/markdown"
)

// Markdown output helper
type Markdown struct {
	value string
}

func (s *Markdown) Println(a ...any) {
	s.value += fmt.Sprintln(a...)
}

func (s *Markdown) Printf(format string, a ...any) {
	s.Println(fmt.Sprintf(format, a...))
}

func (s *Markdown) Print(a ...any) {
	s.value += fmt.Sprint(a...)
}

func (s *Markdown) String() string {
	return s.value
}

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
