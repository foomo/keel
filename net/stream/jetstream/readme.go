package jetstream

import (
	"github.com/foomo/keel/markdown"
)

type (
	// publisher records a publisher created through [Stream.Publisher] for
	// [Readme].
	publisher struct {
		Namespace string
		Stream    string
		Subject   string
	}
	// subscriber records a subscriber created through [Stream.Subscriber]
	// for [Readme].
	subscriber struct {
		Namespace string
		Stream    string
		Subject   string
	}
)

// publishers and subscribers hold the process wide, deduplicated registry
// read by [Readme]. Access is not synchronized.
var (
	publishers  []publisher
	subscribers []subscriber
)

// Readme returns a markdown table of all publishers and subscribers created
// in this process, or an empty string if there are none.
func Readme() string {
	if len(publishers) == 0 && len(subscribers) == 0 {
		return ""
	}

	var rows [][]string

	md := &markdown.Markdown{}
	md.Println("### NATS")
	md.Println("")
	md.Println("List of all registered nats publishers & subscribers.")
	md.Println("")

	if len(publishers) > 0 {
		for _, value := range publishers {
			rows = append(rows, []string{
				markdown.Code(value.Namespace),
				markdown.Code(value.Stream),
				markdown.Code(value.Subject),
				markdown.Code("publish"),
			})
		}
	}

	if len(subscribers) > 0 {
		for _, value := range subscribers {
			rows = append(rows, []string{
				markdown.Code(value.Namespace),
				markdown.Code(value.Stream),
				markdown.Code(value.Subject),
				markdown.Code("subscribe"),
			})
		}
	}

	md.Table([]string{"Namespace", "Stream", "Subject", "Type"}, rows)

	return md.String()
}
