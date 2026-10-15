# `cmd/analyze_habr`

A command-line diagnostic program that fetches the Habr homepage, parses its HTML, and prints a summary of stylesheet references, inline styles, element tag counts, and body text.

## Behavior

Running the command performs these steps:

1. Fetches `https://habr.com` using `internal/fetcher.Fetch`.
2. On fetch failure, writes `Fetch error: <error>` to standard error and exits with status 1. No parsing or report generation occurs.
3. Parses the returned HTML with `internal/dom.Parse`.
4. Prints external stylesheet links selected by `link[rel=stylesheet]`. Each `href` is displayed, truncated to 100 bytes with `...` appended when longer. A missing `href` is printed as an empty value.
5. Prints all `<style>` elements. Their text is truncated to 500 bytes with `...` appended when longer, and each is labeled with its zero-based index.
6. Recursively walks the parsed document tree and prints counts for each distinct element tag name. Tag names appear in the order in which they are first encountered during the depth-first traversal.
7. If at least one `<body>` is found, prints the first body's text. The displayed text is trimmed at its ends and truncated to 2,000 bytes with `...` appended if necessary. The reported character count is calculated from the untrimmed, untruncated body text.

All report output is written to standard output. Query results and document traversal follow the behavior of the internal DOM package.

## Public API

This is a `package main` executable; it exports no library API. Its entry point is the unexported `main()` function, invoked by `go run ./cmd/analyze_habr` or by building and running the command binary.

## Types and dependencies

The command defines no custom types. It uses:

- `internal/fetcher.Fetch` to retrieve the page.
- `internal/dom.Parse` to create the document representation.
- `dom.Node` fields `Type`, `TagName`, and `Children` to count elements recursively.
- `Node.QueryAll`, `Node.GetAttr`, and `Node.TextContent` for document queries and content extraction.

## Implementation details and limitations

- The page URL is hard-coded; there are no command-line flags or environment-variable options.
- The recursive walker counts every node whose type is `dom.ElementNode`, including nested elements.
- The `tags` map stores counts, while `allTags` separately preserves first-seen order for deterministic report ordering relative to the parsed tree.
- Truncation uses Go string byte lengths and byte slicing, not Unicode rune boundaries. A cut may therefore split a multibyte UTF-8 character.
- Errors from parsing and output operations are not handled explicitly by this command.
