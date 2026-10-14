package statuspage

import (
	"fmt"
	"strings"
	"time"
)

type IncidentFeedItem struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PubDate     time.Time `json:"pub_date"`
	Link        string    `json:"link"`
}

type FeedGenerator struct {
	SiteTitle string
	SiteURL   string
}

func NewFeedGenerator(title, url string) *FeedGenerator {
	if title == "" {
		title = "PulseGrid Status Incidents"
	}
	if url == "" {
		url = "https://pulsegrid.io"
	}
	return &FeedGenerator{SiteTitle: title, SiteURL: url}
}

func (g *FeedGenerator) GenerateRSS(items []IncidentFeedItem) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "
")
	sb.WriteString(`<rss version="2.0">` + "
")
	sb.WriteString("<channel>
")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>
", g.SiteTitle))
	sb.WriteString(fmt.Sprintf("  <link>%s/status</link>
", g.SiteURL))
	sb.WriteString("  <description>Real-time incident updates and service disruptions</description>
")

	for _, it := range items {
		sb.WriteString("  <item>
")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>
", it.Title))
		sb.WriteString(fmt.Sprintf("    <link>%s</link>
", it.Link))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>
", it.Description))
		sb.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>
", it.PubDate.Format(time.RFC1123Z)))
		sb.WriteString(fmt.Sprintf("    <guid>%s</guid>
", it.ID))
		sb.WriteString("  </item>
")
	}

	sb.WriteString("</channel>
")
	sb.WriteString("</rss>
")
	return sb.String()
}

func (g *FeedGenerator) GenerateAtom(items []IncidentFeedItem) string {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "
")
	sb.WriteString(`<feed xmlns="http://www.w3.org/2005/Atom">` + "
")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>
", g.SiteTitle))
	sb.WriteString(fmt.Sprintf(`  <link href="%s/status/feed.atom" rel="self"/>`+"
", g.SiteURL))
	sb.WriteString(fmt.Sprintf("  <updated>%s</updated>
", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  <id>%s/status</id>
", g.SiteURL))

	for _, it := range items {
		sb.WriteString("  <entry>
")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>
", it.Title))
		sb.WriteString(fmt.Sprintf(`    <link href="%s"/>`+"
", it.Link))
		sb.WriteString(fmt.Sprintf("    <id>%s</id>
", it.ID))
		sb.WriteString(fmt.Sprintf("    <updated>%s</updated>
", it.PubDate.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("    <summary>%s</summary>
", it.Description))
		sb.WriteString("  </entry>
")
	}

	sb.WriteString("</feed>
")
	return sb.String()
}
