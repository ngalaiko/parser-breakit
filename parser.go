package parser

import (
	"context"
	"fmt"
	"net/url"
	"time"
)

// Parser parses articles from https://breakit.se
type Parser struct{}

// New creates a new parser.
func New() *Parser {
	return &Parser{}
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

	articles, err := p.parse(ctx, depth, startURL)
	if err != nil {
		return nil, err
	}

	return articles, nil
}

func (p *Parser) parse(ctx context.Context, depth int64, url *url.URL) ([]*Article, error) {
	if depth < 0 {
		return nil, nil
	}

	articles := []*Article{}

	article, links, err := p.parsePage(ctx, url)
	if err != nil {
		return nil, err
	}
	articles = append(articles, article)

	for _, link := range links {
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
func (p *Parser) parsePage(ctx context.Context, url *url.URL) (*Article, []*url.URL, error) {
	return nil, nil, fmt.Errorf("not implemented")
}
