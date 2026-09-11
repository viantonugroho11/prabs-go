package analyzer

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	Roots            []string
	Excludes         []string
	IncludeGenerated bool
	Rules            map[string]RuleConfig // by rule id
	OnlyRules        map[string]bool       // if set, restrict to these
	SeverityFilter   Severity              // include only findings >= this severity ("" = all)
	Verbose          bool
}

type Result struct {
	Findings      []Finding
	FilesAnalyzed int
	RulesExecuted int
	Duration      time.Duration
}

type Analyzer struct {
	Registry *Registry
}

func New(reg *Registry) *Analyzer { return &Analyzer{Registry: reg} }

// generatedRe matches the standard "Code generated ... DO NOT EDIT." marker.
func isGenerated(src []byte) bool {
	// Look only in first 4KB.
	head := src
	if len(head) > 4096 {
		head = head[:4096]
	}
	lines := bytes.Split(head, []byte("\n"))
	for _, ln := range lines {
		if bytes.HasPrefix(bytes.TrimSpace(ln), []byte("//")) &&
			bytes.Contains(ln, []byte("Code generated")) &&
			bytes.Contains(ln, []byte("DO NOT EDIT")) {
			return true
		}
	}
	return false
}

func (a *Analyzer) Run(opts Options) (*Result, error) {
	start := time.Now()
	files, err := collectFiles(opts.Roots, opts.Excludes)
	if err != nil {
		return nil, err
	}
	sort.Strings(files)

	rules := a.selectRules(opts)

	fset := token.NewFileSet()
	var findings []Finding
	analyzed := 0
	for _, path := range files {
		src, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", path, err)
		}
		generated := isGenerated(src)
		if generated && !opts.IncludeGenerated {
			continue
		}
		f, err := parser.ParseFile(fset, path, src, parser.ParseComments)
		if err != nil {
			// syntax error: emit finding for this file and skip
			findings = append(findings, Finding{
				RuleID:    "PRABS-PARSE-000",
				RuleName:  "Parse Error",
				Severity:  SeverityMedium,
				Message:   err.Error(),
				File:      path,
				StartLine: 1,
			})
			continue
		}
		analyzed++
		ctx := &Context{
			FileSet:   fset,
			File:      f,
			FilePath:  path,
			Source:    src,
			Generated: generated,
		}
		for _, rule := range rules {
			rc := opts.Rules[rule.ID()]
			if !rc.Enabled {
				continue
			}
			ctx.Config = rc
			out := rule.Analyze(ctx)
			for i := range out {
				out[i].RuleID = rule.ID()
				if out[i].RuleName == "" {
					out[i].RuleName = rule.Name()
				}
				if out[i].Severity == "" {
					out[i].Severity = rule.Severity()
				}
			}
			findings = append(findings, out...)
		}
	}

	if opts.SeverityFilter != "" {
		min := opts.SeverityFilter.Rank()
		filtered := findings[:0]
		for _, f := range findings {
			if f.Severity.Rank() >= min {
				filtered = append(filtered, f)
			}
		}
		findings = filtered
	}
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].File != findings[j].File {
			return findings[i].File < findings[j].File
		}
		if findings[i].StartLine != findings[j].StartLine {
			return findings[i].StartLine < findings[j].StartLine
		}
		return findings[i].RuleID < findings[j].RuleID
	})
	return &Result{
		Findings:      findings,
		FilesAnalyzed: analyzed,
		RulesExecuted: len(rules),
		Duration:      time.Since(start),
	}, nil
}

func (a *Analyzer) selectRules(opts Options) []Rule {
	all := a.Registry.All()
	if len(opts.OnlyRules) == 0 {
		// Enable rules by presence in Rules map (default enabled if unspecified in map, per config expansion).
		return all
	}
	out := make([]Rule, 0, len(opts.OnlyRules))
	for _, r := range all {
		if opts.OnlyRules[r.ID()] {
			out = append(out, r)
		}
	}
	return out
}

func collectFiles(roots, excludes []string) ([]string, error) {
	var files []string
	seen := map[string]bool{}
	for _, root := range roots {
		root = expandRoot(root)
		info, err := os.Stat(root)
		if err != nil {
			return nil, err
		}
		walkRoot := root
		if !info.IsDir() {
			if strings.HasSuffix(root, ".go") {
				abs, _ := filepath.Abs(root)
				if !seen[abs] {
					seen[abs] = true
					files = append(files, abs)
				}
			}
			continue
		}
		err = filepath.WalkDir(walkRoot, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				name := d.Name()
				if name == ".git" || name == "vendor" || name == "node_modules" {
					return fs.SkipDir
				}
				for _, ex := range excludes {
					if name == ex {
						return fs.SkipDir
					}
				}
				return nil
			}
			if !strings.HasSuffix(path, ".go") {
				return nil
			}
			if strings.HasSuffix(path, "_test.go") {
				// still analyze — user may want smells in tests too
			}
			abs, _ := filepath.Abs(path)
			if !seen[abs] {
				seen[abs] = true
				files = append(files, abs)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	return files, nil
}

func expandRoot(root string) string {
	root = strings.TrimSuffix(root, "/...")
	if root == "" {
		return "."
	}
	return root
}

// LineOf returns the line for a token.Pos in the given fileset.
func LineOf(fset *token.FileSet, pos token.Pos) int {
	if !pos.IsValid() {
		return 0
	}
	return fset.Position(pos).Line
}

// ColOf returns the column for a token.Pos.
func ColOf(fset *token.FileSet, pos token.Pos) int {
	if !pos.IsValid() {
		return 0
	}
	return fset.Position(pos).Column
}

// FuncName returns a printable name for a func declaration.
func FuncName(fd *ast.FuncDecl) string {
	if fd == nil {
		return ""
	}
	name := fd.Name.Name
	if fd.Recv != nil && len(fd.Recv.List) > 0 {
		return "(recv)." + name
	}
	return name
}
