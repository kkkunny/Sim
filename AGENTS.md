# AGENTS.md

Sim is a compiler (written in Go) for the "Sim" language. It compiles `.sim`
source to LLVM IR via `github.com/kkkunny/go-llvm`, emits per-package object
files with the LLVM target machine, and links them with `clang` (or `gcc` as
fallback) into `main.out`.

## Running / building (do NOT use `make`)

The `Makefile` is stale: it references a `cmd/` package, a `runtime/build` dir,
and a `tests/` dir that do not exist, and there is no `run` target. The README's
`make run TEST_FILE=...` does not work.

The real entrypoint is the repo-root `package main`, split across build-tagged
files (`lex.go`, `parse.go`, `analyze.go`, `codegen.go`, `compile.go`,
`debug.go`). Run a stage with:

```
go run -tags <stage> . <file.sim>
```

- `lex` / `parse` / `analyze` / `codegen` — print tokens / AST / HIR / generated
  LLVM IR
- `compile` — LLVM codegen, emit the object file, link with clang, emit
  `main.out` in the **current working directory** (not next to the source;
  `config.WorkPath = os.Getwd()`)
- `debug` — compiles and runs `examples/main.sim`

After `compile`, run the result with `./main.out` (from the repo root).

Requires LLVM 23 development libraries (for go-llvm) and `clang` (or `gcc`) on
`PATH`.

## Architecture

- `compiler/` — the pipeline: `reader` → `lex`/`token` → `parse`/`ast` →
  `analyze`/`hir` → `llgen` → `compile`. Package dependencies are ordered via
  a DAG (`compiler/compile/compiler.go`).
- `compiler/llgen/` — the LLVM backend. Layout mirrors the former C backend:
  `llgen.go` (generator state), `context.go` (shared context: `llvm.Context`,
  global symbol names, type cache), `type.go`, `global.go`, `local.go`,
  `expr.go`, `other.go`.
- `compiler/compile/compiler.go` — per-package object emission
  (`target.TargetMachine`) with a `.sim_cache` and final linking.
- `std/` — Sim standard library source (`buildin` is auto-imported; `c` exposes
  C bindings via `@extern`). `std/buildin` is auto-imported by every package.
- `examples/main.sim` — the canonical smoke-test program.

## Code style

See `.ai/docs/code_style.md` (Chinese); follow it for all Go edits.

## Conventions / gotchas

- `.sim_cache/` dirs are generated under each package dir during `compile` and
  contain `<pkg>.o`, `<pkg>.ll`, and a `.backend` version marker used for cache
  invalidation (`config.BackendVersion`). They are gitignored; don't commit
  them.
- `llgen.Context` is shared by all package modules, and so are `llvm.Context`
  named structs; custom type definitions must be emitted exactly once per
  program (guarded by `Context.definedTypes`).
- `target.TargetMachine.ApplyTo(module)` must be called **before** IR
  generation, otherwise loads/stores get default-alignment metadata.
- `llgen.Generate` verifies the module at the end; debug builds of go-llvm also
  validate builder operands aggressively (type mismatches panic).
- No Go unit tests exist (`*_test.go` absent). Verification is done by
  compiling/running `.sim` programs, not `go test`.
- Module is Go 1.27 and depends on `github.com/kkkunny/go-llvm` and
  `github.com/kkkunny/stl`; the `stlerror.Must*` helpers panic on error, so
  failures surface as panics.
