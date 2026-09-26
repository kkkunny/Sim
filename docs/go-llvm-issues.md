# go-llvm 问题记录

本文件记录 Sim LLVM 后端开发中发现的、疑似 `github.com/kkkunny/go-llvm` **库自身**的
bug、API 缺失或行为异常，由项目作者统一反馈给 go-llvm 作者。

**维护规则**

- 只记录**疑似库问题**（与 Sim 代码无关）；Sim 自身缺陷（bug）见
  [`docs/known-issues.md`](known-issues.md)，未来规划见 [`docs/plans.md`](plans.md)。
- **单个问题在 go-llvm 修复后立即从本文件删除**；Sim 侧的临时绕过同步回退。
- 发现即记录：其他任务过程中遇到的疑似库问题/缺失 API，也补录到本文件。
- 每条必须包含可复现的最小信息；无法确认时标注"待确认"。
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
- **影响**：（Sim 哪些任务/功能受影响）
- **临时绕过**：（无 / 具体做法）
- **状态**：待反馈 / 已反馈 / 已修复（附版本）
```

## 记录

当前无未解决项。

> 上一个问题「缺少公开的基本块是否已终结查询 API」已在 go-llvm
> `v0.0.0-20260926104146-5b02948e8001` 提供 `Terminator()`/`IsTerminating()` 后删除，
> Sim 侧自维护的 `terminated` 临时状态已回退（提交 `a5412ee`）。
