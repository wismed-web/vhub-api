package submit

type Post struct {
	Topic       string `json:"topic"`
	Category    string `json:"category"`
	Keywords    string `json:"keywords"`
	ContentHTML string `json:"content_html"`
	ContentTEXT string `json:"content_text"`
}
