# Sim 已知问题（未解决）

本文件汇总当前**已知问题**（未解决项 + 保留的已修复回归项）：前端（parser/analyze）缺陷、
LLVM 后端遗留、编译驱动/诊断（compile/lex/report）缺陷与架构/工程 backlog。
每条都带最小复现；状态在修复后更新（标记 ✅ 并注明提交/版本），修复后不要删除条目。

- 状态图例：`待修复` / `低优先` / `已知限制`（设计取舍，非缺陷）/ `✅ 已修复`（保留作回归用例）
- 库自身问题单独记录在 [`docs/go-llvm-issues.md`](go-llvm-issues.md)
- 历史迁移记录见 [`docs/superpowers/plans/2026-09-26-llvm-backend-migration.md`](superpowers/plans/2026-09-26-llvm-backend-migration.md)

复现方式：把片段存为 `x.sim`，在仓库根运行 `go run -tags analyze . x.sim`（前端）或
`go run -tags compile . x.sim`（含后端），观察输出。诊断统一输出到 **stderr**
（管道捕获时用 `2>&1`）。

> 诊断行为说明（2026-09-26 改造，提交 `1fba40f`）：前端一次编译可报告多条错误
> （parse/analyze 均带错误恢复，以语句/全局为隔离单位）；后端未支持或防御性检查失败
> 不再抛裸 panic，统一渲染为 `internal compiler error: ...` + Go 栈（ICE）。
> 下方各条「现状」中的后端消息均为 ICE 形式（取首行引用，栈省略）。

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
- **临时绕过**：条件/范围用非裸标识符形式：`if b == true { }`、`for x in a[0] { }`、
  `for x in f() { }`。

### F3. 报告器在部分解析错误路径 panic（slice bounds）—— ✅ 已修复（1fba40f）

- **修复**：`compiler/report/report.go` 重写源码行提取与高亮区间计算（切片下标一律 clamp），
  `ReadFromTo` 增加负区间保护；非法列号/越界位置不再崩掉诊断渲染。
- **修复前现象**：`panic: runtime error: slice bounds out of range [:-1]`（`compiler/report/report.go` 位置切片），
  取代本应给出的诊断。
- **最小复现**（现输出正常诊断，保留作为回归用例）：
  ```sim
  let f = () -> i32 { return
  }
  ```
  现在输出：`error[unexpected token]: unexpected token 'br'`
  ```sim
  let main = () {
      let b: bool = true
      if b {
      }
  }
  ```
  现在输出：`error[expected token]: expected '{' but got 'br'`（F2 的 struct 字面量歧义表现）
- **影响（修复前）**：错误输入直接崩掉编译器，报错不可读。
- **临时绕过**：无需（已修复）。

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
- **现状**：llgen/codegen 报
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
- **现状**：llgen/codegen 报
  `internal compiler error: codegen (package ...): cannot allocate storage for unit-typed variable x (type unit)`。
- **修复位置**：前端应拒绝 unit 型 `let`（或规定其忽略语义）。

### F9. 无隐式数值转换 —— 待修复

- **现象**：旧 C 后端靠 C 隐式转换兜底，llgen 严格按 LLVM 类型调用，类型不匹配即 ICE：
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

### F10. `type bool bool` 别名与内建 bool 不 `Equal` —— ✅ 已修复（114afdd，根因见 F13）

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
- **修复**：`CustomType.Equal` 重写（见 F13），与内建类型比较时递归到底层，两个方向均等价；
  `mb as bool` 后参与条件判断不再报 `but got 'bool'`。
- **临时绕过**：无需（已修复）。

### F11. 显式类型标注的 `main` 绕过签名检查 —— ✅ 已修复（114afdd）

- **现象**：`analyzeGlobalLetDecl` 的 `main` 签名校验只在**无类型标注**分支执行。
  `let main: (i32) -> i32 = ...` 直达 codegen，`genEntryWrapper` 按 `() -> unit` 调用 `sim_main`
  生成非法调用 → ICE `ir.Builder.Call: expect at least 1 arguments, got 0`；
  `let main: i32 = 5` 更糟——`sim_main` 静默消失，编译成功但程序无入口（返回 0 什么都不做）。
