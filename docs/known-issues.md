# Sim 已知问题（未解决）

本文件汇总当前**尚未解决**的问题：前端（parser/analyze）缺陷、LLVM 后端遗留与已知限制。
每条都带最小复现；状态在修复后更新（标记 ✅ 并注明提交/版本），修复后不要删除条目。

- 状态图例：`待修复` / `低优先` / `已知限制`（设计取舍，非缺陷）
- 库自身问题单独记录在 [`docs/go-llvm-issues.md`](go-llvm-issues.md)
- 历史迁移记录见 [`docs/superpowers/plans/2026-09-26-llvm-backend-migration.md`](superpowers/plans/2026-09-26-llvm-backend-migration.md)

复现方式：把片段存为 `x.sim`，在仓库根运行 `go run -tags analyze . x.sim`（前端）或
`go run -tags compile . x.sim`（含后端），观察输出。

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
  - `if b { ... }`（条件是裸标识符）报 `expected 'ident' but got 'let'`（或空 body 时触发 F3）；
  - `for x in a { ... }`（range 是裸标识符）报 `expected ':' but got '('`。
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
- **临时绕过**：条件/范围用非裸标识符形式：`if b == true { }`、`for x in a[0] { }`、
  `for x in f() { }`。

### F3. 报告器在部分解析错误路径 panic（slice bounds）—— 待修复

- **现象**：`panic: runtime error: slice bounds out of range [:-1]`（`compiler/report/report.go` 位置切片），
  取代本应给出的诊断。
- **最小复现**（任一）：
  ```sim
  let f = () -> i32 { return
  }
  ```
  ```sim
  let main = () {
      let b: bool = true
      if b {
      }
  }
  ```
- **影响**：错误输入直接崩掉编译器，报错不可读。
- **临时绕过**：避免上述写法（`return` 后补分号；`if` 用 F2 的绕过形式）。

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
- **现状**：llgen 报 `panic: llgen: 非 unit 函数 ... 不能使用空 return（前端漏校验）`（可读防御）。
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
- **现状**：llgen 报 `panic: llgen: 不能为 unit 类型的变量 x 分配存储（类型 unit）`。
- **修复位置**：前端应拒绝 unit 型 `let`（或规定其忽略语义）。

### F9. 无隐式数值转换 —— 待修复

- **现象**：旧 C 后端靠 C 隐式转换兜底，llgen 严格按 LLVM 类型调用，类型不匹配直接 panic：
  `ir.Builder.Call: argument 0 type i64 does not match parameter type i32`。
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

### F10. `type bool bool` 别名与内建 bool 不 `Equal` —— 待修复

- **现象**：`mb as bool`（`mb: bool`）得到的类型在需要内建 `types.Bool` 的位置被拒：
  `expected expression type 'bool' but got 'bool'`。
- **最小复现**：
  ```sim
  let main = () {
      let mb: bool = true
      let b = mb as bool
      let c: i32 = b ? 1 : 2
  }
  ```
  （函数实参、`b == true` 条件路径不复现，仅部分 expect 位置触发。）
- **影响**：`std/buildin` 的 `type bool bool` 使 `bool` 名解析为 `_CustomBooleanType`，
  与 `_BooleanType` 不等价。
- **临时绕过**：避免 `as bool` 往返；直接用原始 bool 值。

---

## 二、LLVM 后端（compiler/codegen）遗留与限制

### B1. 逃逸闭包的 ctx 生命周期（已知限制，继承旧设计）

- 闭包 ctx 在定义点栈分配（入口块 alloca）；捕获了局部变量的闭包作为返回值逃逸出定义函数后再调用
  属未定义行为（旧 C 后端同样，实测返回悬垂垃圾）。
- 安全范围：定义帧内使用（GetBind、闭包作实参、返回无捕获闭包 `{fn,null}`）。
- 另：同一闭包字面量在循环内多次求值会复用同一 ctx 槽位，跨迭代保留多个实例会互相串写。
- 根治需堆分配 ctx（语言层还需所有权/GC 设计）。

### B2. 全局常量初始化覆盖面窄于旧后端（已知限制）

- `genConstExpr` 只支持字面量与其常量嵌套；`let g: i32 = 1 + 1`、`let g: i64 = 65 as i64`、
  union 注入等会在全局作用域报 `panic: llgen: 暂不支持非常量全局初始化 ...`。
- 旧 C 后端把这些交给 C 编译器可编。后续补常量折叠/转换即可。

### B3. `for x in *getp()` 快照语义差异（低优先）

- 旧后端因 `Temporary()` 会把数组值拷进临时变量（快照），llgen 只求一次指针、逐轮读活内存；
  除「循环体内修改同一块内存」外不可观察。

### B4. `genAddr` 数组索引的非 Array 回退分支静默（低优先）

- 对非 `llvm.ArrayType` 的基址静默返回空临时地址（当前唯一可达是零尺寸数组）；
  建议仅对零尺寸数组放行，其余 panic。

### B5. 元组常量索引越界无前置检查（低优先）

- `t[9]` 最终由 `Module.Verify` 英文报错；建议加下标范围检查给中文信息。

### B6. `genCall` 快速路径 2 的求值时序号（低优先）

- 有捕获 `*locals.Func` 字面量被调时先求值实参再构造闭包 ctx，与旧后端「先求被调方」相反；
  仅当实参写回被捕获变量时可观察，副作用次数不变。修法：`buildClosureCtx` 提到 `genCallArgs` 之前。

### B7. 代码清理项（低优先）

- `Ident.ExternalFunc` 死字段（M4 后判定改用 `FuncSymbol`），可删除避免误用。
- `c.ctx.idents` 跨函数局部符号残留（按 `hir.Ident` 指针唯一，无正确性问题，随程序规模增长）。

### B8. 函数值全局不能被更早的函数体前向引用（已知差异）

- `let gadd = add` 这类函数值全局的符号登记晚于 `genFuncDecls`；更早函数体引用会报
  `未找到符号`。旧后端受文件作用域声明顺序限制同样如此，非回归。补齐可让
  `genGlobalVarDecls` 覆盖该形态。

---

## 三、go-llvm 库问题

见 [`docs/go-llvm-issues.md`](go-llvm-issues.md)（当前 1 条：`ir.Block` 缺少公开的
"基本块是否已终结"查询 API，Sim 侧以 `terminated` 状态绕过）。
