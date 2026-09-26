# Sim 已知问题（未修复缺陷）

本文件是**未修复缺陷（bug）**的唯一清单：与既定语义/预期不符的错误行为、崩溃/ICE、
非确定性、诊断缺陷、资源/生命周期错误，覆盖前端（parser/analyze）、LLVM 后端
（compiler/codegen）与编译驱动/诊断（compile/lex/report）。

**边界**

- 收录标准：**需要改代码使其正确**的错误行为；每条必须给出最小复现与期望行为。
- **不收**：功能缺口、语言设计取舍、性能/架构优化、重构、技术债、测试工程 →
  [`docs/plans.md`](plans.md)；疑似 go-llvm 库自身问题/缺 API →
  [`docs/go-llvm-issues.md`](go-llvm-issues.md)。
- 语言语义未规定的行为差异（求值顺序、快照语义等）先在 plans 明确语义，再决定是否成为缺陷。

**维护规则**

- **单个问题修复后立即从本文件删除**（不保留已修复条目；历史见 git log）。
- 其他任务过程中发现、但无法当场修复的**缺陷**必须补录到本文件（格式同下）。
- 每条包含：现象 / 最小复现 / 影响 / 临时绕过（如有）/ 修复位置（如明确）。
- 编号一经分配不复用、不重排（修复删除后编号留空），新问题顺延。

复现方式：把片段存为 `x.sim`，在仓库根运行 `go run -tags analyze . x.sim`（前端）或
`go run -tags compile . x.sim`（含后端）；诊断统一输出到 **stderr**（管道捕获时用 `2>&1`）。

> 诊断行为（2026-09-26 改造，提交 `1fba40f`）：前端一次编译可报告多条错误（parse/analyze 均带
> 错误恢复，以语句/全局为隔离单位）；后端未支持或防御性检查失败不再抛裸 panic，统一渲染为
> `internal compiler error: ...` + Go 栈（ICE）。

---

## 一、前端（parser / analyze）

### F9. analyzeCall 实参不校验类型，不匹配延后为 codegen ICE —— 待修复

- **现象**：数值字面量语义已定为「按期望类型适配，其余表达式必须显式 `as`」，但 `analyzeCall`
  仅按期望类型分析实参、不校验结果类型；非字面量不匹配时直接进入 codegen：
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
- **期望行为**：报带位置的类型不匹配诊断（如 `expected expression type 'i32' but got 'i64'`），
  用户以 `as i32` 或显式标注（`let t: (i32, i32, i32) = ...`）修正。
- **修复位置**：`analyzeCall` 实参改走期望类型校验；并审计其余期望类型消费点
  （返回/赋值/元组/数组/struct 字段已校验；数组下标 LLVM GEP 接受任意位宽整数，无需转换）。

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
  `expected expression type 'U' but got 'f64'`（与已定的「字面量按期望类型适配」语义不符）。
- **最小复现**：
  ```sim
  type U i32 | bool
  let main = () {
      let u: U = 65
  }
  ```
- **临时绕过**：`let x: i32 = 65; let u: U = x`。
- **修复位置**：`analyzeInteger` 在期望 union 时按成员匹配，或 union 注入前对字面量重试成员类型。

---

## 二、LLVM 后端（compiler/codegen）

### B2. 非常量全局初始化器的 ICE 消息不可读 —— 待修复

- **现象**：全局初始化器不受常量集支持时，`genGlobalVar` 的 ICE 消息以 `%s` 打印 HIR 节点，
  但 HIR 节点未实现 `String()`，输出 Go 结构体垃圾：
  `internal compiler error: codegen (package ...): non-constant global initializer is not supported: g = &{%!s(bool=false) %!s(*locals.IdentExpr=&{0x...})}`。
- **最小复现**：
  ```sim
  let g: i32 = 1 + 1
  let main = () { }
  ```
- **影响**：用户无法从诊断得知哪个表达式不受支持（只有 `%T` 类型名有效）。
- **修复位置**：错误消息改用 `hir.Print` 渲染表达式；常量覆盖面补齐见 plans P5（全局常量折叠与初始化器覆盖面）。

### B4. `genAddr` 数组索引的非 Array 回退分支静默（低优先）

- **现象**：对非 `llvm.ArrayType` 的基址静默返回空临时地址（当前唯一可达是零尺寸数组），
  掩盖潜在的错误输入。
- **修复位置**：仅对零尺寸数组放行，其余 panic。

### B5. 元组常量索引越界无前置检查（低优先）

- **现象**：`t[9]`（长度为 3 的元组常量）在 codegen 取元素时触发 Go 运行时 panic
  `index out of range [9] with length 3`，统一渲染为 `internal compiler error: runtime error: ...`。
- **修复位置**：在 analyze/codegen 加下标范围检查，给出带位置的诊断。

### B6. `genCall` 快速路径 2 的求值时序号（低优先）

- **现象**：有捕获 `*locals.Func` 字面量被调时先求值实参再构造闭包 ctx，与既有约定
  「先求被调方」相反；仅当实参写回被捕获变量时可观察，副作用次数不变。
- **修复位置**：`buildClosureCtx` 提到 `genCallArgs` 之前。

### B8. 函数值全局不能被更早的函数体前向引用（已知差异，低优先）

- **现象**：`let gadd = add` 这类函数值全局的符号登记晚于 `genFuncDecls`；更早函数体引用会报
  `internal compiler error: codegen (package ...): symbol gadd is not registered yet (dependency module must be generated first)`。
  旧后端受文件作用域声明顺序限制同样如此，非回归。
- **修复位置**：让 `genGlobalVarDecls` 覆盖该形态。

---

## 三、编译驱动与诊断（compile / lex / report）

### C5. 诊断 `Format` 对共享 reader 有位置副作用 —— 待修复（低优先）

- **现象**：`report.report.Format`/`sourceLines` 会 Seek + ReadRune 移动 `Position.Reader`。
  当前「解析完成后才渲染」的时序下无害，但格式化带隐藏副作用，多线程/流式诊断时是隐患。
- **修复位置**：渲染前保存/恢复 reader 位置，或复制为独立读取器。
