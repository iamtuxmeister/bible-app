package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/net/html"
)

type ApiResponse struct {
	Query       string  `json:"query"`
	Canonical   string  `json:"canonical"`
	Parsed      [][]int `json:"parsed"`
	PassageMeta []struct {
		Canonical    string `json:"canonical"`
		ChapterStart []int  `json:"chapter_start"`
		ChapterEnd   []int  `json:"chapter_end"`
		PrevVerse    int    `json:"prev_verse"`
		NextVerse    int    `json:"next_verse"`
		PrevChapter  []int  `json:"prev_chapter"`
		NextChapter  []int  `json:"next_chapter"`
	} `json:"passage_meta"`
	Passages []string `json:"passages"`
}

type SearchResponse struct {
	Page    int `json:"page"`
	Total   int `json:"total_results"`
	Results []struct {
		Reference string `json:"reference"`
		Content   string `json:"content"`
	} `json:"results"`
}

type Verse struct {
	Reference string `json:"reference"`
	Color     string `json:"color"`
}

type HighlightedVersesResponse struct {
	Verses []Verse `json:"verses"`
} // {verses: [{"reference": "v43003016", "color": "bg-red-600"}]}

func HighlighedVersesHandler(w http.ResponseWriter, r *http.Request) {

	result := HighlightedVersesResponse{
		Verses: []Verse{
			Verse{Reference: "v43003016", Color: "bg-red-600"},
			Verse{Reference: "v43003017", Color: "bg-violet-600"}},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func SearchRequestHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	search := p.Get("search")
	page := "1"
	pageSize := "20"
	params := url.Values{}
	params.Add("q", search)
	params.Add("page", page)
	params.Add("page-size", pageSize)
	api_key := os.Getenv("API_KEY")
	baseURL := "https://api.esv.org/v3/passage/search/"

	// ------------ This is the same as the ApiRequestHandler function ------------- \\

	urlWithParams := baseURL + "?" + params.Encode()

	//log.Println(urlWithParams)
	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Token "+api_key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result SearchResponse                             // Needs to be changed to SearchResponse from ApiResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	//w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)

}

func ApiRequestHandler(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	verse := "John+3:16-21"
	numbers := "false"
	headings := "false"
	extras := "false"

	params := url.Values{}
	if len(p.Get("verse")) > 0 {
		verse = p.Get("verse")
	}
	if p.Get("numbers") == "true" {
		numbers = "true"
	}
	if p.Get("headings") == "true" {
		headings = "true"
	}
	if p.Get("extras") == "true" {
		extras = "true"
	}
	baseURL := "https://api.esv.org/v3/passage/html/"
	api_key := os.Getenv("API_KEY")
	params.Add("q", verse)
	params.Add("include-passage-references", "true")
	params.Add("include-verse-anchors", "true")
	params.Add("include-chapter-numbers", numbers)
	params.Add("include-verse-numbers", numbers)
	params.Add("include-headings", headings)
	params.Add("include-subheadings", headings)
	params.Add("include-footnotes", extras)
	params.Add("include-audio-link", extras)
	// Encode the query parameters and append them to the base URL

	urlWithParams := baseURL + "?" + params.Encode()

	//log.Println(urlWithParams)
	req, err := http.NewRequest("GET", urlWithParams, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Add("Authorization", "Token "+api_key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	var result ApiResponse
	if err := json.Unmarshal(body, &result); err != nil { // Parse []byte to go struct pointer
		fmt.Println("Can not unmarshal JSON")
	}
	text := strings.Join(result.Passages, "")
	fmt.Println(text)
	wrapped, err := wrapVerses(text)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(wrapped)
	result.Passages = []string{wrapped}
	//w.Header().Set("Content-Type", "text/html")
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
	//w.Write([]byte("<html><head><meta charset=\"UTF-8\"></head><body>"))
	//for _, str := range result.Passages {
	//	_, err := w.Write([]byte(str + "\n")) // Add newline for separation
	//	if err != nil {
	//		http.Error(w, "Error writing to response", http.StatusInternalServerError)
	//		return
	//	}
	//log.Println(str)
	//}
	//w.Write([]byte("</body></html>"))

}

// func wrapVerses(htmlInput string) (string, error) {
// 	doc, err := html.Parse(strings.NewReader(htmlInput))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	var traverse func(n *html.Node)
// 	traverse = func(n *html.Node) {
// 		for c := n.FirstChild; c != nil; {
// 			next := c.NextSibling
//
// 			if isVerseAnchorNode(c) {
// 				verseID := getVerseID(c)
// 				var nodesToWrap []*html.Node
// 				nodesToWrap = append(nodesToWrap, c)
//
// 				sib := c.NextSibling
// 				for sib != nil {
// 					// Stop if next anchor or next verse number is found
// 					if isVerseAnchorNode(sib) || isVerseNumberNode(sib) {
// 						break
// 					}
// 					nodesToWrap = append(nodesToWrap, sib)
// 					sib = sib.NextSibling
// 				}
// 				span := &html.Node{
// 					Type: html.ElementNode,
// 					Data: "span",
// 					Attr: []html.Attribute{
// 						{Key: "class", Val: "verse"},
// 						{Key: "data-verse", Val: verseID},
// 					},
// 				}
//
// 				for _, node := range nodesToWrap {
// 					if node.Parent != nil {
// 						node.Parent.RemoveChild(node)
// 					}
// 					span.AppendChild(node)
// 				}
//
// 				n.InsertBefore(span, sib)
// 				c = span.NextSibling // continue after the inserted span
// 			} else {
// 				traverse(c)
// 				c = next
// 			}
// 		}
// 	}
//
// 	traverse(doc)
//
// 	var buf bytes.Buffer
// 	if err := html.Render(&buf, doc); err != nil {
// 		return "", err
// 	}
// 	return buf.String(), nil
// }

// ----------------------------
// wrapInlineVersesInNode handles a generic element (e.g. <p>, but we skip <h3> now).
//
// It scans its direct children. Whenever it sees <a class="va" rel="vXXXXX">, it pulls
// that <a> into the output and then collects all subsequent siblings (text, <sup>, etc.)
// up to—but not including—the next verse‐anchor or next verse‐number <b>. Those collected
// nodes become the contents of one <span class="verse" data-verse="vXXXXX">…</span>. A
// verse‐number <b> always flushes any span under construction (closing it) and is itself
// emitted directly (not wrapped). All other nodes pass through unchanged.
func wrapInlineVersesInNode(n *html.Node) {
	// 1) Collect original children in a slice
	var origChildren []*html.Node
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		origChildren = append(origChildren, c)
	}

	// 2) Build a new list "rebuilt" to replace n’s children
	var rebuilt []*html.Node
	i := 0
	for i < len(origChildren) {
		ch := origChildren[i]

		switch {
		// A) If it’s a verse‐number <b class="verse-num">…</b>, emit it as‐is
		case isVerseNumberNode(ch):
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)
			i++

		// B) If it’s a verse anchor <a class="va" rel="vXXXXX">…</a>:
		case isVerseAnchorNode(ch):
			// 1) Remove it from its parent and append to rebuilt
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)

			// 2) Record verseID
			verseID := getVerseID(ch)

			// 3) Collect all siblings until next anchor or verse‐number
			j := i + 1
			var toWrap []*html.Node
			for ; j < len(origChildren); j++ {
				sib := origChildren[j]
				if isVerseAnchorNode(sib) || isVerseNumberNode(sib) {
					break
				}
				toWrap = append(toWrap, sib)
			}

			// 4) If any nodes to wrap, create <span class="verse" data-verse="…">…</span>
			if len(toWrap) > 0 {
				spanNode := &html.Node{
					Type: html.ElementNode,
					Data: "span",
					Attr: []html.Attribute{
						{Key: "class", Val: "verse"},
						{Key: "data-verse", Val: verseID},
					},
				}
				for _, w := range toWrap {
					if w.Parent != nil {
						w.Parent.RemoveChild(w)
					}
					spanNode.AppendChild(w)
				}
				rebuilt = append(rebuilt, spanNode)
			}
			i = j

		// C) Otherwise, pass through
		default:
			if ch.Parent != nil {
				ch.Parent.RemoveChild(ch)
			}
			rebuilt = append(rebuilt, ch)
			i++
		}
	}

	// 3) Replace n’s old children with rebuilt list
	for old := n.FirstChild; old != nil; old = n.FirstChild {
		n.RemoveChild(old)
	}
	for _, c := range rebuilt {
		n.AppendChild(c)
	}
}

