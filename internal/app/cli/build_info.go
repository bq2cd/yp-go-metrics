package cli

import (
	"fmt"
	"io"
)

// BuildInfo defines information about app's build.
type BuildInfo struct {
	Version string
	Date    string
	Commit  string
}

// PrintTo will print formatted build info to a provided [io.Writer].
func (bi BuildInfo) PrintTo(w io.Writer) {
	fmt.Fprintln(w, "Build version:", bi.formatValue(bi.Version))
	fmt.Fprintln(w, "Build date:", bi.formatValue(bi.Date))
	fmt.Fprintln(w, "Build commit:", bi.formatValue(bi.Commit))
}

func (bi BuildInfo) formatValue(v string) string {
	if v == "" {
		return "N/A"
	}

	return v
}
