package main

import (
	"context"
	"errors"
	"flag"
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

// TODO remove this from global state
var list bool

func main() {
	flag.BoolVar(&list, "l", false, "only list filenames with results")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := realMain(ctx, flag.Args())
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

func find(ps []string, blocks hclsyntax.Blocks) hclsyntax.Blocks {
	var res hclsyntax.Blocks

	for _, b := range blocks {
		if ps[0] != "*" && ps[0] != b.Type {
			continue
		}

		if len(ps) < 2 {
			res = append(res, b)
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

		if len(ps) <= len(b.Labels)+1 {
			res = append(res, b)
			continue
		}

		pi = len(b.Labels) + 1

		for a := range b.Body.Attributes {
			if ps[pi] == a {
				res = append(res, b)
			}
		}

		if len(find(ps[pi:], b.Body.Blocks)) > 0 {
			res = append(res, b)
		}
	}

	return res
}

func grep(w io.Writer, ps []string, path string, src []byte) error {
	f, d := hclsyntax.ParseConfig(src, path, hcl.Pos{Line: 1, Column: 1})
	if d.HasErrors() {
		return fmt.Errorf("parsing %s: %s", path, d.Error())
	}

	blks := find(ps, f.Body.(*hclsyntax.Body).Blocks)

	if list && len(blks) > 0 {
		fmt.Fprintln(w, path)
		return nil
	}

	for _, b := range blks {
		printFound(w, b, src)
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
