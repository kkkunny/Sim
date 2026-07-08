# AGENTS.md

Sim is a compiler (written in Go) for the "Sim" language. It compiles `.sim`
source to C, then to a native binary via `clang` (or `gcc` as fallback).

## Running / building (do NOT use `make`)

The `Makefile` `build`, `run`, and `test` targets are stale: they reference a
`cmd/` package, a `runtime/build` dir, and a `tests/` dir that do not exist, and
there is no `run` target. The README's `make run TEST_FILE=...` does not work.

The real entrypoint is the repo-root `package main`, split across build-tagged
files (`lex.go`, `parse.go`, `analyze.go`, `codegen.go`, `compile.go`,
`debug.go`). Run a stage with:

```
go run -tags <stage> . <file.sim>
```

- `lex` / `parse` / `analyze` / `codegen` — print tokens / AST / HIR / generated C
- `compile` — codegen to C, compile with clang, emit `main.out` in the **current
  working directory** (not next to the source; `config.WorkPath = os.Getwd()`)
- `debug` — compiles and runs `examples/main.sim`

After `compile`, run the result with `./main.out` (from the repo root).

Requires `clang` (or `gcc`) on `PATH`. The README's "llvm==18" dependency is
outdated — output is C compiled with `clang -std=c11`, not LLVM IR.

## Architecture

- `compiler/` — the pipeline: `reader` → `lex`/`token` → `parse`/`ast` →
  `analyze`/`hir` → `codegen`/`cir` → `compile` (C emission + clang link).
  Package dependencies are ordered via a DAG (`compiler/compile/compiler.go`).
- `std/` — Sim standard library source (`buildin` is auto-imported; `c` exposes
  C bindings). `std/buildin` is auto-imported by every package.
- `include/` — C runtime (`buildin.h` / `buildin.c`) linked into every program.
- `examples/main.sim` — the canonical smoke-test program.

## Conventions / gotchas

- `.sim_cache/` dirs are generated build artifacts created under each package
  dir during `compile`; they are gitignored. Don't commit them.
- No Go unit tests exist (`*_test.go` absent). Verification is done by
  compiling/running `.sim` example files, not `go test`.
- Module is Go 1.25 and depends on `github.com/kkkunny/stl`; the `stlerror.Must*`
  helpers panic on error, so failures surface as panics.
