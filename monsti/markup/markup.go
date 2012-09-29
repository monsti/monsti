package markup

import "github.com/russross/blackfriday"

// MarkupToHTML processes the given markup string to HTML.
func MarkupToHTML(in string) string {
	return string(blackfriday.MarkdownCommon([]byte(in)))
}
