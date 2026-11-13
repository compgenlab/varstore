package cli

import (
	"io"
	"os"
	"testing"

	"github.com/compgenlab/cghts/htsio/bgzf"
	"github.com/compgenlab/cghts/htsio/tabix"
)

// bgzipIndex writes src as a bgzipped VCF at dst with a .tbi beside it, for the
// tests that need an indexed VCF as a query backend. It is a fixture step, not
// something varstore offers: this binary creates and queries stores, and leaves
// VCF manipulation to cgkit.
func bgzipIndex(tb testing.TB, src, dst string) {
	tb.Helper()
	in, err := os.Open(src)
	if err != nil {
		tb.Fatal(err)
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		tb.Fatal(err)
	}
	w := bgzf.NewWriter(out)
	if _, err := io.Copy(w, in); err != nil {
		tb.Fatal(err)
	}
	if err := w.Close(); err != nil {
		tb.Fatal(err)
	}
	if err := out.Close(); err != nil {
		tb.Fatal(err)
	}
	if err := tabix.NewIndexWriter(tabix.NewWriterOpts().VCF()).WriteIndex(dst); err != nil {
		tb.Fatal(err)
	}
}
