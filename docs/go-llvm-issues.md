# go-llvm 问题记录

本文件记录在 Sim LLVM 后端开发中发现的、疑似 `github.com/kkkunny/go-llvm` **库自身**的
bug、API 缺失或行为异常，由项目作者统一反馈给 go-llvm 作者。

**记录原则**

- 只记录**疑似库问题**（与 Sim 代码无关）。Sim 自己的 bug 不写在这里。
- 每条记录必须包含可复现的最小信息；无法确认时标注"待确认"。
- 若已在 Sim 侧做临时绕过，务必写清楚，便于库修复后回退。
- 新的在上，旧的在下。

**记录格式**

```markdown
## YYYY-MM-DD <一句话标题>

- **版本**：go-llvm <pseudo-version / commit>，LLVM <version>，平台 <os/arch>
- **现象**：
- **最小复现**：（代码片段或操作步骤）
- **期望行为**：
- **实际行为**：
- **影响**：（Sim 迁移中哪些任务受影响）
- **临时绕过**：（无 / 具体做法）
- **状态**：待反馈 / 已反馈 / 已修复（附版本）
```

## 记录

## 2026-09-26 缺少公开的"基本块是否已终结"查询 API

- **版本**：go-llvm v0.0.0-20260926025656-65bc762ba2e5，LLVM 22.1.8，linux/amd64
- **现象**：`ir.Block` 没有公开的 `IsTerminating()` / `Terminator()` 方法；库内部
  （`ir/inst_cast.go` 的 `requireTerminator`）使用了 `LLVMIsATerminatorInst`，但未暴露给使用者。
- **最小复现**：前端代码生成器在结束一个函数体时，需要判断当前基本块是否已有终结指令
  （ret/br/switch/unreachable），否则插入 `ret` 会产生"终结指令后的指令"非法 IR；
  当前只能自行记录状态或解析 `Block.LastInst()` 的类型（不可靠，void 调用也是 void 类型）。
- **期望行为**：`ir.Block` 提供 `Terminator() (Value[DynT], bool)` 或 `IsTerminating() bool`。
- **实际行为**：无公开 API。
- **影响**：Sim 迁移任务 E2/E3/E4 及所有控制流代码生成（需要在块结束时做兜底终结）。
- **临时绕过**：`compiler/llgen` 自行维护 `terminated` 状态标志，每次 MoveToEnd 重置、
  发射终结指令时置位。
- **状态**：待反馈

