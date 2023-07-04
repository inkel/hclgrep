# Grep your HCL files
`hclgrep` is a `grep(1)` like utility that allows to pass a very simple pattern and will output the HCL block where that pattern is found.

It currently only accepts a pattern for checking if a block, label, or attribute is found.
Future versions might include matching for attribute values.

## Installation
If you have Go set up in your machine, just run:

```bash
go install github.com/inkel/hclgrep@latest
```

## Usage
```
hclgrep <pattern> [path]
```

## License
MIT. See [LICENSE](LICENSE).
