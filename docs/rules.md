# Rule catalog

Nine rules. Nine polite ways for `prabs-go` to inform you that the code you
shipped last Friday is, in fact, exactly what it looks like. Thresholds are
configurable, because taste is a spectrum and denial is a river.


## PRABS-CPLX-001 — High Cyclomatic Complexity
Severity: **high**. Counts branches: `if`, `for`, `range`, `case`, `select`, `&&`, `||`.
Default threshold: 10.

## PRABS-FUNC-001 — Long Function
Severity: **medium**. Lines between `{` and `}` of a function body.
Default threshold: 100.

## PRABS-FUNC-002 — Too Many Parameters
Severity: **medium**. Counts every parameter name (grouped decls counted individually).
Default threshold: 5.

## PRABS-NEST-001 — Deep Nesting
Severity: **medium**. Recursive depth over `if / for / range / switch / type-switch / select`.
Default threshold: 4.

## PRABS-DUP-001 — Duplicate Code
Severity: **medium**. Hashes each function body (printed AST). Flags identical bodies
with ≥ `min_statements` (default 5) top-level statements.

## PRABS-ERR-001 — Ignored Error
Severity: **medium**. Flags:
- `x, _ := f()` — blank in a non-first LHS position.
- Bare call `foo()` whose name matches an error-returning heuristic
  (`read`, `write`, `open`, `close`, `encode`, `decode`, `marshal`, `unmarshal`, `parse`,
  `exec`, `run`, `send`, `recv`, `dial`, `commit`, `rollback`).

## PRABS-ERR-002 — Error Wrapping Missing
Severity: **low**. `fmt.Errorf` uses `%v` while receiving an ident whose name contains
`err` — suggest `%w`.

## PRABS-NAME-001 — Poor Naming
Severity: **low**. Function names in a small blocklist (`foo`, `bar`, `baz`, `tmp`,
`temp`, `data`, `obj`).

## PRABS-FILE-001 — Large File
Severity: **low**. Total lines. Default threshold: 500.