- **最小复现**：
  ```sim
  let main: (i32) -> i32 = (v: i32) -> i32 {
      return v
  }
  ```
- **修复**：签名校验抽为 `checkMainType`，类型标注分支同样执行；
  非函数类型报 `invalid main function`，签名不符报 `unexpected expression`。

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

### F13. `CustomType.Equal` 不对称 + 跨包同名类型互相混淆 —— ✅ 已修复（114afdd，F10 根因）

- **现象（修复前）**：`_CustomBaseType.Equal` 只比 `GetName()`
  （`compiler/hir/types/custom.go`）：
  - **不对称**：内建类型用结构化接口匹配自定义包装，`types.Bool.Equal(customBool)=true`，
    而 `customBool.Equal(types.Bool)=false`（名字比对）——同一对类型两个方向结果相反，
    这是 F10「仅部分 expect 位置触发」的根因；
  - **跨包混淆**：两个包各自的 `type Config ...` 因名字相同被判为同一类型，
    analyze 层接受互相赋值/传参，而 codegen 用按包哈希的 named struct（`stableName`）——
    轻则 LLVM 类型不匹配 ICE，重则布局错乱。
- **最小复现（跨包）**：
  ```sim
  --- p1/p1.sim
  pub type Mine i32
  --- p2/p2.sim
  pub type Mine i32
  --- main
  import p1
  import p2
  let main = () {
      let a: p1::Mine = 1
      let b: p2::Mine = a
  }
  ```
  修复前通过 analyze；修复后报 `expected expression type 'Mine' but got 'Mine'`。
- **修复**：自定义类型按 `GetDef()` 指针身份判等（跨包/别名不混淆）；与内建/字面量类型比较时
  递归到底层，保证两个方向对称。名义区分（同底层不同类型）不受影响。
- **遗留**：报错信息中两个同名类型打印相同（`'Mine' but got 'Mine'`），类型打印缺包限定，
  见 F22。

### F14. 整数字面量与数组长度溢出静默钳制、无范围检查 —— ✅ 已修复（114afdd）

- **现象（修复前）**：`analyzeInteger`/`analyzeType` 用 `strconv.ParseInt` 且忽略错误：
  - `99999999999999999999` 静默变成 `9223372036854775807`（i64 上限）；
  - `let y: i8 = 300` 直接放行（越界值进入 codegen，或静默截断）；
  - `[99999999999999999999]i32` 数组长度同样被钳制。
- **最小复现**：
  ```sim
  let main = () {
      let x = 99999999999999999999
      let y: i8 = 300
  }
  ```
- **修复**：字面量全程 `big.Int` 保精度，按目标整数类型（含自定义类型底层位宽/符号）
  做范围校验，越界报 `integer literal out of range`；数组长度校验 `IsInt64`，
  越界报 `array size out of range`。合法边界值（`u64` 最大值、`i64` 最大值、`i8=127`、`u8=255`）
  验证通过，F9 无隐式数值转换行为不变。

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
- **修复位置**：parseUnaryExpr 支持 `KindEnum.Sub`；或词法器在 `-` 后紧邻数字时合并为负数
  （注意与二元减号消歧，前者更稳妥）。

### F16. 前向类型别名误报 repeated identifier —— ✅ 已修复（114afdd）

- **现象**：`type A B` + `type B i32`（B 在后）时，分析 A 的过程中递归注册了 B，
  外层循环再次处理 B 时 `analyzeCustomTypeDecl` 误报 `the identifier 'B' be redefined`。
- **最小复现**：
  ```sim
  type A B
  type B i32
  let main = () {
      let x: A = 1
  }
  ```
- **修复**：仅当同名的已注册类型来自**另一处声明**（`GetDef()` 指针不同）时才算重定义；
  真重定义（同名声明两次）仍报错。

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
  `type A = A` 这类环在后续表达式（如 `as` 走 `GetUnderlying`）中可能无限递归到栈溢出。
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

