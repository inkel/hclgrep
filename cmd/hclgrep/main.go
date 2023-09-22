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

	"github.com/hashicorp/hcl/v2"
	"github.com/inkel/hclgrep"
)

func main() {
	var verbose bool

	flag.BoolVar(&verbose, "v", false, "display Terraform body")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	err := realMain(ctx, verbose, flag.Args())
	cancel()

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func realMain(ctx context.Context, verbose bool, args []string) error {
	if len(args) == 0 {
		return errors.New("not enough arguments")
	}

	pattern, paths := args[0], args[1:]

	if len(paths) == 0 { // read stdin
		src, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		res, err := hclgrep.Grep(pattern, src, "-")
		if err != nil {
			return err
		}

		print(os.Stdout, "-", src, res, verbose)

		return nil
	}

	var files []string

	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			return err
		}

		if fi.IsDir() {
			if err := fs.WalkDir(os.DirFS(p), ".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				if filepath.Ext(path) == ".tf" {
					files = append(files, filepath.Join(p, path))
				}
				return nil
			}); err != nil {
				return fmt.Errorf("walking: %w", err)
			}

			continue
		}

		if filepath.Ext(p) == ".tf" {
			files = append(files, p)
		}
	}

	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		res, err := hclgrep.Grep(pattern, src, f)
		if err != nil {
			return err
		}

		print(os.Stdout, f, src, res, verbose)
	}

	return nil
}

var lf = []byte{'\n'}

func print(w io.Writer, path string, src []byte, res []hcl.Range, verbose bool) {
	for _, r := range res {
		if verbose {
			fmt.Fprintf(w, "%s @ %d,%d-%d,%d\n", path, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
			w.Write(r.SliceBytes(src)) //nolint:errcheck
			fmt.Fprintf(w, "\n\n")
		} else {
			fmt.Fprintf(w, "%s:%d,%d-%d,%d\n", path, r.Start.Line, r.Start.Column, r.End.Line, r.End.Column)
		}
	}
}
