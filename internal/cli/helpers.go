package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/DisruptiveWorks/archipulse/internal/cliconfig"
)

// globalFlags are the connection flags shared across all resource commands.
type globalFlags struct {
	server  string
	context string
	output  string // "table" | "json"
}

// addGlobalFlags registers --server, --context, and --output on fs.
func addGlobalFlags(fs *flag.FlagSet, f *globalFlags) {
	fs.StringVar(&f.server, "server", "", "Override server URL")
	fs.StringVar(&f.context, "context", "", "Override active context")
	fs.StringVar(&f.output, "output", "table", "Output format: table or json")
}

// resolveClient resolves connection params and returns a ready Client.
// Returns an error if no token is available (user needs to run login).
func resolveClient(f globalFlags) (*Client, error) {
	r := cliconfig.Resolve(cliconfig.ResolveOptions{
		Server:  f.server,
		Context: f.context,
	})
	if r.Token == "" {
		return nil, fmt.Errorf("not authenticated — run 'archipulse login' first")
	}
	return NewClient(r.Server, r.Token), nil
}

// stdoutWriter returns os.Stdout as an io.Writer — used for JSON encoding.
func stdoutWriter() *os.File { return os.Stdout }

// table prints rows as a left-aligned fixed-width table with a header.
// headers and each row must have the same number of columns.
func table(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for i, h := range headers {
		widths[i] = len(h)
	}
	for _, row := range rows {
		for i, cell := range row {
			if i < len(widths) && len(cell) > widths[i] {
				widths[i] = len(cell)
			}
		}
	}

	printRow := func(cols []string) {
		var sb strings.Builder
		for i, col := range cols {
			if i > 0 {
				sb.WriteString("  ")
			}
			sb.WriteString(col)
			if i < len(widths)-1 {
				sb.WriteString(strings.Repeat(" ", widths[i]-len(col)))
			}
		}
		fmt.Println(sb.String())
	}

	printRow(headers)

	sep := make([]string, len(headers))
	for i, w := range widths {
		sep[i] = strings.Repeat("-", w)
	}
	printRow(sep)

	for _, row := range rows {
		printRow(row)
	}
}