// -------------------
// wrapInlineVersesInIndentBlock handles <p class="block-indent">.
//
// It looks for each <span class="line"> or <span class="indent line"> inside it,
// and calls wrapInlineVersesInNode on those spans. Everything else is untouched.
func wrapInlineVersesInIndentBlock(p *html.Node) {
	for c := p.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "span" {
			for _, a := range c.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "line") {
					wrapInlineVersesInNode(c)
					break
				}
			}
		}
	}
}

// -------------------
// wrapVerses: traverses the entire document, but only applies wrapping inside <p> (and indent).
// <h3> is skipped entirely (no wrapping inside).
func wrapVerses(htmlInput string) (string, error) {
	// 1) Parse into a full document tree
	doc, err := html.Parse(strings.NewReader(htmlInput))
	if err != nil {
		return "", err
	}

	// 2) Recursively walk
	var recurse func(n *html.Node)
	recurse = func(n *html.Node) {
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			// If it’s a <p> element:
			if c.Type == html.ElementNode && c.Data == "p" {
				// Check if class="block-indent"
				isIndent := false
				for _, a := range c.Attr {
					if a.Key == "class" && strings.Contains(a.Val, "block-indent") {
						isIndent = true
						break
					}
				}

				if isIndent {
					// First recurse inside so nested spans are processed
					recurse(c)
					// Then apply verse‐wrapping to each span.line inside
					wrapInlineVersesInIndentBlock(c)
				} else {
					// Non‐indented <p>: wrap directly, then recurse deeper for footnotes, etc.
					wrapInlineVersesInNode(c)
					recurse(c)
				}

			} else {
				// Any other node (including <h3>) is left alone; just recurse into children
				recurse(c)
			}
		}
	}

	recurse(doc)

	// 3) Find the <body> node
	var body *html.Node
	var findBody func(n *html.Node)
	findBody = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "body" {
			body = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if body == nil {
				findBody(c)
			}
		}
	}
	findBody(doc)
	if body == nil {
		// In the unlikely event there's no <body>, just render entire doc
		var fullBuf bytes.Buffer
		if err := html.Render(&fullBuf, doc); err != nil {
			return "", err
		}
		return fullBuf.String(), nil
	}

	// 4) Render only each child of <body> (so no <html> or <body> wrappers)
	var buf bytes.Buffer
	for c := body.FirstChild; c != nil; c = c.NextSibling {
		html.Render(&buf, c)
	}
	return buf.String(), nil
}

