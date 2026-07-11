package glob

import "strings"

func Match(glob string, path string) bool {
	return match(NewParser(glob).parse(), path)
}

type Glob []globNode

type globNode interface {
	globNode()
}

func (glob Glob) Match(path string) bool {
	return match(glob, path)
}

func NewGlob(pattern string) Glob {
	return NewParser(pattern).parse()
}

type literal string

func (l literal) globNode() {}

type wildcardSingle struct{}

func (w wildcardSingle) globNode() {}

type wildcardDouble struct{}

func (w wildcardDouble) globNode() {}

type wildcard struct{}

func (w wildcard) globNode() {}

type Parser struct {
	src      string
	position int
}

func NewParser(value string) *Parser {
	return &Parser{src: value}
}

const stateStart = 0
const stateLiteral = 1

func (parser *Parser) parse() Glob {
	state := stateStart
	var sb strings.Builder

	for !parser.AtEnd() {
		switch {
		case state == stateStart && parser.Skip("**"):
			{
				return append(
					Glob{wildcardDouble{}},
					parser.parse()...,
				)
			}
		case state == stateStart && parser.Skip("*"):
			{
				return append(
					Glob{wildcard{}},
					parser.parse()...,
				)
			}
		case state == stateStart && parser.Skip("?"):
			{
				return append(
					Glob{wildcardSingle{}},
					parser.parse()...,
				)
			}
		case state == stateLiteral && (parser.Read("*") || parser.Read("?")):
			{
				return append(
					Glob{literal(sb.String())},
					parser.parse()...,
				)
			}
		default:
			sb.WriteByte(parser.Peek())
			parser.Next()
			state = stateLiteral
		}
	}
	if sb.Len() > 0 {
		return Glob{literal(sb.String())}
	}
	return Glob{}
}

func (parser *Parser) AtEnd() bool {
	return parser.position >= len(parser.src)
}

func (parser *Parser) Next() {
	parser.position++
}

func (parser *Parser) Skip(value string) bool {
	if parser.Read(value) {
		parser.position = parser.position + len(value)
		return true
	}
	return false
}

func (parser *Parser) IsWhitespace() bool {
	c := parser.Peek()
	return c == ' ' || c == '\t'
}

func (parser *Parser) Rest() string {
	return parser.src[parser.position:]
}

func (parser *Parser) Read(prefix string) bool {
	if len(prefix) < 1 {
		return true
	} else if strings.HasPrefix(parser.Rest(), prefix) {
		parser.position += len(prefix) - 1
		return true
	} else {
		return false
	}
}

func (parser *Parser) Peek() byte {
	if parser.position >= len(parser.src) {
		return 0
	}
	return parser.src[parser.position]
}

func match(pattern Glob, path string) bool {
	if len(pattern) == 0 {
		return len(path) == 0
	}
	first, rest := pattern[0], pattern[1:]
	switch node := first.(type) {
	case literal:
		{
			if strings.HasPrefix(path, string(node)) {
				return match(rest, path[len(node):])
			} else {
				return false
			}
		}
	case wildcard:
		{
			for i := 0; i < len(path); i++ {
				if path[i] == '/' {
					return match(rest, path[i:])
				}
				if match(rest, path[i:]) {
					return true
				}
			}
			return match(rest, "")
		}
	case wildcardDouble:
		{
			for i := 0; i < len(path); i++ {
				if match(rest, path[i:]) {
					return true
				}
			}
			return match(rest, "")
		}
	case wildcardSingle:
		{
			if len(path) == 0 {
				return false
			}
			return match(rest, path[1:])
		}
	default:
		return false
	}

}
