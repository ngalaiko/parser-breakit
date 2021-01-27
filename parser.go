package parser

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// Parser parses articles from https://breakit.se
type Parser struct {
	crawler *crawler
	logger  *logger
}

// New creates a new parser.
func New(verbose bool) *Parser {
	return &Parser{
		crawler: newCrawler(),
		logger:  newLogger(verbose),
	}
}

// Article is a parsed article.
type Article struct {
	URL         *url.URL
	PublishedAt time.Time
	Title       string
	Subtitle    string
	Summary     string
}

// Parse starts parsing.
func (p *Parser) Parse(ctx context.Context, depth int64, concurrency int64) ([]*Article, error) {
	// todo: use concurrency

	startURL, _ := url.Parse("https://breakit.se")

	articles, err := p.parse(ctx, depth+1, startURL)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (p *Parser) parse(ctx context.Context, depth int64, url *url.URL) ([]*Article, error) {
	if depth < 0 {
		return nil, nil
	}

	p.logger.Debugf("parsing %s", url)

	article, links, err := p.parsePage(ctx, url)
	if err != nil {
		return nil, err
	}

	articles := []*Article{article}
	for _, link := range links {
		p.logger.Debugf("%s links to %s", url, link)

		linkedArticles, err := p.parse(ctx, depth-1, link)
		if err != nil {
			return nil, err
		}

		articles = append(articles, linkedArticles...)
	}

	return articles, nil
}

// returns:
// * a parsed article, if it's an article page
// * a list of links found on the page
func (p *Parser) parsePage(ctx context.Context, link *url.URL) (*Article, []*url.URL, error) {
	content, err := p.crawler.Crawl(ctx, link)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch '%s': %w", link, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse '%s': %w", link, err)
	}

	return extractContent(link, doc), extractLinks(link, doc), nil
}

func extractContent(src *url.URL, doc *goquery.Document) *Article {
	return &Article{
		URL: src,
	}
}

func extractLinks(src *url.URL, doc *goquery.Document) []*url.URL {
	links := []*url.URL{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			for _, attribute := range n.Attr {
				if attribute.Key != "href" {
					continue
				}

				link, err := url.Parse(attribute.Val)
				if err != nil {
					continue
				}

				if !isArticle(link) {
					continue
				}

				link.Scheme = src.Scheme
				link.Host = src.Host

				links = append(links, link)
			}
		}
	})

	return links
}

func isArticle(link *url.URL) bool {
	switch link.Hostname() {
	case "":
		fallthrough
	case "breakit.se":
		return strings.HasPrefix(link.Path, "/artikel/")
	default:
		return false
	}
}