func isVerseAnchorNode(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "va") {
				for _, a := range n.Attr {
					if a.Key == "rel" && strings.HasPrefix(a.Val, "v") {
						return true
					}
				}
			}
		}
	}
	return false
}

func isVerseNumberNode(n *html.Node) bool {
	if n.Type == html.ElementNode && n.Data == "b" {
		for _, attr := range n.Attr {
			if attr.Key == "class" && strings.Contains(attr.Val, "verse-num") {
				return true
			}
		}
	}
	return false
}

func getVerseID(n *html.Node) string {
	for _, attr := range n.Attr {
		if attr.Key == "rel" && strings.HasPrefix(attr.Val, "v") {
			return attr.Val
		}
	}
	return ""
}

// func wrapVerses(htmlInput string) (string, error) {
// 	// Determines end of a verse section
// 	isEndOfVerse := func(n *html.Node) bool {
// 		// Naive check — end at next <b class="verse-num"> or block
// 		if n.Type == html.ElementNode && n.Data == "b" {
// 			for _, attr := range n.Attr {
// 				if attr.Key == "class" && strings.Contains(attr.Val, "verse-num") {
// 					return true
// 				}
// 			}
// 		}
// 		if n.Type == html.ElementNode && (n.Data == "p" || n.Data == "div") {
// 			return true
// 		}
// 		return false
// 	}
//
// 	doc, err := html.Parse(strings.NewReader(htmlInput))
// 	if err != nil {
// 		return "", err
// 	}
//
// 	var wrapCurrentVerse func(*html.Node)
// 	var currentVerseID string
//
// 	// Buffers for output and verses
// 	var output bytes.Buffer
//
// 	// Walk the DOM and collect verse chunks
// 	var traverse func(n *html.Node)
// 	traverse = func(n *html.Node) {
// 		if n.Type == html.ElementNode && n.Data == "a" {
// 			var isVa bool
// 			for _, attr := range n.Attr {
// 				if attr.Key == "class" && strings.Contains(attr.Val, "va") {
// 					isVa = true
// 				}
// 				if attr.Key == "rel" && strings.HasPrefix(attr.Val, "v") {
// 					currentVerseID = attr.Val
// 				}
// 			}
// 			if isVa && currentVerseID != "" {
// 				// Start of verse - wrap next siblings
// 				wrapCurrentVerse(n)
// 				currentVerseID = ""
// 				return
// 			}
// 		}
// 		// Continue traversal
// 		for c := n.FirstChild; c != nil; c = c.NextSibling {
// 			traverse(c)
// 		}
// 	}
//
// 	// Wrap relevant siblings of the <a class="va"...> in a <span>
// 	wrapCurrentVerse = func(aNode *html.Node) {
// 		var verseNodes []*html.Node
//
// 		// Collect nodes that are part of the verse
// 		for n := aNode; n != nil; n = n.NextSibling {
// 			verseNodes = append(verseNodes, n)
// 			if isEndOfVerse(n) {
// 				break
// 			}
// 		}
//
// 		// Build the wrapper <span class="verse" data-verse="..." />
// 		span := &html.Node{
// 			Type: html.ElementNode,
// 			Data: "span",
// 			Attr: []html.Attribute{
// 				{Key: "class", Val: "verse"},
// 				{Key: "data-verse", Val: currentVerseID},
// 			},
// 		}
// 		for _, node := range verseNodes {
// 			span.AppendChild(node)
// 		}
//
// 		// Replace in the tree
// 		parent := aNode.Parent
// 		if parent != nil {
// 			parent.InsertBefore(span, verseNodes[0])
// 			for _, node := range verseNodes {
// 				parent.RemoveChild(node)
// 			}
// 		}
// 	}
//
// 	// Traverse and transform the DOM
// 	traverse(doc)
//
// 	// Render the modified DOM back to HTML
// 	if err := html.Render(&output, doc); err != nil {
// 		return "", err
// 	}
//
// 	return output.String(), nil
// }

