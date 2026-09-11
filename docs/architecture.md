# Architecture

```
cmd/prabs                 CLI entry point
internal/analyzer         Rule interface, registry, driver
internal/rules            Built-in rules
internal/output           text / JSON / SARIF writers
internal/config           .prabs.yaml loader (indent-based YAML subset)
internal/mutation         Sandbox + mutators + verify engine
internal/git              Worktree helpers (changed files, tree copy)
testdata/smells           Positive fixtures per rule
```

## Analyzer flow

```
CLI ──▶ config.Load ──▶ analyzer.Options
                                 │
                                 ▼
                         collect files
                                 │
                                 ▼
       parse (skip generated unless --include-generated)
                                 │
                                 ▼
      for each rule enabled in config → Analyze(ctx)
                                 │
                                 ▼
              sort, filter by --severity, emit
```

## Mutation flow

```
prabs verify PATH
        │
        ▼
NewSandbox(PATH) → temp dir under $TMPDIR (never PATH itself)
        │
        ▼
For each Mutator m:
        m.Mutate(sandbox)  ← writes ONLY inside sandbox
        analyzer.Run(sandbox, OnlyRules={m.TargetRule()})
        record detected/not-detected
        Sandbox.Cleanup()  ← refuses anything outside $TMPDIR
```

## Extending

- **Add a rule**: implement `analyzer.Rule` (`ID`, `Name`, `Description`, `Severity`,
  `Analyze`) in `internal/rules/` and append it to `rules.All()`.
- **Add a mutator**: implement `mutation.Mutator` and append it to `mutation.All()`.
- **Add a new output format**: extend `output.Write` with a new `output.Format`.

## Safety invariants

- The analyzer never writes to disk.
- Mutators write only under a `Sandbox.Dir` that is always a fresh `os.MkdirTemp`
  path under `os.TempDir()`.
- `Sandbox.Cleanup` refuses paths outside `os.TempDir()` (covered by test).
- `git.CopyTree` skips `.git`, so mutation sandboxes never inherit repository
  state that could be used to modify the original.
