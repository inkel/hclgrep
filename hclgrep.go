package hclgrep

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func Grep(pattern string, src []byte, file string) ([]hcl.Range, error) {
	f, d := hclsyntax.ParseConfig(src, file, hcl.Pos{Line: 1, Column: 1})
	if d.HasErrors() {
		return nil, fmt.Errorf("parsing %s: %s", file, d.Error())
	}

	pat, err := parsePat(pattern)
	if err != nil {
		return nil, err
	}

	var res []hcl.Range

	for _, b := range f.Body.(*hclsyntax.Body).Blocks {
		if r, ok := pat.Match(b); ok {
			res = append(res, r)
			continue
		}
	}

	return res, nil
}