// func renderToken(tok html.Token) string {
// 	var buf bytes.Buffer
// 	node := &html.Node{
// 		Type: html.ElementNode,
// 		Data: tok.Data,
// 		Attr: tok.Attr,
// 	}
// 	html.Render(&buf, node)
// 	return buf.String()
// }
//
// func wrapVerses(htmlInput string) (string, error) {
// 	tokenizer := html.NewTokenizer(strings.NewReader(htmlInput))
// 	var output bytes.Buffer
// 	var verseBuffer bytes.Buffer
// 	var currentVerseID string
// 	collecting := false
//
// 	flushVerse := func() {
// 		if currentVerseID != "" {
// 			output.WriteString(fmt.Sprintf(`<span class="verse" data-verse="%s">%s</span>`, currentVerseID, verseBuffer.String()))
// 			verseBuffer.Reset()
// 			currentVerseID = ""
// 		}
// 	}
//
// 	for {
// 		tt := tokenizer.Next()
// 		if tt == html.ErrorToken {
// 			if tokenizer.Err() == io.EOF {
// 				flushVerse()
// 				break
// 			}
// 			return "", tokenizer.Err()
// 		}
//
// 		tok := tokenizer.Token()
//
// 		if tt == html.StartTagToken && tok.Data == "a" {
// 			isVa := false
// 			verseID := ""
// 			for _, attr := range tok.Attr {
// 				if attr.Key == "class" && strings.Contains(attr.Val, "va") {
// 					isVa = true
// 				}
// 				if attr.Key == "rel" && strings.HasPrefix(attr.Val, "v") {
// 					verseID = attr.Val
// 				}
// 			}
// 			if isVa && verseID != "" {
// 				flushVerse()
// 				currentVerseID = verseID
// 				collecting = true
// 			}
// 		}
//
// 		rendered := renderToken(tok)
//
// 		if collecting {
// 			if tt == html.EndTagToken && (tok.Data == "p" || tok.Data == "div") {
// 				flushVerse()
// 				collecting = false
// 				output.WriteString(rendered)
// 				continue
// 			}
// 			verseBuffer.WriteString(rendered)
// 		} else {
// 			output.WriteString(rendered)
// 		}
// 	}
//
// 	return output.String(), nil
// }

