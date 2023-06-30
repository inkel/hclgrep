package main

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
	var (
		pattern = "module.*.cluster_name"
		path    = "/Users/inkel/dev/grafana/repos/deployment_tools/terraform/clusters/prod-us-west-0"
	)

	pattern = os.Args[1]
	path = os.Args[2]

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := realMain(ctx, pattern, path)
	cancel()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, pattern string, paths ...string) error {
	walkFn := func(root string) fs.WalkDirFunc {
		return func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if d.IsDir() {
				return nil
			}

			if filepath.Ext(path) == ".tf" {
				path := filepath.Join(root, path)
				src, err := os.ReadFile(path)
				if err != nil {
					return err
				}

				f, d := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
				if d.HasErrors() {
					return fmt.Errorf("parsing %s: %s", path, d.Error())
				}

				ps := strings.Split(pattern, ".")

				for _, b := range f.Body.(*hclsyntax.Body).Blocks {
					if ps[0] != "*" && ps[0] != b.Type {
						continue
					}

					if len(ps) < 2 {
						printFound(b)
						continue
					}

					var pi int
					var found bool

					for _, l := range b.Labels {
						pi++
						found = ps[pi] == "*" || ps[pi] == l
						if !found {
							break
						}
					}

					if !found {
						continue
					}

					pi++

					if pi == len(ps) {
						printFound(b)
						continue
					}

					for a := range b.Body.Attributes {
						if ps[pi] == a {
							printFound(b)
							break
						}
					}
				}
			}

			return nil
		}
	}

	for _, path := range paths {
		if err := fs.WalkDir(os.DirFS(path), ".", walkFn(path)); err != nil {
			return fmt.Errorf("walking %s: %w", path, err)
		}
	}

	return nil
}

func printFound(b *hclsyntax.Block) {
	r := b.Range()
	var s strings.Builder

	fmt.Fprintf(&s, "%s:%d:%d: %s", r.Filename, r.Start.Line, r.Start.Column, b.Type)

	for _, l := range b.Labels {
		fmt.Fprintf(&s, " %q", l)
	}

	fmt.Println(s.String())

	// blk := b.AsHCLBlock()
	// fmt.Printf("%T\n", blk)
	// _, ok := blk.Body.(*hclwrite.Body)
	// fmt.Println(ok)
}
