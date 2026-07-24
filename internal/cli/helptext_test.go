package cli

import (
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

// helpFlagSet builds a fresh flag set with all CLI flags registered, so
// tests can enumerate registrations without running the parser.
func helpFlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("gematria", pflag.ContinueOnError)
	registerFlags(fs, &cliFlags{})
	return fs
}

// Every registered flag must be documented in helpText: the long name as
// "--name", and the shorthand (when present) as "-x," matching the
// "-m, --mispar" layout.
func TestHelpText_documentsAllRegisteredFlags(t *testing.T) {
	helpFlagSet().VisitAll(func(f *pflag.Flag) {
		if !strings.Contains(helpText, "--"+f.Name) {
			t.Errorf("helpText missing registered flag --%s", f.Name)
		}
		if f.Shorthand != "" && !strings.Contains(helpText, "-"+f.Shorthand+",") {
			t.Errorf("helpText missing shorthand -%s, for flag --%s", f.Shorthand, f.Name)
		}
	})
}

// Every "--flag" token mentioned anywhere in helpText must be a registered
// flag, so renamed or removed flags cannot leave stale help entries behind.
func TestHelpText_mentionsOnlyRegisteredFlags(t *testing.T) {
	fs := helpFlagSet()
	longFlag := regexp.MustCompile(`--[a-z][a-z0-9-]*`)
	for _, token := range longFlag.FindAllString(helpText, -1) {
		name := strings.TrimPrefix(token, "--")
		if fs.Lookup(name) == nil {
			t.Errorf("helpText mentions %s, which is not a registered flag", token)
		}
	}
}
