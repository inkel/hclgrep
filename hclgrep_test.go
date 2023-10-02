package hclgrep

import (
	_ "embed"
	"testing"

	"github.com/hashicorp/hcl/v2"
)

//go:embed test.tf
var src []byte

func TestGrep(t *testing.T) {
	tests := map[string]struct {
		pat string
		exp []hcl.Pos // TODO this should be []hcl.Range
	}{
		"no matches": {"lorem", nil},

		"match block type - single": {
			"locals",
			[]hcl.Pos{
				{Line: 5, Column: 1},
			},
		},

		"match block and label": {
			"variable.foo",
			[]hcl.Pos{
				{Line: 14, Column: 1},
			},
		},

		"match block and labels": {
			"resource.null_resource.foo",
			[]hcl.Pos{
				{Line: 25, Column: 1},
			},
		},

		"match block and 1 label": {
			"resource.null_resource",
			[]hcl.Pos{
				{Line: 25, Column: 1},
				{Line: 27, Column: 1},
				{Line: 42, Column: 1},
			},
		},

		"match block, label & property": {
			"resource.null_resource.bar.count",
			[]hcl.Pos{
				{Line: 28, Column: 3},
			},
		},

		"match block property": {
			"resource.null_resource.*.provisioner",
			[]hcl.Pos{
				{Line: 34, Column: 3},
				{Line: 49, Column: 3},
			},
		},

		"match resource without label": {
			"terraform.required_providers",
			[]hcl.Pos{
				{Line: 72, Column: 3},
			},
		},
	}

	for n, tt := range tests {
		t.Run(n, func(t *testing.T) {
			got, err := Grep(tt.pat, src, "test.tf")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if e, g := len(tt.exp), len(got); e != g {
				t.Fatalf("expecting %d results, got %d", e, g)
			}

			for i, e := range tt.exp {
				// TODO not the best test, but it's a start
				if g := got[i].Start; e.Line != g.Line || e.Column != g.Column {
					t.Errorf("different result\nexp: %#v\ngot: %#v", e, g)
				}
			}
		})
	}
}
