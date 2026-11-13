package cli

import (
	"bytes"
	"testing"

	"github.com/compgenlab/varstore/internal/cmdtest"
	"github.com/spf13/cobra"
)

// runVcf executes a subcommand against a fresh root and returns its stdout.
// Command flags are bound to package globals, so they are reset to their
// defaults before each run to keep tests independent of ordering.
func runVcf(t *testing.T, args ...string) string {
	t.Helper()
	root, buf := vcfTestRoot(args...)
	if err := root.Execute(); err != nil {
		t.Fatalf("Execute(%v): %v", args, err)
	}
	return buf.String()
}

// runVcfErr runs a subcommand expecting an error, returning it.
func runVcfErr(t *testing.T, args ...string) error {
	t.Helper()
	root, _ := vcfTestRoot(args...)
	return root.Execute()
}

// vcfTestRoot builds a fresh root with all command flags reset to defaults.
func vcfTestRoot(args ...string) (*cobra.Command, *bytes.Buffer) {
	root := &cobra.Command{Use: "varstore"}
	InitCmd(root)
	// Slice flags need Replace rather than Set(DefValue); see cmdtest.ResetFlags.
	// That mattered here beyond tidiness: an extra bogus --sample is a harmless
	// no-op selector, but an extra bogus --region restricts the site axis and
	// silently filters every row away.
	cmdtest.ResetFlags(root)
	vcfVarQuerySamples = nil
	vcfVarQueryVariants = nil
	vcfToVarstoreForce = false
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	return root, &buf
}
