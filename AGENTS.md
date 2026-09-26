# AGENTS.md

Sim is a compiler (written in Go) for the "Sim" language. It compiles `.sim`
source to C, then to a native binary via `clang` (or `gcc` as fallback).

> **进行中：** C 后端正在迁移为 LLVM 后端（`github.com/kkkunny/go-llvm`，LLVM 22）。
> 迁移计划与 TODO 见 `docs/superpowers/plans/2026-09-26-llvm-backend-migration.md`。
> 迁移完成前，下述 C 管线仍是生效管线。

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
- `llvmgen` — 迁移期临时驱动：analyze → LLVM IR 文本（`compiler/llgen`，尚未接入 compile）
- `llvmcompile` — 迁移期临时驱动：analyze → llgen → `.o` → clang 链接 `main.out`（不写缓存）
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
  compiling/running `.sim` example files, not `go test`. 迁移期每个小功能用
  `/tmp/opencode/sim-cases/` 下的最小 `.sim` 片段做简单验证，正式测试集后续统一补充。
- Module is Go 1.27 and depends on `github.com/kkkunny/stl` and
  `github.com/kkkunny/go-llvm`; the `stlerror.Must*` helpers panic on error, so
  failures surface as panics.

## go-llvm 问题记录（重要）

LLVM 后端基于 `github.com/kkkunny/go-llvm`，该库迭代快、目前尚不稳定，可能存在 bug 或
API 缺失。开发中若发现**疑似 go-llvm 自身**的问题（非 Sim 代码 bug）：

1. **不要**只在 Sim 侧绕过就算了；
2. 记录到 `docs/go-llvm-issues.md`（格式：日期 / 版本 / 现象 / 最小复现 / 期望行为 /
   实际行为 / 影响 / 临时绕过 / 状态）；
3. 由项目作者统一反馈给 go-llvm 作者；修复后回退临时绕过。

依赖版本固定为伪版本（pseudo-version），升级 go-llvm 需单独评估并跑通
`examples/main.sim`。
