package converter

import (
	"github.com/gomarkdown/markdown/ast"
	"github.com/gomarkdown/markdown/parser"
)

type Spoiler struct {
	ast.Container
}

func registerSpoiler(p *parser.Parser) {
	prev := p.RegisterInline('|', nil)
	p.RegisterInline('|', spoilerInline(prev))
}

func spoilerInline(prev parser.InlineParser) parser.InlineParser {
	return func(p *parser.Parser, original []byte, offset int) (int, ast.Node) {
		data := original[offset:]
		if len(data) < 4 || data[0] != '|' || data[1] != '|' {
			if prev != nil {
				return prev(p, original, offset)
			}
			return 0, nil
		}

		for i := 2; i+1 < len(data); i++ {
			if data[i] == '\n' {
				return 0, nil
			}
			if data[i] == '|' && data[i+1] == '|' {
				if i == 2 {
					return 0, nil
				}
				spoiler := &Spoiler{}
				p.Inline(spoiler, data[2:i])
				return i + 2, spoiler
			}
		}

		return 0, nil
	}
}
