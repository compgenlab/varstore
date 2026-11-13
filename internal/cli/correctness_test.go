package cli

import (
	"strings"
	"testing"
)

// A command given no arguments must not write its help page into a redirected
// stdout and exit 0 -- "varstore query > out.tsv && next-step out.tsv" then
// proceeds with a help page as its input.
func TestNoArgsIsAUsageErrorNotHelpOnStdout(t *testing.T) {
	for _, cmd := range []string{
		"convert", "create", "refblock-calc", "refblock-attach", "query", "summary",
	} {
		root, buf := vcfTestRoot(cmd)
		err := root.Execute()
		if err == nil {
			t.Errorf("%s with no arguments returned nil; a pipeline would treat that as success", cmd)
		}
		if strings.Contains(buf.String(), "Usage:") && err == nil {
			t.Errorf("%s printed usage but reported success", cmd)
		}
	}
}
