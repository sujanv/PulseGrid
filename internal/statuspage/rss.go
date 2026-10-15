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
	sb.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	sb.WriteString("<rss version=\"2.0\">\n")
	sb.WriteString("<channel>\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", g.SiteTitle))
	sb.WriteString(fmt.Sprintf("  <link>%s/status</link>\n", g.SiteURL))
	sb.WriteString("  <description>Real-time incident updates and service disruptions</description>\n")

	for _, it := range items {
		sb.WriteString("  <item>\n")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", it.Title))
		sb.WriteString(fmt.Sprintf("    <link>%s</link>\n", it.Link))
		sb.WriteString(fmt.Sprintf("    <description>%s</description>\n", it.Description))
		sb.WriteString(fmt.Sprintf("    <pubDate>%s</pubDate>\n", it.PubDate.Format(time.RFC1123Z)))
		sb.WriteString(fmt.Sprintf("    <guid>%s</guid>\n", it.ID))
		sb.WriteString("  </item>\n")
	}

	sb.WriteString("</channel>\n")
	sb.WriteString("</rss>\n")
	return sb.String()
}

func (g *FeedGenerator) GenerateAtom(items []IncidentFeedItem) string {
	var sb strings.Builder
	sb.WriteString("<?xml version=\"1.0\" encoding=\"utf-8\"?>\n")
	sb.WriteString("<feed xmlns=\"http://www.w3.org/2005/Atom\">\n")
	sb.WriteString(fmt.Sprintf("  <title>%s</title>\n", g.SiteTitle))
	sb.WriteString(fmt.Sprintf("  <link href=\"%s/status/feed.atom\" rel=\"self\"/>\n", g.SiteURL))
	sb.WriteString(fmt.Sprintf("  <updated>%s</updated>\n", time.Now().UTC().Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("  <id>%s/status</id>\n", g.SiteURL))

	for _, it := range items {
		sb.WriteString("  <entry>\n")
		sb.WriteString(fmt.Sprintf("    <title>%s</title>\n", it.Title))
		sb.WriteString(fmt.Sprintf("    <link href=\"%s\"/>\n", it.Link))
		sb.WriteString(fmt.Sprintf("    <id>%s</id>\n", it.ID))
		sb.WriteString(fmt.Sprintf("    <updated>%s</updated>\n", it.PubDate.Format(time.RFC3339)))
		sb.WriteString(fmt.Sprintf("    <summary>%s</summary>\n", it.Description))
		sb.WriteString("  </entry>\n")
	}

	sb.WriteString("</feed>\n")
	return sb.String()
}
