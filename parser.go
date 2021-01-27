package parser

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
)

// Parser parses articles from https://breakit.se
type Parser struct {
	crawler *crawler
	logger  *logger

	seen *sync.Map
}

// New creates a new parser.
func New(verbose bool) *Parser {
	return &Parser{
		crawler: newCrawler(),
		logger:  newLogger(verbose),
		seen:    &sync.Map{},
	}
}

// Article is a parsed article.
type Article struct {
	URL         *url.URL
	PublishedAt time.Time
	Title       string
	Preamble    string
	Summary     *string
	Links       []*url.URL
	Depth       int64
}

// Parse starts parsing.
func (p *Parser) Parse(ctx context.Context, depth int64, concurrency int64) ([]*Article, error) {
	startURL, _ := url.Parse("https://breakit.se")

	articlesStream := make(chan *Article)
	errorsStream := make(chan error)
	done := make(chan struct{})

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		p.parse(ctx, depth+1, -1, startURL, make(chan struct{}, concurrency), articlesStream, errorsStream, wg)
		wg.Done()
	}()

	go func() {
		wg.Wait()
		close(done)
	}()

	articles := []*Article{}
	for {
		select {
		case <-ctx.Done():
			return articles, nil
		case <-done:
			return articles, nil
		case article := <-articlesStream:
			if article.Depth == -1 {
				continue
			}
			articles = append(articles, article)
		case err := <-errorsStream:
			if err != nil {
				p.logger.Logf("error: %s", err)
			}

			return articles, err
		}
	}
}

var loc, _ = time.LoadLocation("Europe/Stockholm")

func (p *Parser) parse(ctx context.Context, depth int64, pathLength int64, url *url.URL, sem chan struct{}, articles chan *Article, errors chan error, wg *sync.WaitGroup) {
	if depth < 0 {
		return
	}

	if _, seen := p.seen.Load(url.String()); seen {
		return
	}
	p.seen.Store(url.String(), struct{}{})

	sem <- struct{}{}

	p.logger.Debugf("parsing %s", url)

	article, err := p.parsePage(ctx, url)
	if err != nil {
		errors <- err
		return
	}

	article.Depth = pathLength

	<-sem

	articles <- article

	if len(article.Links) == 0 {
		return
	}

	for _, link := range article.Links {
		link := link
		wg.Add(1)
		go func() {
			p.parse(ctx, depth-1, pathLength+1, link, sem, articles, errors, wg)
			wg.Done()
		}()
	}
}

// returns:
// * a parsed article, if it's an article page
// * a list of links found on the page
func (p *Parser) parsePage(ctx context.Context, link *url.URL) (*Article, error) {
	content, err := p.crawler.Crawl(ctx, link)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch '%s': %w", link, err)
	}

	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse '%s': %w", link, err)
	}

	article := &Article{
		URL:   link,
		Links: extractLinks(link, doc),
	}
	extractContent(article, doc)

	return article, nil
}

func extractContent(to *Article, doc *goquery.Document) {
	to.Title = doc.Find(".article__title").Text()
	to.Preamble = doc.Find(".article__preamble").Text()

	if s := doc.Find(".article__body").First().First().Text(); s != "" {
		to.Summary = &s
	}

	if date := doc.Find(".article__date"); date != nil && len(date.Nodes) != 0 {
		if datetime := getAttribute(date.Nodes[0], "datetime"); datetime != nil {
			if publishedAt, err := time.Parse("2006-01-02 15:04:05", *datetime); err == nil {
				to.PublishedAt = publishedAt.In(loc)
			}
		}
	}
}

func extractLinks(src *url.URL, doc *goquery.Document) []*url.URL {
	links := []*url.URL{}
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		for _, n := range s.Nodes {
			href := getAttribute(n, "href")
			if href == nil {
				continue
			}

			link, err := url.Parse(*href)
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
	})

	return links
}

func getAttribute(node *html.Node, name string) *string {
	for _, attribute := range node.Attr {
		if attribute.Key != name {
			continue
		}
		return &attribute.Val
	}

	return nil
}

func isArticle(link *url.URL) bool {
	switch link.Hostname() {
	case "www.breakit.se", "breakit.se", "":
		return strings.HasPrefix(link.Path, "/artikel/")
	default:
		return false
	}
}
