# Sim 已知问题（未修复）

本文件是**未修复问题**的唯一清单，覆盖前端（parser/analyze）、LLVM 后端（compiler/codegen）
与编译驱动/诊断（compile/lex/report）。

**维护规则**

- **单个问题修复后立即从本文件删除**（不保留已修复条目；历史见 git log）。
- 其他任务过程中发现、但无法当场修复的问题，必须补录到本文件（格式同下）。
- 每条包含：现象 / 最小复现 / 影响 / 临时绕过（如有）/ 修复位置（如明确）。
- 编号一经分配不复用、不重排（修复删除后编号留空），新问题顺延。
- 未来规划与待办见 [`docs/plans.md`](plans.md)；go-llvm 库自身问题见
  [`docs/go-llvm-issues.md`](go-llvm-issues.md)。

复现方式：把片段存为 `x.sim`，在仓库根运行 `go run -tags analyze . x.sim`（前端）或
`go run -tags compile . x.sim`（含后端）；诊断统一输出到 **stderr**（管道捕获时用 `2>&1`）。

> 诊断行为（2026-09-26 改造，提交 `1fba40f`）：前端一次编译可报告多条错误（parse/analyze 均带
> 错误恢复，以语句/全局为隔离单位）；后端未支持或防御性检查失败不再抛裸 panic，统一渲染为
> `internal compiler error: ...` + Go 栈（ICE）。

---

## 一、前端（parser / analyze）

### F9. 无隐式数值转换 —— 待修复

- **现象**：旧 C 后端靠 C 隐式转换兜底，codegen 严格按 LLVM 类型调用，类型不匹配即 ICE：
  `internal compiler error: ir.Builder.Call: argument 0 type i64 does not match parameter type i32`（含 Go 栈）。
- **最小复现**：
  ```sim
  @extern(putchar)
  pub let putchar: (i32) -> i32

  let main = () {
      let t = (65, 66, 67)
      putchar(t[0])
  }
  ```
- **影响**：`analyzeCall` 只对字面量/union 注入做转换，非字面量表达式的隐式转换缺口暴露。
- **临时绕过**：显式标注类型或 `as` 转换（`let t: (i32, i32, i32) = ...`）。
- **修复位置**：`analyzeCall` 实参改走期望类型校验（字面量适配 + 非字面量报诊断），
  并审计其余期望类型消费点。

### F17. 赋值左结合，链式赋值不可用 —— 待修复（低优先）

- **现象**：`parseBinaryExpr` 对赋优先级用 `prec+1` 递归，`a = b = c` 解析为 `(a = b) = c`，
  analyze 因左值临时值/不可变报 `must not temporary`（C/Go 等主流语言赋值右结合）。
- **修复位置**：赋值运算符递归时用 `minPrec = prec`（右结合）。

### F18. PkgScope 的 include 解析顺序非确定 —— 待修复（低优先）

- **现象**：`PkgScope.LookupValue/LookupType` 遍历 `includes`（map 序）找符号，
  多个 include 包导出同名符号时解析结果随 map 迭代序漂移，且无「歧义」诊断。
  （`LookupBind` 已按包名稳定排序遍历，不受此条影响。）
- **影响**：同名导出符号的行为不可复现。修法：按确定序遍历 + 歧义报错，或显式禁止歧义。

### F19. 循环类型诊断后不短路，`GetUnderlying` 潜在无限递归 —— 待修复（低优先）

- **现象**：`analyzeGlobalType` 报告 `CircularReference` 后仍继续 `analyzeGlobalValue`；
  `type A A` 这类环在后续表达式（如 `as` 走 `GetUnderlying`）中可能无限递归到栈溢出。
- **最小复现**：
  ```sim
  type A A
  let main = () {
      let x = 1 as A
  }
  ```
- **修复位置**：类型阶段有错时短路后续阶段（或在 `GetUnderlying` 加环检测防御）。

### F21. `lex.peek` 失败走裸 panic —— 待修复（低优先）

- **现象**：`peek` 的 `Seek` 失败用 `stlerror.MustWith` 裸 panic，未包装为 `lex.Error`，
  不会被 Parser 的 `scan` 恢复逻辑转换为诊断（直接成为 ICE）。
- **修复位置**：与 `next` 一致，panic(`&Error{...}`)。

### F22. 同名自定义类型的诊断无法区分包 —— 待修复（低优先）

- **现象**：跨包同名类型赋值（`114afdd` 起）会正确报错，但消息为
  `expected expression type 'Mine' but got 'Mine'`——类型打印（`CustomType.String()`）只有定义名。
- **影响**：用户无法看出是哪个包的类型不匹配。
- **修复位置**：`globals.TypeDef` 记录所属包（或稳定限定名），打印时带包前缀。

### F23. union 注入不接受未标注的数值字面量 —— 待修复

- **现象**：union 期望类型下整数字面量被默认为 f64（`analyzeInteger` 的 else 分支），
  而 union 注入分支只接受与成员完全 `Equal` 的值，因此 `let u: U = 65` 报
  `expected expression type 'U' but got 'f64'`。
