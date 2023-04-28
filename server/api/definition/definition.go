package definition

type Post struct {
	Topic       string `json:"topic"`        // user input
	Type        string `json:"type"`         // type: post or reply
	Category    string `json:"category"`     // user input
	Keywords    string `json:"keywords"`     // user input
	ContentHTML string `json:"content_html"` // user input
	ContentTEXT string `json:"content_text"` // user input
}
