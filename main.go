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
	authType       string
	excludeReports stringSliceFlag
}

// parseFlags builds a config from the given argument slice.
func parseFlags(args []string) (config, error) {
	fs := flag.NewFlagSet("mcp-sql-reports", flag.ContinueOnError)
	var cfg config
	fs.StringVar(&cfg.compartment, "compartment", "", "Default compartment OCID for list_reports")
	fs.StringVar(&cfg.profile, "profile", "DEFAULT", "OCI config profile name")
	fs.StringVar(&cfg.authType, "auth", "api_key", "Authentication type: api_key, instance_principal, or security_token")
	fs.Var(&cfg.excludeReports, "exclude-report", "Exclude report OCID from results; repeatable")
	if err := fs.Parse(args); err != nil {
		return config{}, err
	}
	switch cfg.authType {
	case "api_key", "instance_principal", "security_token":
	default:
		return config{}, fmt.Errorf("invalid -auth value %q: must be api_key, instance_principal, or security_token", cfg.authType)
	}
	return cfg, nil
}

func main() {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "usage error: %v\n", err)
		os.Exit(1)
	}
	client, err := newOCIClient(cfg.authType, cfg.profile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create OCI client: %v\n", err)
		os.Exit(1)
	}
	serve(os.Stdin, os.Stdout, cfg, client)
}