- **最小复现**：
  ```sim
  type U i32 | bool
  let main = () {
      let u: U = 65
  }
  ```
- **影响**：union 最自然的字面量注入写法不可用（与 F9 同源：缺隐式数值转换/期望类型传播）。
- **临时绕过**：`let x: i32 = 65; let u: U = x`。
- **修复位置**：`analyzeInteger` 在期望 union 时按成员匹配，或 union 注入前对字面量重试成员类型。

---

## 二、LLVM 后端（compiler/codegen）遗留与限制

### B1. 逃逸闭包的 ctx 生命周期（已知限制，继承旧设计）

- 闭包 ctx 在定义点栈分配（入口块 alloca）；捕获了局部变量的闭包作为返回值逃逸出定义函数后再调用
  属未定义行为（旧 C 后端同样，实测返回悬垂垃圾）。
- 安全范围：定义帧内使用（GetBind、闭包作实参、返回无捕获闭包 `{fn,null}`）。
- 另：同一闭包字面量在循环内多次求值会复用同一 ctx 槽位，跨迭代保留多个实例会互相串写。
- 根治需堆分配 ctx（语言层还需所有权/GC 设计）；长期规划见 `docs/plans.md` P13。

### B2. 全局常量初始化覆盖面窄于旧后端（已知限制）

- `genConstExpr` 只支持字面量与其常量嵌套；`let g: i32 = 1 + 1`、`let g: i64 = 65 as i64`、
  union 注入等会在全局作用域报
  `internal compiler error: codegen (package ...): non-constant global initializer is not supported: g = ... (*locals.Binary)`
  （消息中的表达式以 `%s` 打印 HIR 节点，但 HIR 节点未实现 `String()`，实际输出 Go 结构体垃圾，
  如 `f = &{%!s(bool=false) %!s(*locals.IdentExpr=&{0x...})}`——只有 `%T` 类型名有效；
  建议消息改用 `hir.Print` 渲染）。
- 旧 C 后端把这些交给 C 编译器可编。后续补常量折叠/转换即可。

### B3. `for x in *getp()` 快照语义差异（低优先）

- 旧后端因 `Temporary()` 会把数组值拷进临时变量（快照），codegen 只求一次指针、逐轮读活内存；
  除「循环体内修改同一块内存」外不可观察。

### B4. `genAddr` 数组索引的非 Array 回退分支静默（低优先）

- 对非 `llvm.ArrayType` 的基址静默返回空临时地址（当前唯一可达是零尺寸数组）；
  建议仅对零尺寸数组放行，其余 panic。

### B5. 元组常量索引越界无前置检查（低优先）

- `t[9]`（长度为 3 的元组常量）在 codegen 取元素时触发 Go 运行时 panic
  `index out of range [9] with length 3`，统一渲染为 `internal compiler error: runtime error: ...`；
  建议在 analyze/codegen 加下标范围检查，给出带位置的诊断。

### B6. `genCall` 快速路径 2 的求值时序号（低优先）

- 有捕获 `*locals.Func` 字面量被调时先求值实参再构造闭包 ctx，与旧后端「先求被调方」相反；
  仅当实参写回被捕获变量时可观察，副作用次数不变。修法：`buildClosureCtx` 提到 `genCallArgs` 之前。

### B7. 代码清理项（低优先）

- `Ident.ExternalFunc` 死字段（M4 后判定改用 `FuncSymbol`），可删除避免误用。
- `c.ctx.idents` 跨函数局部符号残留（按 `hir.Ident` 指针唯一，无正确性问题，随程序规模增长）。

### B8. 函数值全局不能被更早的函数体前向引用（已知差异）

- `let gadd = add` 这类函数值全局的符号登记晚于 `genFuncDecls`；更早函数体引用会报
  `internal compiler error: codegen (package ...): symbol gadd is not registered yet (dependency module must be generated first)`。
  旧后端受文件作用域声明顺序限制同样如此，非回归。补齐可让 `genGlobalVarDecls` 覆盖该形态。

---

## 三、编译驱动与诊断（compile / lex / report）

### C4. `cacheLock.Close` 先关闭 fd 再解锁 —— 待修复（低优先）

- **现象**：`compiler/compile/cache.go` 的 `Close` 先 `f.Close()` 再 `Unlock()`——对已关闭 fd
  做 `flock(LOCK_UN)` 返回 EBADF，错误被 `defer locker.Close()` 吞掉；行为靠「close 自动释放
  flock」侥幸正确。
- **修复位置**：改为先 `Unlock()` 再 `Close()`，并记录/返回错误。

### C5. 诊断 `Format` 对共享 reader 有位置副作用 —— 待修复（低优先）

- **现象**：`report.report.Format`/`sourceLines` 会 Seek + ReadRune 移动 `Position.Reader`。
  当前「解析完成后才渲染」的时序下无害，但格式化带隐藏副作用，多线程/流式诊断时是隐患。
- **修复位置**：渲染前保存/恢复 reader 位置，或复制为独立读取器。
