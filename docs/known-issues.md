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

### F1. 三元运算符优先级错误 —— 待修复

- **现象**：`1 < 2 ? a : b` 被解析为 `1 < (2 ? a : b)`。`?` 在 `parseSuffixExpr` 的
  suffix 循环里被消费，早于二元运算符循环（`compiler/parse/expr.go`）。
- **最小复现**：
  ```sim
  @extern(putchar)
  pub let putchar: (i32) -> i32

  let main = () {
      putchar(1 < 2 ? 65 : 66)
  }
  ```
  实际：`error[unexpected expression]: expected expression type 'bool' but got 'f64'`
- **影响**：条件含比较/算术的三元必须加括号，易误导师也给不出正确指向的报错。
- **临时绕过**：给条件加括号 `(1 < 2) ? 65 : 66`。

### F2. struct 字面量歧义：`if <ident> {` / `for x in <ident> {` —— 待修复

- **现象**：`parsePrimaryExpr` 对 `Ident` 后紧跟 `{` 一律按 struct 字面量解析，因此：
  - `if b { <非空 body> }` 报 `expected 'ident' but got 'let'`（`{` 被当作 struct 字面量字段列表，
    随后恢复还可能再报一条 `unexpected token '}'`）；
  - `if b { }`（空 body）报 `expected '{' but got 'br'`（`b {}` 整个被当作 struct 字面量消费）；
  - `for x in a { <非空 body> }` 同样报 `expected 'ident' but got ...`，
    `for x in a { }` 报 `expected '{' but got 'br'`。
  - **更一般**：条件/范围**表达式以标识符结尾**时同样触发——`if t1 == t2 { ... }`、`if b == c { ... }`、
    `while x > y { ... }` 都会被 `t2`/`c`/`y` 后的 `{` 误判为 struct 字面量；
    条件以 `true`/`false`/数字/`)`/`]` 结尾时可正常解析（这也是既有绕过写法偶发可用的原因）。
- **最小复现**：
  ```sim
  let main = () {
      let b: bool = true
      if b {
          let y: i32 = 1
      }
  }
  ```
- **影响**：`if`/`for-in` 的最自然写法不可用；`examples` 恰好未触发。
- **临时绕过**：条件/范围用非裸标识符形式：`if b == true { }`（右操作数为关键字）、
  `if (a == b) { }`（整个条件加括号）、`for x in a[0] { }`、`for x in f() { }`。

### F4. `analyzeFor` 未把循环变量加入作用域 —— 待修复

- **现象**：`for x in ... { ... }` 体内引用 `x` 报 `unknown identifier 'x'`；
  `compiler/analyze/local.go` 创建了 `hir.Param` 但没有 `scope.AddValue(param)`。
- **最小复现**（用 F2 绕过写法让 range 可解析）：
  ```sim
  @extern(putchar)
  pub let putchar: (i32) -> i32

  let geta = () -> [2]i32 { return [65, 66] }

  let main = () {
      for x in geta() {
          putchar(x)
      }
  }
  ```
- **影响**：for-in 无法端到端使用（llgen 已生成正确的元素绑定，等前端修复后即可工作）。
- **临时绕过**：无（只能用索引循环手动取值）。

### F5. `DeRef` 可变性检查方向错误 —— 待修复

- **现象**：`let p = &mut x; *p = v` 被拒 `must mutable`——检查的是绑定 `p` 的可变性，
  而非引用目标的可变性。
- **最小复现**：
  ```sim
  let main = () {
      let mut x: i32 = 65
      let p = &mut x
      *p = 66
  }
  ```
- **影响**：不可变绑定持有可变引用时无法写回，与直觉/语义不符。
- **临时绕过**：`let mut p = &mut x`。

### F6. `analyzeBinary` 可变性检查漏 `MulAssign` —— 待修复

- **现象**：`x *= v` 不做可变性/临时值检查，而 `x += v` 会拒绝。
- **最小复现**（应被拒绝，实际通过）：
  ```sim
  let main = () {
      let x: i32 = 2
      x *= 3
  }
  ```
- **影响**：对不可变变量/临时值的 `*=` 静默通过（生成后端仍会写非法存储，属隐患）。
- **临时绕过**：无（自律）。

### F7. 无值 `return` 在非 unit 函数中不被拒绝 —— 待修复（后端已有防御）

- **最小复现**：
  ```sim
  let f = () -> i32 { return; }
  let main = () { }
  ```
- **现状**：codegen 报
  `internal compiler error: codegen (package ...): non-unit function ... cannot use a bare return (frontend missed validation)`
  （可读防御；含 Go 栈）。
- **影响**：合法输入集被前端放宽，错误延后到后端。
- **修复位置**：`analyzeReturn` 的 else 分支应校验函数返回类型。

