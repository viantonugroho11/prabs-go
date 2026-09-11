package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/prabs/prabs-go/internal/analyzer"
	"github.com/prabs/prabs-go/internal/config"
	"github.com/prabs/prabs-go/internal/mutation"
	"github.com/prabs/prabs-go/internal/output"
	"github.com/prabs/prabs-go/internal/rules"
)

const Version = "0.1.0"

const usage = `prabs-go — Go code smell analyzer & mutation testing

Usage:
  prabs-go scan [paths...] [flags]
  prabs-go mutate [path] [flags]
  prabs-go verify [path] [flags]
  prabs-go rules
  prabs-go version

Flags:
  --format text|json|sarif
  --severity info|low|medium|high|critical  (min severity)
  --rule RULE-ID                             (restrict to one rule / one mutator)
  --config path                              (default: .prabs.yaml)
  --fail-on info|low|medium|high|critical
  --verbose
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	switch cmd {
	case "scan":
		os.Exit(cmdScan(args))
	case "mutate":
		os.Exit(cmdMutate(args))
	case "verify":
		os.Exit(cmdVerify(args))
	case "rules":
		os.Exit(cmdRules())
	case "version", "--version", "-v":
		fmt.Println("prabs-go", Version)
		os.Exit(0)
	case "help", "--help", "-h":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n%s", cmd, usage)
		os.Exit(2)
	}
}

type commonFlags struct {
	format   string
	severity string
	rule     string
	config   string
	failOn   string
	verbose  bool
}

func parseCommon(fs *flag.FlagSet, args []string) (*commonFlags, []string, error) {
	c := &commonFlags{}
	fs.StringVar(&c.format, "format", "text", "output format: text|json|sarif")
	fs.StringVar(&c.severity, "severity", "", "minimum severity to include")
	fs.StringVar(&c.rule, "rule", "", "restrict to a single rule id / mutator target")
	fs.StringVar(&c.config, "config", ".prabs.yaml", "path to config")
	fs.StringVar(&c.failOn, "fail-on", "", "severity that causes non-zero exit")
	fs.BoolVar(&c.verbose, "verbose", false, "verbose output")
	flags, positional := splitFlags(args, fs)
	if err := fs.Parse(flags); err != nil {
		return nil, nil, err
	}
	return c, append(positional, fs.Args()...), nil
}

// splitFlags reorders args so all recognized flag tokens come first,
// letting positional arguments appear anywhere on the command line.
func splitFlags(args []string, fs *flag.FlagSet) ([]string, []string) {
	known := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) { known[f.Name] = true })
	var flags, positional []string
	i := 0
	for i < len(args) {
		a := args[i]
		if len(a) > 1 && a[0] == '-' {
			name := a
			for len(name) > 0 && name[0] == '-' {
				name = name[1:]
			}
			eq := -1
			for k := 0; k < len(name); k++ {
				if name[k] == '=' {
					eq = k
					break
				}
			}
			key := name
			if eq >= 0 {
				key = name[:eq]
			}
			if known[key] {
				flags = append(flags, a)
				if eq < 0 && i+1 < len(args) && !boolFlag(fs, key) {
					flags = append(flags, args[i+1])
					i += 2
					continue
				}
				i++
				continue
			}
		}
		positional = append(positional, a)
		i++
	}
	return flags, positional
}

func boolFlag(fs *flag.FlagSet, name string) bool {
	f := fs.Lookup(name)
	if f == nil {
		return false
	}
	if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok {
		return bf.IsBoolFlag()
	}
	return false
}

func loadCfg(path string) (*config.Config, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return config.Default(), nil
	}
	return config.Load(path)
}

func cmdScan(args []string) int {
	fs := flag.NewFlagSet("scan", flag.ContinueOnError)
	c, positional, err := parseCommon(fs, args)
	if err != nil {
		return 2
	}
	cfg, err := loadCfg(c.config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 2
	}
	roots := positional
	if len(roots) == 0 {
		roots = []string{"."}
	}
	reg := analyzer.NewRegistry()
	rules.RegisterAll(reg)

	opts := analyzer.Options{
		Roots:    roots,
		Excludes: cfg.Exclude,
		Rules:    cfg.Rules,
		Verbose:  c.verbose,
	}
	if c.severity != "" {
		if sv, ok := analyzer.ParseSeverity(c.severity); ok {
			opts.SeverityFilter = sv
		}
	}
	if c.rule != "" {
		opts.OnlyRules = map[string]bool{c.rule: true}
	}
	a := analyzer.New(reg)
	res, err := a.Run(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "analyze:", err)
		return 3
	}
	if err := output.Write(os.Stdout, output.Format(c.format), res); err != nil {
		fmt.Fprintln(os.Stderr, "output:", err)
		return 3
	}
	failOn := cfg.FailOn
	if c.failOn != "" {
		if sv, ok := analyzer.ParseSeverity(c.failOn); ok {
			failOn = sv
		}
	}
	for _, f := range res.Findings {
		if f.Severity.Rank() >= failOn.Rank() {
			return 1
		}
	}
	return 0
}

func cmdRules() int {
	for _, r := range rules.All() {
		fmt.Printf("%-16s [%s] %s\n", r.ID(), r.Severity(), r.Name())
		fmt.Printf("                  %s\n", r.Description())
	}
	return 0
}

func cmdMutate(args []string) int {
	fs := flag.NewFlagSet("mutate", flag.ContinueOnError)
	c, positional, err := parseCommon(fs, args)
	if err != nil {
		return 2
	}
	src := "."
	if len(positional) > 0 {
		src = positional[0]
	}
	abs, _ := filepath.Abs(src)
	sb, err := mutation.NewSandbox(abs)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sandbox:", err)
		return 3
	}
	defer sb.Cleanup()
	fmt.Println("Sandbox:", sb.Dir)
	mctx := &mutation.MutationContext{Dir: sb.Dir}
	for _, m := range mutation.All() {
		if c.rule != "" && m.TargetRule() != c.rule {
			continue
		}
		if !m.CanMutate(mctx) {
			continue
		}
		res, err := m.Mutate(mctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, m.ID(), "failed:", err)
			continue
		}
		fmt.Printf("  applied %s (target %s) -> %s\n", res.MutatorID, res.TargetRule, strings.TrimPrefix(res.File, sb.Dir+string(filepath.Separator)))
	}
	fmt.Println("Sandbox retained for inspection? no — cleaned up on exit.")
	return 0
}

func cmdVerify(args []string) int {
	fs := flag.NewFlagSet("verify", flag.ContinueOnError)
	c, positional, err := parseCommon(fs, args)
	if err != nil {
		return 2
	}
	src := "."
	if len(positional) > 0 {
		src = positional[0]
	}
	abs, _ := filepath.Abs(src)
	cfg, err := loadCfg(c.config)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 2
	}
	reg := analyzer.NewRegistry()
	rules.RegisterAll(reg)

	mutators := mutation.All()
	if c.rule != "" {
		filtered := mutators[:0]
		for _, m := range mutators {
			if m.TargetRule() == c.rule {
				filtered = append(filtered, m)
			}
		}
		mutators = filtered
	}

	fmt.Println("PRABS Mutation Verification")
	fmt.Println()
	detected, total := 0, 0
	for _, m := range mutators {
		total++
		sb, err := mutation.NewSandbox(abs)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sandbox:", err)
			continue
		}
		mctx := &mutation.MutationContext{Dir: sb.Dir}
		mres, err := m.Mutate(mctx)
		if err != nil {
			fmt.Fprintln(os.Stderr, m.ID(), "mutate failed:", err)
			sb.Cleanup()
			continue
		}
		a := analyzer.New(reg)
		opts := analyzer.Options{
			Roots:     []string{sb.Dir},
			Excludes:  cfg.Exclude,
			Rules:     cfg.Rules,
			OnlyRules: map[string]bool{m.TargetRule(): true},
		}
		res, err := a.Run(opts)
		sb.Cleanup()
		if err != nil {
			fmt.Fprintln(os.Stderr, m.ID(), "analyze failed:", err)
			continue
		}
		found := false
		for _, f := range res.Findings {
			if f.RuleID == m.TargetRule() {
				found = true
				break
			}
		}
		mark := "✓"
		if !found {
			mark = "✗"
		} else {
			detected++
		}
		fmt.Printf("%s %s\n  mutation: %s\n  detected: %s\n\n", mark, m.TargetRule(), mres.MutatorID, ynStr(found))
	}
	score := 0
	if total > 0 {
		score = detected * 100 / total
	}
	fmt.Printf("Result:\n%d/%d mutations detected\nMutation Detection Score: %d%%\n", detected, total, score)
	if detected < total {
		return 1
	}
	return 0
}

func ynStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
