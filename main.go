package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := realMain(ctx, os.Args[1:])
	cancel()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return errors.New("not enough arguments")
	}

	pattern, paths := args[0], args[1:]

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
						printFound(b, src)
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
						printFound(b, src)
						continue
					}

					for a := range b.Body.Attributes {
						if ps[pi] == a {
							printFound(b, src)
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

var styleInfo = lipgloss.NewStyle().Bold(true)

func printFound(b *hclsyntax.Block, src []byte) {
	r := b.Range()
	var s strings.Builder

	s.WriteString(styleInfo.Render(fmt.Sprintf("%s:%d:%d:", r.Filename, r.Start.Line, r.Start.Column)))
	s.WriteRune('\n')
	bs := r.SliceBytes(src)
	s.Write(bs)

	fmt.Println(s.String())
}
