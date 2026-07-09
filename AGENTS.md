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

## Testing (REQUIRED)

Tests live in `tests/` and are run with:

```
go test ./tests/...
```

The framework (`tests/compiler_test.go`) builds the `sim` binary once
(`go build -tags compile`), then compiles and runs every `.sim` file under
`tests/success/` (expected to compile **and** run, exit code 0) and
`tests/failed/` (expected to fail to compile **or** crash at runtime, i.e. a
non-zero exit). It uses the `SIM_OUTPUT` env var to place each produced binary
in a temp dir. Cases run in parallel; the file-lock in `compiler/compile`
serializes the shared `std/*/.sim_cache/` artifacts.

### Adding / modifying a language feature

**Every time a feature is added or changed, you MUST add or update the
corresponding `.sim` test(s) and ensure the whole suite still passes.** This is
mandatory, not optional:

- For a new capability, add a `tests/success/<feature>.sim` that exercises it
  and exits cleanly. If the feature can fail in a new way, add a
  `tests/failed/<feature>.sim` that triggers that error/panic.
- For a behavior change, update the affected `tests/success/*.sim` /
  `tests/failed/*.sim` so they reflect the new semantics.
- Before considering any change done, run `go test ./tests/...` and confirm
  **all** cases pass (0 failures). Do not commit with failing tests.

## Conventions / gotchas

- `.sim_cache/` dirs are generated build artifacts created under each package
  dir during `compile`; they are gitignored. Don't commit them.
- `tests/success/` and `tests/failed/` hold `.sim` integration tests (see
  "Testing" above). `tests/compiler_test.go` is the runner — do not add
  in-process Go unit tests for the compiler packages (parser/analyzer call
  `os.Exit(1)` and cannot be unit-tested in-process).
- Module is Go 1.25 and depends on `github.com/kkkunny/stl`; the `stlerror.Must*`
  helpers panic on error, so failures surface as panics.
