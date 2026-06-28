package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// stringSliceFlag implements flag.Value to support repeatable flags.
type stringSliceFlag []string

func (s *stringSliceFlag) String() string { return strings.Join(*s, ",") }
func (s *stringSliceFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

type config struct {
	compartment    string
	profile        string
	excludeReports stringSliceFlag
}

// parseFlags builds a config from the given argument slice.
func parseFlags(args []string) (config, error) {
	fs := flag.NewFlagSet("mcp-sql-reports", flag.ContinueOnError)
	var cfg config
	fs.StringVar(&cfg.compartment, "compartment", "", "Default compartment OCID for list_reports")
	fs.StringVar(&cfg.profile, "profile", "DEFAULT", "OCI config profile name")
	fs.Var(&cfg.excludeReports, "exclude-report", "Exclude report OCID from results; repeatable")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	return cfg, nil
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "usage error: %v\n", err)
		os.Exit(1)
	}
	client, err := newOCIClient(cfg.profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create OCI client: %v\n", err)
		os.Exit(1)
	}
	serve(os.Stdin, os.Stdout, cfg, client)
}