### F8. unit 局部变量被 analyze 放行 —— 待修复（后端已有防御）

- **最小复现**：
  ```sim
  @extern(putchar)
  pub let putchar: (i32) -> i32

  let g = () { putchar(71) }

  let main = () {
      let x = g()
  }
  ```
- **现状**：codegen 报
  `internal compiler error: codegen (package ...): cannot allocate storage for unit-typed variable x (type unit)`。
- **修复位置**：前端应拒绝 unit 型 `let`（或规定其忽略语义）。

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

### F12. 取函数符号地址无前端检查（后端 ICE）—— 待修复

- **现象**：`&g`（`g` 为全局函数）analyze 放行，codegen `genAddr` 的 `FuncSymbol` 分支 panic：
  `taking the address of function symbol ... is not supported (function values are not lvalues)`；
  写在全局初始化器中则报 `non-constant global initializer is not supported: f = &{%!s(...)...}`
  （ICE 消息还是 Go 结构体垃圾，见 B2）。
- **最小复现**：
  ```sim
  let g = () {
  }
  let main = () {
      let x = &g
  }
  ```
- **影响**：合法语法形态得到 ICE 而非带位置的诊断。
- **临时绕过**：无（避免对函数符号取地址）。
- **修复位置**：`analyzeGetReference` 检查操作数是否为函数符号（`*locals.Let` + `FuncType`）。

### F15. 负数字面量不支持 —— 待修复

- **现象**：`-1` 报 `unexpected token '-'`——`parseUnaryExpr` 只处理 `!` 与 `*`，未处理 `-`
  （`compiler/parse/expr.go`）；词法器也不产生带符号整数。
- **最小复现**：
  ```sim
  let main = () {
      let x = -1
  }
  ```
- **影响**：负数只能写成 `0 - 1` 或（有变量时）`x - 1`，常见控制流/算术写法不可用。
- **临时绕过**：`0 - 1`。
- **修复位置**：`parseUnaryExpr` 支持 `KindEnum.Sub`；或词法器在 `-` 后紧邻数字时合并为负数
  （注意与二元减号消歧，前者更稳妥）。

### F17. 赋值左结合，链式赋值不可用 —— 待修复（低优先）

- **现象**：`parseBinaryExpr` 对赋优先级用 `prec+1` 递归，`a = b = c` 解析为 `(a = b) = c`，
  analyze 因左值临时值/不可变报 `must not temporary`（C/Go 等主流语言赋值右结合）。
- **修复位置**：赋值运算符递归时用 `minPrec = prec`（右结合）。

### F18. PkgScope 的 include 解析顺序非确定 —— 待修复（低优先）

- **现象**：`PkgScope.LookupValue/LookupType` 遍历 `includes`（map 序）找符号，
  多个 include 包导出同名符号时解析结果随 map 迭代序漂移，且无「歧义」诊断。
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

### F20. 语言无注释语法 —— 待修复（语言缺口）

- **现象**：`skipWhite` 只跳空格/制表/回车；`//`、`/* */` 都会被当作 token 导致解析错误，
  `.sim` 源码无法写注释。
- **影响**：示例与 std 源码目前零注释；对使用者是明显缺口。
- **修复位置**：词法器 `skipWhite` 中增加行注释/块注释跳过；若为有意取舍应在 README 注明。

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

- **现象**：union 期望类型下整数字面量被默认为 `f64`（`analyzeInteger` 的 else 分支），
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

### F24. 方法绑定（bind）不能跨包查找 —— 待修复

- **现象**：包内 `pub let m | T = ...` 定义的方法，导入该包后对 `pkg::T` 值调用 `v.m()`
  报 `unknown identifier 'm'`——`analyzeMember` 用当前 analyzer 的 `scope.LookupBind` 查找，
  而被导入包的 binds 只登记在其自身 `PkgScope`，未随 `AddInclude/AddExternal` 合并
  （`compiler/analyze/global.go` 内有 `// TODO: 只能绑定本包定义的类型`）。
- **最小复现**：
  ```sim
  --- p1/p1.sim
  pub type S struct {
      name: str
  }
  pub let getname | S = (self: &Self) -> str {
      return self.name
  }
  --- app.sim
  import p1
  let main = () {
      let s = p1::S{name: "x"}
      let n = s.getname()
  }
  ```
  报：`error[unknown identifier]: unknown identifier 'getname'`
- **影响**：类型可导出但方法不可跨包调用，限制了面向对象能力的实际使用。
- **临时绕过**：在调用方包内为同一类型补一个 bind，或改为自由函数。
- **修复位置**：`PkgScope.LookupBind` 递归 includes/externals；或把 bind 表挂到 `TypeDef` 上。

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
