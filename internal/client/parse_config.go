package client

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var parseHTML = goquery.NewDocumentFromReader //nolint:gochecknoglobals // HTML parse test seam

func parseUser(html []byte) *UserInfo {
	doc, err := loadSeam(&parseHTML)(bytes.NewReader(html))
	if err != nil {
		return nil
	}

	src := configScript(doc)
	if src == "" {
		return nil
	}

	obj := cutFirstObject(src)
	if obj == nil {
		return nil
	}

	var cfg configJSON
	if err := json.Unmarshal(obj, &cfg); err != nil {
		return nil
	}

	return projectUser(cfg.User)
}

func configScript(doc *goquery.Document) string {
	var src string

	doc.Find("script").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		text := sel.Text()
		if !strings.Contains(text, "__config__") {
			return true
		}

		src = text

		return false
	})

	return src
}

func projectUser(raw *userJSON) *UserInfo {
	if raw == nil || (raw.ID == 0 && raw.Name == "") {
		return nil
	}

	return &UserInfo{
		ID:      raw.ID,
		Name:    raw.Name,
		Chicken: raw.Chicken,
		Level:   raw.Level,
	}
}

func cutFirstObject(src string) []byte {
	start := strings.Index(src, "{")
	if start < 0 {
		return nil
	}

	return scanJSONObject(src, start)
}

func scanJSONObject(src string, start int) []byte {
	depth := 0
	inString := false
	escape := false

	for idx := start; idx < len(src); idx++ {
		nextString, nextEscape, nextDepth, done := stepJSON(src[idx], inString, escape, depth)
		inString = nextString
		escape = nextEscape
		depth = nextDepth

		if done {
			return []byte(src[start : idx+1])
		}
	}

	return nil
}

func stepJSON(char byte, inString, escape bool, depth int) (bool, bool, int, bool) {
	if inString {
		still, nextEscape := stepJSONString(char, escape)

		return still, nextEscape, depth, false
	}

	switch char {
	case '"':
		return true, false, depth, false
	case '{':
		return false, false, depth + 1, false
	case '}':
		depth--

		return false, false, depth, depth == 0
	default:
		return false, false, depth, false
	}
}

func stepJSONString(char byte, escape bool) (bool, bool) {
	if escape {
		return true, false
	}

	switch char {
	case '\\':
		return true, true
	case '"':
		return false, false
	default:
		return true, false
	}
}
