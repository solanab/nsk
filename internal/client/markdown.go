package client

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var ansiEscape = regexp.MustCompile(
	`\x1b\[[0-9;:?]*[ -/]*[@-~]|\x1b[\]].*?(?:\x07|\x1b\\)|\x1b[@-Z\\-_]`,
)

// FormatPost renders a post as Markdown (title + floors). v1 strips ANSI.
func FormatPost(detail *PostDetail) string {
	if detail == nil {
		return ""
	}

	var builder strings.Builder
	if detail.Title != "" {
		builder.WriteString("# ")
		builder.WriteString(detail.Title)
		builder.WriteString("\n\n")
	}

	for i, floor := range detail.Floors {
		if i > 0 {
			builder.WriteString("\n\n")
		}

		writeFloor(&builder, floor)
	}

	return stripANSI(strings.TrimSpace(builder.String()))
}

// FormatPost renders a post as Markdown. It does not touch the network.
func (*Client) FormatPost(detail *PostDetail) string {
	return FormatPost(detail)
}

func writeFloor(builder *strings.Builder, floor Floor) {
	builder.WriteString("## #")
	builder.WriteString(strconv.Itoa(floor.Number))

	if floor.Author != "" {
		builder.WriteByte(' ')
		builder.WriteString(floor.Author)
	}

	builder.WriteString("\n\n")
	builder.WriteString(floor.Markdown)
}

func htmlToMarkdown(sel *goquery.Selection) string {
	if sel == nil || sel.Length() == 0 {
		return ""
	}

	var builder strings.Builder
	writeChildren(sel, &builder)

	return stripANSI(strings.TrimSpace(builder.String()))
}

func writeChildren(sel *goquery.Selection, builder *strings.Builder) {
	sel.Contents().Each(func(_ int, node *goquery.Selection) {
		writeNode(node, builder)
	})
}

func writeNode(node *goquery.Selection, builder *strings.Builder) {
	switch goquery.NodeName(node) {
	case "#text":
		builder.WriteString(node.Text())
	case "br":
		builder.WriteByte('\n')
	case "p":
		writeChildren(node, builder)
		builder.WriteString("\n\n")
	case "img":
		writeImage(node, builder)
	case "a":
		writeLink(node, builder)
	case "pre":
		writeFence(plainText(node), builder)
	case "div":
		writeDiv(node, builder)
	case "span":
		writeSpan(node, builder)
	default:
		writeChildren(node, builder)
	}
}

func writeDiv(node *goquery.Selection, builder *strings.Builder) {
	if node.HasClass("nsk-magic-tabs") {
		writeMagicTabs(node, builder)

		return
	}

	writeChildren(node, builder)
}

func writeSpan(node *goquery.Selection, builder *strings.Builder) {
	if node.AttrOr("data-ansicode", "") == "27" {
		builder.WriteByte('\x1b')

		return
	}

	writeChildren(node, builder)
}

func writeImage(node *goquery.Selection, builder *strings.Builder) {
	if node.HasClass("sticker") {
		builder.WriteString("[Sticker]")

		return
	}

	alt := node.AttrOr("alt", "")
	src := node.AttrOr("src", "")

	builder.WriteString("![")
	builder.WriteString(alt)
	builder.WriteString("](")
	builder.WriteString(src)
	builder.WriteString(")")
}

func writeLink(node *goquery.Selection, builder *strings.Builder) {
	var inner strings.Builder
	writeChildren(node, &inner)

	href := node.AttrOr("href", "")
	text := inner.String()

	if href == "" {
		builder.WriteString(text)

		return
	}

	builder.WriteByte('[')
	builder.WriteString(text)
	builder.WriteString("](")
	builder.WriteString(href)
	builder.WriteByte(')')
}

func writeMagicTabs(node *goquery.Selection, builder *strings.Builder) {
	titles := node.ChildrenFiltered(".nsk-magic-tab-title")
	bodies := node.ChildrenFiltered(".nsk-magic-tab-body")
	count := min(titles.Length(), bodies.Length())

	for i := range count {
		title := strings.TrimSpace(titles.Eq(i).Text())
		body := strings.TrimSpace(plainText(bodies.Eq(i)))

		if title != "" {
			body = title + "\n" + body
		}

		writeFence(body, builder)
	}
}

func writeFence(body string, builder *strings.Builder) {
	builder.WriteString("```\n")
	builder.WriteString(strings.TrimSpace(body))
	builder.WriteString("\n```\n\n")
}

func plainText(sel *goquery.Selection) string {
	var builder strings.Builder
	writePlain(sel, &builder)

	return builder.String()
}

func writePlain(sel *goquery.Selection, builder *strings.Builder) {
	sel.Contents().Each(func(_ int, node *goquery.Selection) {
		switch goquery.NodeName(node) {
		case "#text":
			builder.WriteString(node.Text())
		case "br":
			builder.WriteByte('\n')
		case "img":
			if node.HasClass("sticker") {
				builder.WriteString("[Sticker]")
			}
		case "span":
			if node.AttrOr("data-ansicode", "") == "27" {
				builder.WriteByte('\x1b')

				return
			}

			writePlain(node, builder)
		default:
			writePlain(node, builder)
		}
	})
}

func stripANSI(text string) string {
	return ansiEscape.ReplaceAllString(text, "")
}
