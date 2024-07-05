package jetstream

import (
	"github.com/foomo/keel/markdown"
)

type (
	publisher struct {
		Namespace string
		Stream    string
		Subject   string
	}
	subscriber struct {
		Namespace string
		Stream    string
		Subject   string
	}
)

var (
	publishers  []publisher
	subscribers []subscriber
)

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
