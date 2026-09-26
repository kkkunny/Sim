# AGENTS.md

Sim is a compiler (written in Go) for the "Sim" language. It compiles `.sim`
source to LLVM IR (via `github.com/kkkunny/go-llvm`), emits an object file per
package, and links a native binary with `clang` (or `gcc` as fallback).

> C 后端已迁移为 LLVM 后端（LLVM 22）。未修复缺陷见 `docs/known-issues.md`，
> 未来规划见 `docs/plans.md`，go-llvm 库问题见 `docs/go-llvm-issues.md`。

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

- `lex` / `parse` / `analyze` / `codegen` — print tokens / AST / HIR / LLVM IR
- `compile` — codegen to LLVM IR, emit `.sim_cache/<pkg>.o` per dependency
  package, link with clang, emit `main.out` in the **current working directory**
  (not next to the source; `config.WorkPath = os.Getwd()`)
- `debug` — compiles and runs `examples/main.sim`

After `compile`, run the result with `./main.out` (from the repo root).

Requires `clang` (or `gcc`) on `PATH` **and** LLVM 22 with development headers
(go-llvm binds system libLLVM via cgo; standard Linux/macOS layout works out of
the box, non-standard prefixes need go-llvm's `make config`).

## Architecture

- `compiler/` — the pipeline: `reader` → `lex`/`token` → `parse`/`ast` →
  `analyze`/`hir` → `codegen` (HIR → `ir.Module` per package, `mod.Verify()`) →
  `compile` (per-package `EmitToFile` + clang link). Package dependencies are
  ordered via a DAG (`compiler/compile/compiler.go`); cross-package symbols are
  declared lazily in the referencing module.
- `std/` — Sim standard library source (`buildin` is auto-imported; `c` exposes
  C bindings). `std/buildin` is auto-imported by every package.
- `examples/main.sim` — the canonical smoke-test program.

## Conventions / gotchas

- `.sim_cache/` dirs are generated build artifacts created under each package
  dir during `compile`; they are gitignored. Don't commit them. Each cache dir
  contains `<pkg>.o`, a `.backend` marker (invalidates caches from other
  backends/ABI versions) and a `.lock` file.
- No Go unit tests exist (`*_test.go` absent). Verification is done by
  compiling/running `.sim` example files, not `go test`. 迁移期每个小功能用
  `/tmp/opencode/sim-cases/` 下的最小 `.sim` 片段做简单验证；正式测试集的建设计划见
  `docs/plans.md` P11。
- Module is Go 1.27 and depends on `github.com/kkkunny/stl` and
  `github.com/kkkunny/go-llvm`; the `stlerror.Must*` helpers panic on error, so
  failures surface as panics.
- `codegen` 的 `genAddr`/`genExpr` 是 lvalue/rvalue 双通道；局部变量 alloca 在函数入口块
  （循环内 `let` 不会每次迭代增长栈）。基本块终结状态直接查询 `Block.IsTerminating()`
  （go-llvm v0.0.0-20260926104146 起提供；早期自维护的 `terminated` 标志已回退，
  见 `docs/go-llvm-issues.md`）。

## 问题与规划记录（三个活文档）

- `docs/known-issues.md`：所有**未修复**问题（前端/后端/驱动/工具链），每条带最小复现与
  临时绕过；**修完即删**。其他任务过程中发现、无法当场修复的问题必须补录到本文件。
- `docs/plans.md`：未来规划与待办事项（优化/重构/技术债/测试等），**完成即删**；
  新发现的未来事项补录到对应分类。
- `docs/go-llvm-issues.md`：疑似 go-llvm 库自身的问题/缺失 API，**库修复后删除**（同步回退
  Sim 侧临时绕过）；发现即记录。

## go-llvm 问题记录（重要）

LLVM 后端基于 `github.com/kkkunny/go-llvm`，该库迭代快、目前尚不稳定，可能存在 bug 或
API 缺失。开发中若发现**疑似 go-llvm 自身**的问题（非 Sim 代码 bug）：

1. **不要**只在 Sim 侧绕过就算了；
2. 记录到 `docs/go-llvm-issues.md`（格式：日期 / 版本 / 现象 / 最小复现 / 期望行为 /
   实际行为 / 影响 / 临时绕过 / 状态）；
3. 由项目作者统一反馈给 go-llvm 作者；**库修复后回退临时绕过并从该文档删除条目**。

依赖版本固定为伪版本（pseudo-version），升级 go-llvm 需单独评估并跑通
`examples/main.sim`。