- **现象**：F13 修复后跨包同名类型赋值会正确报错，但消息为
  `expected expression type 'Mine' but got 'Mine'`——类型打印（`CustomType.String()`）只有定义名。
- **影响**：用户无法看出是哪个包的类型不匹配。
- **修复位置**：`globals.TypeDef` 记录所属包（或稳定限定名），打印时带包前缀。


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
  union 注入等会在全局作用域报
  `internal compiler error: codegen (package ...): non-constant global initializer is not supported: g = ... (*locals.Binary)`
  （消息中的表达式以 `%s` 打印 HIR 节点，但 HIR 节点未实现 `String()`，实际输出 Go 结构体垃圾，
  如 `f = &{%!s(bool=false) %!s(*locals.IdentExpr=&{0x...})}`——只有 `%T` 类型名有效；
  建议消息改用 `hir.Print` 渲染。2026-09-26 review 确认，见 F12 复现）。
- 旧 C 后端把这些交给 C 编译器可编。后续补常量折叠/转换即可。

### B3. `for x in *getp()` 快照语义差异（低优先）

- 旧后端因 `Temporary()` 会把数组值拷进临时变量（快照），llgen 只求一次指针、逐轮读活内存；
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

### C1. 增量编译缓存不感知依赖变化 —— ✅ 已修复（114afdd）

- **现象（修复前）**：`isCacheValid` 只比对包自身 `.sim` 源文件的 mtime 与 `.o`；
  依赖包源码变更后，依赖方 `.o` 不会重编。跨包 struct 布局/常量在各自模块 codegen 时物化，
  旧依赖方 `.o` 仍按旧布局生成 GEP → 与新 `.o` 混链 → 静默错误行为（读错字段）。
- **最小复现**（三包级联 `app → mid → leaf`，leaf 的 struct 插入字段使后续字段偏移变化）：
  修复前 app 读到错位字段输出 `0`，修复后输出正确值 `2`。
- **修复**：缓存有效性加入「依赖产物 mtime 必须不新于本包产物」检查；
  依赖重编使 `.o` 变新（Topo 序遍历先于依赖方），从而级联失效依赖方缓存。
- **备注**：缓存命中仍会全量 codegen（见 D1），本次修复的是正确性。

### C2. 非 ASCII 源码诊断错位 —— ✅ 已修复（114afdd）

- **现象（修复前）**：词法器 `l.offset++` 按 rune 计数，而 `report.ReadFromTo` 用它做字节级
  `Seek`；`sourceLines` 的「列号→偏移」减法与 `highlightRange` 的 `line[:begin]` 字节切片
  也混用 rune/byte —— 多字节字符（如中文）之后，报错行显示为乱码残片、高亮区间错位。
- **最小复现**：
  ```sim
  let main = () {
      let s = "中文测试"
      let x: i32 = true
  }
  ```
  修复前显示 `M-fM-5M-^KM-hM-/M-^U"` 之类乱码行；修复后正确显示 `let x: i32 = true`。
- **修复**：lexer 偏移改按字节累计；诊断行首改为从 `BeginOffset` 向前做字节回扫；
  高亮列做 rune→byte 换算后再切片。
- **顺带**：`ReadFromTo` 的单次 `Read` 改 `io.ReadFull`（短读可能截断超长行诊断）。

### C3. 重复 import 导致 DAG 重复边，编译失败报裸 UUID —— ✅ 已修复（114afdd）

- **现象（修复前）**：`analyzeImport` 每条 import 语句都向 `Dependencies` append（不去重），
  `buildDAG` 对同一对包重复 `AddEdge`，heimdalr/dag 返回 `EdgeDuplicateError`：
  `error: edge between 'a141ac34-...' and 'fde04dfa-...' is already known`。
  同一包用两个别名 import、或显式 `import std::buildin`（本就自动注入）即触发。
- **修复**：`Dependencies` 按包指针去重（`addDependency`）；`buildDAG` 亦按顶点去重防御。

