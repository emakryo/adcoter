package util

import (
	"strings"
	"golang.org/x/net/html"
)

type Condition interface {
	IsMet(*html.Node) bool
}

type ElementCondition struct {
	data []string
}

func NewElemCond(data ...string) *ElementCondition {
	return &ElementCondition{data}
}

func (c *ElementCondition) IsMet(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}
	for _, d := range c.data {
		if node.Data == d {
			return true
		}
	}
	return false
}

type Table = [][]string

func FindAll(node *html.Node, cond Condition) (nodes []*html.Node) {
	if cond.IsMet(node) {
		nodes = append(nodes, node)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		nodes = append(nodes, FindAll(c, cond)...)
	}
	return nodes;
}

func FindFirst(node *html.Node, cond Condition) *html.Node {
	if cond.IsMet(node) {
		return node
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		m := FindFirst(c, cond);
		if m != nil {
			return m
		}
	}
	return nil;
}

func InnerText(node *html.Node) string {
	if node.Type == html.TextNode && node.FirstChild == nil {
		return strings.TrimSpace(node.Data)
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		s := InnerText(c)
		if s != "" {
			return s
		}
	}

	return ""
}

func ParseTable(node *html.Node) (tab Table) {
	rows := FindAll(node, &ElementCondition{[]string{"tr"}})
	tab = make([][]string, len(rows))
	for i, r := range rows {
		cols := FindAll(r, &ElementCondition{[]string{"th", "td"}})
		tab[i] = make([]string, len(cols))
		for j, c := range cols {
			tab[i][j] = InnerText(c)
		}
	}
	return
}
