package parser

import (
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
	URL         url.URL
	PublishedAt time.Time
	Title       string
	Subtitle    string
	Summary     string
}

// Parse starts parsing.
func (p *Parser) Parse(depth int64, concurrency int64) ([]*Article, error) {
	return nil, fmt.Errorf("not implemented")
}
