package cli

import (
	"fmt"
	"io"
	"iter"
	"math"

	"github.com/compgenlab/cghts/htsio"
	"github.com/compgenlab/cghts/vcf"
	"github.com/compgenlab/varstore/internal/buildinfo"
	"github.com/compgenlab/varstore/internal/cmdio"
	"github.com/spf13/cobra"
)

// stampVcfProvenance updates the header's ##fileDate and appends varstore command
// and version provenance lines, recording how the output file was produced. It
// is called by any command that writes a VCF file.
func stampVcfProvenance(h *vcf.VcfHeader, cmdName string) {
	h.SetFileDate(buildinfo.Date())
	h.AddLine("##varstore_" + cmdName + "Command=" + buildinfo.CommandLine())
	h.AddLine("##varstore_" + cmdName + "Version=" + buildinfo.String())
}

// openVcfInput opens a streaming VCF reader for filename, reading from stdin
// when filename is "-". Input is transparently gunzipped, and may be a local
// path or a remote locator (http(s)://, s3://).
//
// Note that a streaming read of a remote object transfers the whole thing;
// there is no index to skip with. Commands taking --region seek instead.
func openVcfInput(cmd *cobra.Command, filename string) (*vcf.VcfReader, error) {
	if filename == "-" {
		return vcf.NewVcfReader(cmd.InOrStdin())
	}
	r, err := vcf.OpenVcfFile(cmd.Context(), filename)
	if err != nil {
		// The locator is worth repeating: for a remote read the underlying
		// error is an HTTP status or an SDK message with no path in it.
		return nil, fmt.Errorf("opening %s: %w", filename, err)
	}
	return r, nil
}

// openOutput returns the writer for output, using stdout when output is "" or
// "-". The returned closer is nil when writing to stdout.
func openOutput(cmd *cobra.Command, output string) (io.Writer, func() error, error) {
	out, err := cmdio.CreateOutput(cmd, "-o/--output", output)
	if err != nil {
		return nil, nil, err
	}
	if output == "" || output == "-" {
		return out.W, nil, nil
	}
	return out.W, out.Close, nil
}

// recordSource is a uniform record stream over either a plain (streaming) VCF or
// a tabix-indexed region query. next returns io.EOF when exhausted.
type recordSource struct {
	header *vcf.VcfHeader
	next   func() (*vcf.VcfRecord, error)
	close  func() error
}

// openRecordSource opens filename for reading. When region is non-empty it must
// name a tabix-indexed file (not stdin); records are limited to that region
// (1-based inclusive "chrom:start-end", or bare "chrom"). Otherwise the file is
// read as a stream.
func openRecordSource(cmd *cobra.Command, filename, region string) (*recordSource, error) {
	if region != "" {
		if filename == "-" {
			return nil, fmt.Errorf("--region requires an indexed VCF file, not stdin")
		}
		ref, start, end, err := htsio.ParseRegion(region)
		if err != nil {
			return nil, err
		}
		if end < 0 {
			end = math.MaxInt32
		}
		ir, err := vcf.OpenIndexedVcfReader(cmd.Context(), filename)
		if err != nil {
			return nil, err
		}
		header, err := ir.Header()
		if err != nil {
			ir.Close()
			return nil, err
		}
		seq, err := ir.Query(ref, start, end)
		if err != nil {
			ir.Close()
			return nil, err
		}
		next, stop := iter.Pull2(seq)
		return &recordSource{
			header: header,
			next: func() (*vcf.VcfRecord, error) {
				rec, qerr, ok := next()
				if !ok {
					return nil, io.EOF
				}
				if qerr != nil {
					return nil, qerr
				}
				return rec, nil
			},
			close: func() error {
				stop()
				return ir.Close()
			},
		}, nil
	}

	reader, err := openVcfInput(cmd, filename)
	if err != nil {
		return nil, err
	}
	header, err := reader.Header()
	if err != nil {
		reader.Close()
		return nil, err
	}
	return &recordSource{
		header: header,
		next:   reader.NextRecord,
		close: func() error {
			reader.Close()
			return nil
		},
	}, nil
}

// Flags shared by many commands, backed by one variable each.
//
// This follows vcfTbi: only one subcommand runs per process, so a variable per
// command bought nothing and cost something. --passing had fourteen of them and
// --region five, each registered with its own hand-written help string, and
// every one had to be added by hand to the reset list in vcf_commands_test.go
// or it leaked into the next test.
var (
	// vcfPassing backs --passing. The phrasing differs per command -- "output",
	// "export", "convert" -- so the help text stays a caller's argument even
	// though the flag does not.
	vcfPassing bool

	// vcfRegion backs --region. All five declarations were character-identical.
	vcfRegion string

	// vcfVerbose backs -v/--verbose. What gets reported differs per command, so
	// the help text is still the caller's.
	vcfVerbose bool
)

// addVerboseFlag registers -v/--verbose with command-specific help.
func addVerboseFlag(cmd *cobra.Command, help string) {
	cmd.Flags().BoolVarP(&vcfVerbose, "verbose", "v", false, help)
}

// addPassingFlag registers --passing with command-specific help.
func addPassingFlag(cmd *cobra.Command, help string) {
	cmd.Flags().BoolVar(&vcfPassing, "passing", false, help)
}

// addRegionFlag registers --region, which needs a tabix-indexed input.
func addRegionFlag(cmd *cobra.Command) {
	cmd.Flags().StringVar(&vcfRegion, "region", "",
		"Only variants in this 1-based region (chrom:start-end, or chrom); requires a tabix-indexed file")
}
