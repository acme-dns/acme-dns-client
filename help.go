package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

var (
	usageExamples = map[string]string{
		"check": `
EXAMPLE USAGE:
  Check the configuration for all domains configured in the system:
    acme-dns-client check

  Check the configuration for two domains; example.org and test.example.org:
    acme-dns-client check -d 'example.org,test.example.org'
`,
		"register": `
EXAMPLE USAGE:
  Register a new acme-dns account for domain example.org, using acme-dns instance at acmedns.example.org:
    acme-dns-client register -d example.org -s auth.acmedns.example.org
  
  Register a new acme-dns account for domain example.org, allow updates only from 198.51.100.0/24:
    acme-dns-client register -d example.org -allow 198.51.100.0/24
`}
)

type UsageFlag struct {
	Name        string
	Description string
	Default     string
}

//PrintFlag prints out the flag name, usage string and default value
func (f *UsageFlag) PrintFlag(max_length int) {
	// Create format string, used for padding
	format := fmt.Sprintf("   -%%-%ds %%s", max_length)
	if f.Default != "" {
		format = format + " (default: %s)\n"
		fmt.Printf(format, f.Name, f.Description, f.Default)
	} else {
		format = format + "\n"
		fmt.Printf(format, f.Name, f.Description)
	}
}

func UsageCommandHeader(command string) {
	fmt.Printf(`acme-dns-client - v%s

Usage:	%s %s [OPTIONS]
`, VERSION, filepath.Base(os.Args[0]), command)
}

func UsageGeneric() {
	fmt.Printf(`acme-dns-client - v%s

Usage:	%s COMMAND [OPTIONS]

Commands:
  register		Register a new acme-dns account for a domain
  check			Check the configuration and settings of existing acme-dns accounts
  list			List all the existing acme-dns accounts and perform simple CNAME checks for them

Options:
  --help		Print this help text

To get help for specific command, use:
  %s COMMAND --help
`, VERSION, filepath.Base(os.Args[0]), filepath.Base(os.Args[0]))

	// Usage examples.
	fmt.Printf(`
EXAMPLE USAGE:
  Register a new acme-dns account for domain example.org:
    acme-dns-client register -d example.org
  
  Register a new acme-dns account for domain example.org, allow updates only from 198.51.100.0/24:
    acme-dns-client register -d example.org -allow 198.51.100.0/24

  Check the configuration of example.org and the corresponding acme-dns account:
    acme-dns-client check -d example.org

  Check the configuration of all the domains and acme-dns accounts registered on this machine:
    acme-dns-client check

  Print help for a "register" command:
    acme-dns-client register --help

`)
}

func FSUsage(fset *flag.FlagSet) func() {
	return func() {
		UsageCommandHeader(fset.Name())
		fmt.Printf("\nOptions for %s:\n", fset.Name())
		flags := make([]UsageFlag, 0)
		max_length := 0
		fset.VisitAll(func(f *flag.Flag) {
			if f.Name == "dangerous" {
				return
			}
			flags = append(flags, UsageFlag{
				Name:        f.Name,
				Description: f.Usage,
				Default:     f.DefValue,
			})
			if len(f.Name) > max_length {
				max_length = len(f.Name)
			}
		})

		// Print out the flag info
		for _, f := range flags {
			f.PrintFlag(max_length)
		}
		fmt.Printf(usageExamples[fset.Name()])
		fmt.Printf("\n")
	}
}
