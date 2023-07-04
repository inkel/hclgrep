package main

import (
	"context"
	"errors"
	"fmt"
	"io"
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
	ps := strings.Split(pattern, ".")

	if len(paths) == 0 { // read stdin
		src, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		return grep(os.Stdout, ps, "-", src)
	}

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

				return grep(os.Stdout, ps, path, src)
			}

			return nil
		}
	}

	for _, path := range paths {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if err := fs.WalkDir(os.DirFS(path), ".", walkFn(path)); err != nil {
				return fmt.Errorf("walking %s: %w", path, err)
			}
		} else {
			src, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			if err := grep(os.Stdout, ps, path, src); err != nil {
				return err
			}
		}
	}

	return nil
}

func grep(w io.Writer, ps []string, path string, src []byte) error {
	f, d := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
	if d.HasErrors() {
		return fmt.Errorf("parsing %s: %s", path, d.Error())
	}

	for _, b := range f.Body.(*hclsyntax.Body).Blocks {
		if ps[0] != "*" && ps[0] != b.Type {
			continue
		}

		if len(ps) < 2 {
			printFound(w, b, src)
			continue
		}

		var pi int
		var found bool

		for _, l := range b.Labels {
			pi++
			found = ps[pi] == "*" || ps[pi] == l
			if found {
				break
			}
			if len(ps) > pi {
				break
			}
		}

		if !found {
			continue
		}

		pi++
		if pi == len(b.Labels) {
			printFound(w, b, src)
			continue
		}

		for a := range b.Body.Attributes {
			if ps[pi] == a {
				printFound(w, b, src)
				break
			}
		}
	}

	return nil
}

var styleInfo = lipgloss.NewStyle().Bold(true)

func printFound(w io.Writer, b *hclsyntax.Block, src []byte) {
	r := b.Range()

	fmt.Fprintln(w, styleInfo.Render(fmt.Sprintf("%s:%d:%d:", r.Filename, r.Start.Line, r.Start.Column)))

	bs := r.SliceBytes(src)
	w.Write(bs)
	w.Write([]byte{'\n'})
}