### C4. `cacheLock.Close` 先关闭 fd 再解锁 —— 待修复（低优先）

- **现象**：`compiler/compile/cache.go` 的 `Close` 先 `f.Close()` 再 `Unlock()`——对已关闭 fd
  做 `flock(LOCK_UN)` 返回 EBADF，错误被 `defer locker.Close()` 吞掉；行为靠「close 自动释放
  flock」侥幸正确。
- **修复位置**：改为先 `Unlock()` 再 `Close()`，并记录/返回错误。

### C5. 诊断 `Format` 对共享 reader 有位置副作用 —— 待修复（低优先）

- **现象**：`report.report.Format`/`sourceLines` 会 Seek + ReadRune 移动 `Position.Reader`。
  当前「解析完成后才渲染」的时序下无害，但格式化带隐藏副作用，多线程/流式诊断时是隐患。
- **修复位置**：渲染前保存/恢复 reader 位置，或复制为独立读取器。

---

## 四、架构与工程 backlog（非缺陷：优化 / 技术债）

### D1. 缓存命中仍全量执行 codegen
- `compileDepPkg` 即使缓存有效也调用 `gen.Generate()`（为填充共享 `codegen.Context`），
  只省下 `EmitToFile`。需把「Context 填充」与「IR 发射」解耦/持久化接口信息（配合 C1 才有意义）。

### D2. 无任何优化管线
- `TargetMachine` 固定 `target.OptNone`（`compiler/codegen/context.go`），模块无 pass、无 LTO，
  产物恒为 -O0。建议至少 `OptDefault` + 可选 `-O` 开关。

### D3. 包编译串行
- `Dagger.OrderedWalk` 串行访问；DAG 上独立包可并行，但 `codegen.Context` 的
  `idents/typeCache/eqFuncs` 是共享 map，需先按包隔离或加锁。

### D4. 字符串常量不去重
- `genString` 每处字面量新建 private global；同模块相同字面量可复用。

### D5. 后序遍历重复实现
- `codegen.go` 驱动的 `collectPkgs` 与 `compile/compiler.go` 的 `collectDeps` 是两套等价遍历。

### D6. HIR 打印顺序不确定
- `locals.Struct.Fields` 是 map，HIR 打印字段顺序随 map 迭代；codegen 已按字段序处理
  （无正确性问题），但测试/快照输出需要稳定顺序。

### D7. bind（方法）语义散落三处
- `analyzeMember`、`GetBind.IsStatic`、`codegen.genGetBind` 各实现一部分，且 `genGetBind`
  在 codegen 期动态构造 HIR（`locals.NewFunc/NewCall`）绕过 analyze 校验。
  建议在 HIR lowering 统一 desugar。

### D8. 无测试体系
- 仓库无 `*_test.go`；known-issues 的最小复现全靠手工跑。建议先为 lex/parse/report 建
  golden 测试，并把本文各条最小复现收进回归脚本。

### D9. 代码卫生
- `compiler/parse/expr.go` 遗留死类型 `type A struct { a int }`；
- `codegen.Ident.ExternalFunc` 死字段（B7）；
- `hir/types/*.go` 多处 `WriteFormat(t.String())`（非恒定格式串，`go vet` 告警；
  当前类型名不含 `%`，无实际风险，但应改 `WriteString`）；
- README 的 `make run` 与 Makefile 过时（与 AGENTS.md 说明重复，建议同步或删除）。

### D10. std 的 HIR 无序列化缓存（长期）
- 每次编译都重新解析/分析 std；buildin 很小可接受，规模扩大后可做序列化缓存。

---

## 五、go-llvm 库问题

见 [`docs/go-llvm-issues.md`](go-llvm-issues.md)（当前 0 条未解决：`ir.Block` 缺少公开的
"基本块是否已终结"查询 API 已在 go-llvm v0.0.0-20260926104146-5b02948e8001 提供
`Terminator()`/`IsTerminating()`；Sim 侧 `terminated` 临时状态已回退为直接查询块）。
