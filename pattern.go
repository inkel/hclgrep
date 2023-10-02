package hclgrep

import (
	"errors"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

type pat struct {
	ps  []string
	cur int
}

var ErrInvalidPattern = errors.New("invalid pattern")

func parsePat(p string) (*pat, error) {
	ps := strings.Split(p, ".")
	if len(ps) < 1 {
		return nil, ErrInvalidPattern
	}

	return &pat{
		ps: ps,
	}, nil
}

func (p pat) Match(b *hclsyntax.Block) (hcl.Range, bool) {
	if m := p.ps[p.cur]; m != "*" && m != b.Type {
		return hcl.Range{}, false
	}

	if len(p.ps) == p.cur+1 {
		return b.Range(), true
	}

	if len(b.Labels) > 0 {
		var found bool
		n := len(p.ps) - 1
		if l := len(b.Labels); l < n {
			n = l
		}

		for _, l := range b.Labels[:n] {
			p.cur++
			m := p.ps[p.cur]
			found = m == "*" || m == l
			if !found {
				break
			}
		}

		if !found {
			return hcl.Range{}, false
		}
	}

	p.cur++

	if len(p.ps) == p.cur {
		return b.Range(), true
	}

	if len(p.ps) == p.cur+1 { // attribute match
		if a, ok := b.Body.Attributes[p.ps[p.cur]]; ok {
			return a.Range(), true
		}
	}

	for _, ib := range b.Body.Blocks {
		if r, ok := p.Match(ib); ok {
			return r, true
		}
	}

	return hcl.Range{}, false
}