// func wrapVerses(htmlInput string) (string, error) {
// 	tokenizer := html.NewTokenizer(strings.NewReader(htmlInput))
// 	var output bytes.Buffer
// 	var currentVerseID string
// 	var collecting bool
// 	var verseBuffer bytes.Buffer
//
// 	flushVerse := func() {
// 		if currentVerseID != "" {
// 			output.WriteString(fmt.Sprintf(`<span class="verse" data-verse="%s">%s</span>`, currentVerseID, verseBuffer.String()))
// 			verseBuffer.Reset()
// 			currentVerseID = ""
// 		}
// 	}
//
// 	for {
// 		tokenType := tokenizer.Next()
// 		if tokenType == html.ErrorToken {
// 			err := tokenizer.Err()
// 			if err == io.EOF {
// 				flushVerse()
// 				break
// 			}
// 			return "", err
// 		}
//
// 		tok := tokenizer.Token()
//
// 		if tokenType == html.StartTagToken && tok.Data == "a" {
// 			// Check if this is a verse anchor
// 			isVa := false
// 			for _, attr := range tok.Attr {
// 				if attr.Key == "class" && strings.Contains(attr.Val, "va") {
// 					isVa = true
// 				}
// 			}
// 			if isVa {
// 				// Extract verse ID
// 				for _, attr := range tok.Attr {
// 					if attr.Key == "rel" && strings.HasPrefix(attr.Val, "v") {
// 						flushVerse()
// 						currentVerseID = attr.Val // Remove the leading "v"
// 						collecting = true
// 						break
// 					}
// 				}
// 			}
// 		}
//
// 		// Buffer or write depending on state
// 		rendered := renderToken(tok)
// 		if collecting {
// 			verseBuffer.WriteString(rendered)
// 		} else {
// 			output.WriteString(rendered)
// 		}
// 	}
// 	return output.String(), nil
// }

// func renderToken(tok html.Token) string {
// 	var b strings.Builder
// 	switch tok.Type {
// 	case html.StartTagToken:
// 		b.WriteString("<" + tok.Data)
// 		for _, attr := range tok.Attr {
// 			b.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
// 		}
// 		b.WriteString(">")
// 	case html.EndTagToken:
// 		b.WriteString("</" + tok.Data + ">")
// 	case html.SelfClosingTagToken:
// 		b.WriteString("<" + tok.Data)
// 		for _, attr := range tok.Attr {
// 			b.WriteString(fmt.Sprintf(` %s="%s"`, attr.Key, attr.Val))
// 		}
// 		b.WriteString(" />")
// 	case html.TextToken:
// 		b.WriteString(tok.Data)
// 	case html.CommentToken:
// 		b.WriteString("<!--" + tok.Data + "-->")
// 	case html.DoctypeToken:
// 		b.WriteString("<!DOCTYPE " + tok.Data + ">")
// 	}
// 	return b.String()
// }
