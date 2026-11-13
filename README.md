# varstore

A sparse, queryable genotype store for cohort VCFs, and the command-line tool
that builds and queries it.

A store keeps only ALT-carrying genotypes plus enough context to tell a
confidently-called reference genotype (`0/0`) from a position that was never
assayed. Queries against a store and against the VCF it came from return the same
answer; the store is just faster to ask. gVCF reference blocks can be attached
per sample, so a store can also answer at positions no variant record mentions.

The format and its query API live in
[`cghts/varstore`](https://github.com/compgenlab/cghts); this repo is the CLI.
General VCF tooling is in [`cgkit`](https://github.com/compgenlab/cgkit).

## Building

```bash
make build     # bin/varstore.{darwin,linux}_{arm64,amd64}, windows_amd64.exe
make test
```

## Pipeline

```bash
# 1. Convert the joint-called pVCF, one volume per chromosome
varstore convert --out cohort/chr1 chr1.vcf.gz
varstore convert --out cohort/chr2 chr2.vcf.gz

# 2. Combine the volumes into one store
varstore create cohort/

# 3. (optional) Reference blocks from each sample's gVCF -- one job per sample
varstore refblock-calc --roster cohort --out frags/S0001 S0001.g.vcf.gz

# 4. (optional) Attach them; --replace S0001 later swaps in a redone sample
varstore refblock-attach cohort/ frags/

# 5. Query: sites, samples, or both
varstore query cohort --variant chr1:1000:A:T
varstore query cohort --variant panel.bed --sample S0001 --hom-ref
varstore query cohort.vcf.gz --variant chr1:1000:A:T   # same answer, from the VCF

# Describe a store or a VCF without reading genotypes
varstore summary cohort
```

Run `varstore <command> --help` for flags. `--profile=FILE` writes a CPU profile
from anywhere on the command line.

## Commands

| Command | Description |
|---------|-------------|
| `convert` | Convert one or more VCFs into a store volume ([format docs](docs/convert.md)) |
| `create` | Combine volumes, disjoint by chromosome, into one store |
| `refblock-calc` | Extract one sample's gVCF reference blocks into a fragment |
| `refblock-attach` | Merge fragments into a store's `ref_blocks` table |
| `query` | Query genotypes by site, by sample, or both, from a store or a VCF |
| `summary` | Describe a store or VCF: samples, contigs, provenance, census |

## Remote inputs

Inputs may be local paths, `http(s)://` URLs or `s3://` objects, and index
sidecars resolve over the same transport, so an indexed query still seeks.
Outputs are local, except that `convert --out s3://bucket/store` writes a store
straight to S3. S3 credentials come from the standard AWS chain;
`AWS_ENDPOINT_URL` points at an S3-compatible gateway.
