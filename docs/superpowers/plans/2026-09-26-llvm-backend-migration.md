# Sim LLVM 后端迁移 实现计划

> **面向 AI 代理的工作者：** 必需子技能：使用 superpowers:executing-plans 逐任务实现此计划。
> 步骤使用复选框（`- [ ]`）语法来跟踪进度。每个小功能完成后提交一次 commit。

**目标：** 将 `compiler/codegen` + `compiler/cir` 构成的 C 文本后端整体替换为
基于 `github.com/kkkunny/go-llvm`（LLVM 22）的 LLVM IR 后端，语义与现行 C 后端完全对等。

**架构：** `analyze`(HIR) → 新的 `compiler/llgen`（HIR → `ir.Module`，每包一个模块，`mod.Verify()`）
→ `compiler/compile`（`TargetMachine.EmitToFile` 产出 `.sim_cache/<pkg>.o`）→ clang 链接 → `main.out`。
删除 `compiler/cir`；跨包声明不再经由 `.h` 头文件，而是按 mangled name 在当前模块惰性创建
external declaration。

**技术栈：** Go 1.27、`github.com/kkkunny/go-llvm`（master，LLVM 22）、clang（链接驱动）、现有独立
analyze/HIR 管线。

---

## 0. 执行约定（阅读者必读）

1. **不做正式测试集**：不加 `tests/*.sim`、不加 runner 脚本（用户后续统一补测试）。
   每完成一个小功能，用最小 `.sim` 片段做**简单验证**：编译 + 运行 + 检查输出/退出码；
   临时用例放 `/tmp/opencode/sim-cases/`，不提交。
2. **go-llvm 问题记录**：go-llvm 迭代快、尚不稳定；若发现疑似库本身的 bug / API 缺失 /
   行为异常，**不要在本仓库绕过就算了**，记录到 `docs/go-llvm-issues.md`
   （格式：日期 / 现象 / 最小复现 / 期望行为 / 影响 / 临时绕过），由用户统一反馈给库作者。
   此约定已写入 `AGENTS.md`。
3. 每个任务完成即 commit，commit message 引用任务编号（如 `feat(llvm): A1 ...`）。
4. 迁移期间保持旧 C 后端可用（新包 `compiler/llgen` 并行开发），切流后再删除旧代码。

## 1. 已确认的决策

| 决策点 | 结论 |
|---|---|
| 架构路线 | **直接 HIR → LLVM IR**，删除 `compiler/cir`，不保留 C 后端 |
| 依赖库 | `github.com/kkkunny/go-llvm` **master**（pseudo-version 固定），LLVM 22 / Go 1.27 |
| 链接 | go-llvm 只产 `.o`，最终用 **clang** 链接（libc/libm/crt，`@extern` C 互操作） |
| 附加范围 | **模块 Verify** + **IR 文本输出调试**；不做 `-O`、不做全局动态初始化 |

已用冒烟测试验证（LLVM 22.1.8 + Go 1.27.1 + clang）：`NamedStruct+SetBody`、`ConstString+ConstGEP`、
外部函数声明/调用、`ICmp/CondBr/PHI`、`Load/ExtractValue`、`mod.Verify()`、
`TargetMachine.EmitToFile(ObjectFile)`、clang 链接运行全部正常；`str` 布局 16B/align 8 与 C ABI 一致。
标准安装布局**无需** go-llvm 的 `make config`。

## 2. 范围

**做**：`compiler/llgen` 新后端（每包一个 `ir.Module`、跨包 external 声明、Verify）；`compile`
产出 `.o` 并 clang 链接；`codegen`/`compile`/`debug` 三个驱动 stage 适配；IR 文本调试输出；
删除 `compiler/cir`、`include/buildin.h/.c`；文档更新。

**不做**（非目标）：优化管线（`-O`）、全局变量动态初始化（ctor）、DWARF 调试信息、JIT、
保留/兼容 C 后端、跨平台验证（仅本机 Linux x86-64）、正式测试集。

## 3. 现状机制盘点：C 宏 → LLVM 落点

| C 后端机制 | 语义 | LLVM 落点 |
|---|---|---|
| `i8..u64/f32/f64/bool` 宏类型 | 标量 | `ctx.Int(8..64)` / `ctx.Float(...)` / `ctx.Bool()`(i1)；符号性在指令层 |
| `str {data,length}` | 字符串值类型 | named struct `{ptr, i64}` |
| `string(s)` | 字符串字面量 | `ConstString` + `ConstGEP` + `ConstNamedStruct` 全局常量 |
| `ADD/SUB/...` `_Generic` | 多态算术 | 按 HIR 类型选 `Add/SDiv/UDiv/FAdd/...` |
| `REM` 浮点 `fmod` | 浮点取模 | **`FRem` 指令**（无需 libm；链接仍带 `-lm` 备用） |
| `SHR` | 有/无符号移位 | 有符号 `AShr`，无符号 `LShr` |
| `NOT` | bool 取反 / 位取反 | `Xor true` / `Not`(xor -1) |
| `SELFADD(x)` `x++` | 自增 | load → add 1 → store |
| `LOGIC_AND/OR` | 短路 | 基本块 + 合流（保留副作用语义） |
| `Ternary` `?:` | 条件表达式 | 分支块 + 合流（**不可**用 `Select`，会双求值） |
| `FUNC_TYPE/FUNC_EXPR_F/FUNC_EXPR_C` | 函数值/闭包 | 胖结构 `{ptr fn, ptr ctx}` |
| `FUNC_CALL` | 间接调用 | ctx 判空分支调用（静态已知的直接调用走快速路径） |
| `FUNC_EQ/FUNC_NEQ` | 函数值比较 | fn/ctx 两个 `ICmp` |
| `ARRAY_TYPE/ARRAY_INDEX` | 数组 | `[N x T]` 一等值 + GEP/ExtractValue（去掉 C 包装 struct） |
| `TUPLE_INDEX` | 元组访问 | named struct 字段 GEP/ExtractValue |
| `UNION_TYPE_INDEX/UNION_VALUE_INDEX` | tagged union | `{i8 tag, payload}`（见 §4.5） |
| `ZERO_TYPE_VALUE` | 零尺寸值 | `{}` 空结构体 `zeroinitializer`/undef |
| `#include`/`.h` 头文件 | 跨包声明 | **消失**：当前模块内按需创建 external declaration |
| `buildin.c` 的 `main` 包装 | 入口 | codegen 生成 `main`（调 `sim_main` 后 `ret 0`） |

## 4. 总体设计

### 4.1 架构与数据流

```
analyze (不变) → HIR ──compiler/llgen（新）──▶ ir.Module（每包一个）
                                                │ mod.Verify()
                                                ▼
compile: TargetMachine.EmitToFile ──▶ .sim_cache/<pkg>.o ──clang──▶ main.out
```

- 新包 `compiler/llgen`（与旧 `compiler/codegen` 并行开发；J1 时改名搬回 `compiler/codegen`）。
- `CodeGenerator` 持有 `*ir.Module` + `*ir.Builder`；保留现有三遍生成结构：
  `genTypeDecl`（opaque named struct 预声明）→ `genTypeDef`（`SetBody` 填充）→
  `genGlobalValue`（变量/函数/入口）。该结构天然对应 LLVM 递归类型需求。
- `llgen.Context`（跨包共享）：`*llvm.Context`、`TargetMachine`、`dataLayout` 字符串、
  `idents map[hir.Ident]→符号描述符`、`typeCache`、辅助函数缓存。
- 每包模块创建时 `SetTargetTriple` + `SetDataLayout`（来自共享 TargetMachine），使
  `DataLayout` 尺寸/对齐查询权威可信（union payload 依赖它）。
- 跨包引用时在**当前模块**按 mangled name 惰性创建 external declaration。
- 辅助函数（相等性、闭包包装）在表达式生成期间惰性创建到当前模块。

### 4.2 类型映射

| Sim (HIR) | LLVM | 备注 |
|---|---|---|
| `i8/i16/i32/i64`、`u8..u64` | `i8..i64` | 符号性只体现在指令选择 |
| `f32/f64` | `float/double` | |
| `bool` | `i1` | |
| `unit`（仅返回位） | `void` | 与 C 后端一致：参数/聚合位的 unit 不支持 |
| `str` | named struct `{ptr, i64}` | ABI 与 C 一致 |
| `&T` / `&mut T` | `ptr`（opaque） | 解引用 load/store 目标类型 |
| tuple | named struct（`e1..en`） | |
| `[T; N]` | `[N x T]` | N=0 或零尺寸元素走 `{}` |
| struct | named struct | 递归/互递归：`NamedStruct` + `SetBody` 两遍 |
| 自定义类型 | 标量别名透传；聚合走 named struct | 复用 `stableName` |
| func 类型（胖） | named struct `{ptr fn, ptr ctx}` | §4.4 |
| tagged union | named struct `{i8, payload}` | §4.5 |
| 零尺寸类型 | `{}` | 值 `zeroinitializer`，相等恒真 |

### 4.3 值模型（lvalue / rvalue 双通道）

- `genExpr(e) llvm.AnyValue` — 右值；`genAddr(e) llvm.Value[llvm.PtrT]` — 左值地址。
- 局部变量与参数统一 `Alloca`（入口块），load/store 访问（支持 `&x`、闭包捕获、范围循环）。
- `Assign`：`addr=genAddr(left)` → `genExpr(right)` → `Store`；复合赋值 `a op= b` 重构为
  addr 只求值一次（现状 `NewAssign(left, Binary(op,left,right))` 复用树节点会双求值）。
- 短路 `&&/||`、`?:`：分支块 + 入口块临时 alloca 合流（统一处理 void/聚合，不用 phi 特判）。

### 4.4 闭包模型

- 函数值 = `named struct { ptr fn; ptr ctx }`：
  - 无捕获：`{fnptr, null}`（`FUNC_EXPR_F` 对应）；`@extern`/纯函数同。
  - 有捕获：捕获收集为 ctx named struct（`_f1.._fn`），包装函数首参数为 ctx `ptr`，
    值为 `{fnptr, &ctx}`（`FUNC_EXPR_C` 对应）；ctx 用 alloca + `InsertValue`/`Store` 构造。
- 调用（`FUNC_CALL` 对应）：`genCall` 保留快速路径（静态已知形态直接 `Call`）；一般函数值走
  运行时分支 `ctx == null ? fn(args...) : fn(ctx, args...)`（两个 `CallIndirect` + 合流）。
- 函数相等：fn 指针 + ctx 各一个 `ICmp`。
- 捕获变量访问（`captureVarsMap`）→ ctx 结构体字段 GEP。
- `GetBind`（方法绑定）保持 HIR 层构造闭包 `locals.Func` 后走 `genFunc`（现有逻辑不动）。

### 4.5 tagged union 模型

- 表示：`{ i8 tag; P payload }`，与 C `struct { u8 t; union {...} v; }` 布局逐一对应。
- payload `P` 构造：取对齐要求最大的成员 `m`，`P = Struct{ m, [size(union)-size(m) x i8] }`；
  `size(union)=roundup(max成员size, max对齐)`，用 `mod.DataLayout().ABISizeOfType/ABIAlignOfType`
  精确计算。
- 注入（`*locals.Union`）：alloca P → bitcast 存成员 → load P → `InsertValue` 组 tag。
- 判别/取值：tag `ExtractValue`；payload 成员 = GEP payload → bitcast → load/store。
- 全零尺寸 union → `{}`。

### 4.6 相等性辅助函数

`genEqual`（数组逐元素、元组逐字段、struct 逐字段、union switch 分派、函数指针）保留
"按需合成辅助函数"结构，改为生成 `ir.Function`；按 HIR 类型键缓存去重；union 用
`builder.Switch`；零尺寸类型恒真短路。

### 4.7 跨包、产物、链接、缓存

- 每包一个 `ir.Module`；引用外部符号时 `GetFunction/GetGlobal` 缺失则创建 external declaration
  （签名由 HIR 推导）。类型按 `stableName` 各模块独立定义，布局确定一致。
- 链接性：`pub` → external（默认）；非 pub → `SetLinkage(LinkageInternal)`；`@extern` → 纯声明。
- **无初始化全局**：包内 `let mut x: T`（无值）在 C 是 tentative definition，LLVM 无 initializer
  的 Global 是声明——需显式 `SetInitializer(zeroinitializer)` 定义；`@extern` 无值才是声明。
- 每包 `EmitToFile(.sim_cache/<pkg>.o)`；缓存有效性检查去掉 `.h`（只比 `.o` 与源 mtime）；
  文件锁不变。
- 链接：`clang <main.o> <dep.o...> -lm -o main.out`（沿用 `util.LookupCCompiler`；输出位置 cwd 不变）。
- 主包生成 `main`：`call sim_main; ret 0`（替代 `include/buildin.c`）。

### 4.8 驱动集成与调试

- `go run -tags codegen .`：打印各包 LLVM IR 文本（`mod.String()`）。
- 每包生成后 `mod.Verify()`；错误带包名 + `*llvm.Error` 结构化诊断。
- `compile`/`debug` stage 行为不变（产物 `main.out` 于 cwd；`debug` 跑 `examples/main.sim`）。

## 5. 文件结构变更清单

| 文件 | 动作 | 职责 |
|---|---|---|
| `go.mod` | 修改 | `go 1.27`；加 `github.com/kkkunny/go-llvm`（pseudo-version 固定） |
| `compiler/llgen/llgen.go` | 新建 | `CodeGenerator`：模块/Builder、三遍 Generate、Verify |
| `compiler/llgen/context.go` | 新建 | 共享 Context：llvm.Context、TargetMachine、符号表、typeCache、辅助函数缓存 |
| `compiler/llgen/type.go` | 新建 | HIR 类型 → LLVM 类型（§4.2）、union payload 计算 |
| `compiler/llgen/global.go` | 新建 | typedef 两遍、全局变量、函数/extern 声明、`main` wrapper |
| `compiler/llgen/expr.go` | 新建 | `genExpr`/`genAddr` 双通道（§4.3） |
| `compiler/llgen/local.go` | 新建 | let/return/if/while/for/block |
| `compiler/llgen/closure.go` | 新建 | 闭包 ctx、包装函数、调用、捕获（§4.4） |
| `compiler/llgen/equal.go` | 新建 | 相等性辅助函数合成（§4.6） |
| `compiler/llgen/other.go` | 新建 | `stableName` 等工具 |
| `llvmgen.go`（根，临时） | 新建 | `//go:build llvmgen` 开发用驱动：analyze → llgen → 打印 IR |
| `compiler/compile/compiler.go` | 重写 | 每包 EmitToFile、clang 链接（§4.7） |
| `compiler/compile/cache.go` | 修改 | 缓存有效性只看 `.o` |
| `codegen.go`（根） | 修改 | 打印 LLVM IR |
| `compiler/cir/`（整包） | **删除** | |
| `compiler/codegen/`（整包） | **删除** | |
| `include/buildin.h`、`include/buildin.c` | **删除** | 宏与 main 包装由 llgen 承担 |
| `compiler/config/filepath.go` | 修改 | 移除 `IncludePath` |
| `AGENTS.md`、`README.md` | 修改 | 架构描述、依赖、go-llvm 问题记录约定 |

## 6. TODO（功能 / 特性粒度）

### A. 基础设施

- [x] **A1 依赖升级**：`go.mod` → `go 1.27.0`，固定 `go-llvm` pseudo-version。
      验证：`go build ./...`；旧管线 `go run -tags compile . examples/main.sim && ./main.out` 输出 `123`。
- [x] **A2 llgen 骨架**：新包 + `Context`（llvm.Context/TargetMachine/dataLayout/符号表/typeCache）、
      `CodeGenerator`（模块/Builder）、三遍 `Generate()`、`mod.Verify()`；
      临时 `-tags llvmgen` 驱动打印 IR。
      验证：`go run -tags llvmgen . examples/main.sim` 输出合法（空）`.ll`，Verify 通过。
- [x] **A3 符号表与缓存**：`idents` 存符号描述符（名称/是否外部/局部存储地址）；
      `typeCache` 按 `*globals.TypeDef` 指针缓存（避免跨包重名冲突）+ named struct 按名去重；
      辅助函数缓存留待 G7。验证：`m1_puts`（std::c 的 puts 在 main 模块生成外部声明）。

### B. 类型系统映射（`type.go`）

- [x] **B1 标量**：i8..i64/u8..u64/f32/f64/bool（unit→void 仅返回位）。
      验证：i32/u8/i64/bool 局部（`m1_types`，IR 截断正确）；浮点字面量语言层（lexer）暂不支持，映射代码已就绪。
- [x] **B2 `str`**：named struct `{ptr, i64}`。验证：`m1_puts` 输出 hello，IR `%str = type { ptr, i64 }`。
- [x] **B3 引用/指针**：`&T`/`&mut T` → `ptr`。验证：取址/解引用（`m2_lvalue`）；
      指针比较待 D4；`type R &T` 一类引用别名待 B7 修正（`genCustomTypeDecl` 目前按聚合预声明）。
- [ ] **B4 tuple**：named struct `e1..en`。验证：构造 + 索引读写。
- [ ] **B5 数组**：`[N x T]`；N=0 / 零尺寸元素 → `{}`。验证：字面量、索引、整体赋值、传参。
- [ ] **B6 struct**：named struct；递归/互递归（`&Self`）走 opaque+`SetBody`。
      验证：`examples/main.sim`（`S{name}` + `&s` + 方法）。
- [ ] **B7 自定义类型**：标量别名透传；聚合 named struct（两遍 + `stableName`）。
      备注：`type R &T` 引用别名属于**后端 B7 问题**——`genCustomTypeDecl` 目前把 `RefType` 也
      预声明为 opaque named struct，应改为透传 `ptr`（详见 §11 修订记录/审查报告）。
      验证：`std/buildin` 的 `type i8 i8` 等 + `type S struct`。
- [ ] **B8 函数胖类型**：`{ptr fn, ptr ctx}`。验证：函数变量赋值/传递。
- [ ] **B9 tagged union**：`{i8, payload}` + DataLayout 精确计算。
      验证：union 注入/判别/取值，含 `{i8[16], f64}` 这类 size≠最大对齐成员的用例。
- [ ] **B10 零尺寸类型**：`{}` 语义（`zeroinitializer`）。验证：空 tuple/空 struct 值传递。

### C. 常量与全局（`global.go`）

- [x] **C1 数值字面量**：`ConstIntOfString`、`ConstFloat`、`ConstBool`。
      验证：i32 字面量作 extern 实参、u8/i64 局部初始化（`m1_types`）；边界值待统一测试集。
- [x] **C2 字符串字面量**：`ConstString` + `ConstGEP` + `ConstNamedStruct` 全局常量。
      验证：`m1_puts` 输出 hello，IR `@_str.1 = constant [6 x i8] c"hello\00"`。
- [ ] **C3 聚合字面量**：tuple/array/struct 常量路径（全局初始化）+ 运行时 `InsertValue` 路径。
      验证：全局/局部聚合初始化。
- [ ] **C4 全局变量**：定义（无值 → `SetInitializer(zeroinitializer)`）、`pub/static` 链接性、
      `@extern` 外部全局声明。验证：跨包读写 pub 全局 + extern 全局。
- [x] **C5 函数定义/声明**：`pub/static` 链接性、`@extern` 纯声明、`stableName` 命名。
      验证：`m1_putchar`（外部函数调用）、`m1_puts`（跨包声明与调用）。
- [x] **C6 入口**：`main` → `sim_main` + `main` wrapper（`call sim_main; ret 0`）。
      验证：三个 M1 用例均经 wrapper 正常运行。

### D. 表达式（`expr.go`）

- [x] **D1（部分）lvalue/rvalue 双通道**：`genAddr`/`genExpr` 拆分。验证：局部变量/解引用赋值（`m2_lvalue`）；
      数组/元组索引、字段左值（待 B4/B5/B6）已留分支并 panic。
- [x] **D2 标识符**：局部 alloca load、参数 alloca、全局 load、捕获字段 GEP。
      验证：局部读取（`m1_local` 输出 B）；全局读取待 C4、捕获待 F7。
- [x] **D3 算术/位运算/移位**：按符号性 `SDiv/UDiv/SRem/URem/FAdd.../frem`、`AShr/LShr`、`And/Or/Xor`。
      验证：有/无符号（`m2_ops`：udiv/ashr/lshr；`m2_ops_int2`：srem/urem/and/or/xor/shl）、
      浮点 `m2_ops_float`（fadd/fsub/fmul/fdiv/frem，无 libm）；typedef 解包待 M3 用例。
- [x] **D4（部分）比较**：`ICmp`/`FCmp` 谓词映射 → i1（有/无符号、有序浮点、bool、指针 Eq/Neq、
      `CustomType` 递归解包；**浮点 Neq 用 `UNE`，保 C `!=` 的 NaN 语义**）。验证：i64/i8/u8 比较与谓词 IR
      （`m2_cmp_paren`/`m2_edge`）、NaN 用例（`m2_nan`：`nan != nan` → true）；
      复合类型相等比较待 G1~G5（现为清晰 panic）。
- [x] **D5 一元运算**：`BitsReverse`→`Not`、`BooleanReverse`→`Xor true`、`GetRef`→genAddr、
      `DeRef`→load。验证：`m2_lvalue`（取址/解引用）、`m2_unary`（`xor i32 %v, -1` / `xor i1 %v, true`）。
- [x] **D6 赋值/复合赋值**：左值单次求值（`genAddr` 只调用一次）。验证：`*p = x + 1`、`x += 1`
      （`m2_lvalue`）；`a[f()] += x` 类副作用左值待 B5。
- [x] **D7 短路 `&&/||`**：分支块 + 合流（右侧惰性求值、最多一次）。验证：右侧带副作用的 `m2_shortcircuit`
      输出 `FTFT`，IR 中右侧调用只在 `logic.rhs*` 分支块内。
- [x] **D8 三元 `?:`**：分支块 + 合流（禁止 `Select`；未命中分支不执行）。验证：分支带副作用的
      `m2_sideeffect` 输出 `AABBABB`、unit 结果 `m2_ternary_void` 输出 `AB`。
- [x] **D9 数值转换** `NumberCovert`：`Trunc/ZExt/SExt/FPToSI/FPToUI/SIToFP/UIToFP/FPTrunc/FPExt`。
      验证：全组合转换表用例（`m2_convert`/`m2_convert2`/`m2_edge` IR）。
- [ ] **D10 TypedefCovert**：标量透传/转换（与 C 后端一致只支持标量层）——标量路径已随 D9 实现
      （聚合层待 B 系列，现为清晰 panic）。验证：`type MyInt i32` / `type MyBool bool` 转换与比较。
- [ ] **D11 索引**：数组 `GEP`、元组 `ExtractValue`/GEP；右值/左值两路径。验证：嵌套索引赋值。
- [ ] **D12 字段访问** `GetField`：struct/带 Self 指针自动解引用。验证：`examples/main.sim` 的 `self.name`。
- [ ] **D13 调用**：直接 `Call`、外部调用已实现并验证（`m1_putchar`/`m1_puts`）；
      函数值 `CallIndirect`（ctx 判空双分支 + 快速路径）待 F4。
- [x] **D14 自增语义** `SELFADD`：**不适用**（HIR 无自增节点：`Unary` 仅 BitsReverse/BooleanReverse/
      GetRef/DeRef，`SELFADD` 只出现在旧 `cir` 的 for 动作位，随 E5 for 循环再评估）。验证：for 计数循环。

### E. 语句（`local.go`）

- [x] **E1 `let`**：入口块 alloca + store。验证：`m1_local`（`let x: i32 = 66`）。
- [x] **E2 `return`**：`RetVoid`/`Ret` + 函数体兜底（void→`ret void`，非 void→`unreachable`）。
      验证：空函数体 `m1_empty` 正常运行。
- [x] **E3 `if/else-if/else`**：块结构递归。验证：嵌套 if 链。
- [x] **E4 `while`**：cond→body→cond 块环。验证：条件副作用/出口。
- [ ] **E5 `for`（range 数组）**：索引变量 + 数组遍历。验证：遍历求和。
- [x] **E6 嵌套 block 作用域**。验证：内层 let 遮蔽外层。

### F. 函数与闭包（`closure.go`）

- [ ] **F1 函数体生成**：参数 alloca、入口块、签名（`ctx.Fn`）。验证：多参函数。
- [ ] **F2 闭包构造**：捕获收集 → ctx named struct → 静态包装函数（首参 ctx）。验证：捕获外部 let。
- [ ] **F3 闭包值**：`{fn, null}` / `{fn, &ctx}` 构造。验证：函数字面量赋值给变量。
- [ ] **F4 闭包调用**：ctx 判空双 `CallIndirect` + 合流；已知形态直接调用。
      验证：同一变量先后装纯函数与闭包再调用。
- [ ] **F5 函数值相等**：fn+ctx 双比较。验证：函数变量判等。
- [ ] **F6 方法绑定** `GetBind`：HIR 层包装 + F2/F3。验证：`(*ss).getname()`。
- [ ] **F7 捕获变量访问**：`captureVarsMap` → ctx GEP。验证：多层嵌套闭包捕获。

### G. 相等性辅助函数（`equal.go`）

- [ ] **G1 标量/布尔/引用相等**：直接 `ICmp`。验证：`==/!=` 全标量。
- [ ] **G2 数组逐元素**。验证：数组判等/不等。
- [ ] **G3 元组逐字段**。验证：嵌套元组判等。
- [ ] **G4 struct 逐字段**。验证：`S{name} == S{name}`。
- [ ] **G5 union 判别 + `Switch` 分派**。验证：同/异 tag、同 tag 异值。
- [ ] **G6 函数值相等**（并入 F5，独立可测）。
- [ ] **G7 辅助函数去重缓存**：同类型只生成一次。验证：IR 文本中只出现一个 eq 函数。

### H. 跨包、产物与链接（`compiler/compile`）

- [x] **H1 每包一模块 + 外部声明**：按需创建 declaration。
      验证：`m1_puts`（main 模块内生成 `declare void @puts(%str)`）；examples 全量待 M3。
- [ ] **H2 `EmitToFile`** 产出 `.sim_cache/<pkg>.o`（临时驱动已产出到临时目录；正式接入见 H5）。
- [x] **H3 clang 链接**：`clang main.o dep.o... -lm -o main.out`。验证：M1 用例运行输出正确（临时驱动）。
- [ ] **H4 缓存**：有效性只比 `.o`/源 mtime；文件锁保留。验证：连续两次 compile，第二次命中缓存。
- [ ] **H5 compile 主/依赖包流程改造**：删除 `.h` 输出与 `#include` 拼接。验证：全量编译 + 缓存命中。

### I. 驱动集成与调试

- [ ] **I1 codegen stage 打印 IR**：主包 + 依赖包都打印。验证：`-tags codegen` 输出合法 `.ll`。
- [ ] **I2 Verify 诊断**：错误带包名与 `*llvm.Error` 信息。验证：人为制造的 IR 错误信息可读。
- [ ] **I3 compile/debug stage 适配**：行为不变。验证：`-tags debug .` 退出码/输出。
- [ ] **I4 正式测试集**：**不做**（用户后续统一添加）。每个 TODO 的"验证"列即临时验证方式。

### J. 清理与文档

- [ ] **J1 删除 `compiler/cir` 与旧 `compiler/codegen`**，`llgen` 改名搬回 `compiler/codegen`。
      验证：`go build ./...` + `grep -r cir compiler/` 为空。
- [ ] **J2 删除 `include/buildin.h/.c`**、清理 `config.IncludePath`。验证：全量编译不受影响。
- [ ] **J3 文档**：AGENTS.md（架构、依赖、go-llvm 问题约定）、README 依赖说明。
      验证：文档命令实测一致。

## 7. 验证策略

1. **每功能简单验证**：最小 `.sim` 片段（`/tmp/opencode/sim-cases/`）+ 编译运行 + 输出/退出码检查。
2. **IR 级验证**：`mod.Verify()` 内建于每包生成；`-tags llvmgen` / `-tags codegen` 打印 IR 文本。
3. **回归锚点**：M3 起 `examples/main.sim` 必须始终通过。
4. 正式测试集由用户后续统一补充（见 §6 I4）。

## 8. 风险与注意点

| 风险 | 对策 |
|---|---|
| go-llvm 迭代快、可能不稳定 | pseudo-version 固定；问题记录到 `docs/go-llvm-issues.md`（见 §0.2） |
| LLVM 22 严格 IR 校验 | Verify 前置于出 `.o`；错误带包名 |
| C ABI 细节（struct 按值传参跨 `@extern`） | `str` 已验证；用 extern 复合参数用例复核 |
| union payload 对齐 | DataLayout 精确计算（§4.5）+ 专项用例 |
| 无初始化全局语义差异 | C4 明确 `zeroinitializer` 定义 vs extern 声明 |
| go-llvm 生命周期（Close / `panic(*llvm.Error)`） | `defer Close()`；panic 风格与 `stlerr.Must*` 一致 |
| 短路/三元误译成 `Select` 双求值 | D7/D8 禁止 + 副作用断言 |
| 复合赋值左值双求值 | D6 addr 单次求值 + 专项用例 |
| 缓存未含编译器版本（现状即如此） | 维持现状，后续可把 `config.ABIVersion` 纳入缓存键（非本次范围） |

## 9. 实施顺序与里程碑

依赖顺序 **A → B → C → D/E → F → G → H → I → J**。

1. **M1 端到端最小闭环**（A + B1/B2 + C1/C2/C5/C6 + E2 + H）：`main` 返回常量、`puts("hello")`
   经新后端跑通，链接/缓存/Verify 全链路就位。
2. **M2 标量语言**（B1 + D1~D9/D14 + E1~E6）：算术/控制流/转换。
3. **M3 复合类型**（B3~B7/B10 + C3/C4 + D10~D12 + G1~G4）：`examples/main.sim` 通过。
4. **M4 函数值与闭包**（B8 + D13 + F1~F7 + G6）。
5. **M5 union 与完整相等性**（B9 + G5/G7）。
6. **M6 收尾**（I + J）：`examples/main.sim` 回归 + 旧代码删除 + 文档更新。

## 10. 进度记录

- **2026-09-26**：A1/A2 完成（go-llvm 引入、llgen 骨架、`llvmgen` 驱动）。
- **2026-09-26**：**M1 核心闭环打通**（临时驱动 `llvmcompile`）——
  空 `main`、`putchar(65)`→"A"、`import std::c` + `puts("hello")`、局部变量读取（→"B"）、
  u8/i64/bool 类型映射（i8 截断 -56 正确），全部经 LLVM IR → `.o` → clang 链接运行验证；
  对应任务 A3/B1/B2/C1/C2/C5/C6/D2/E1/E2/H1/H3。
  下一步：接入 `compiler/compile`（H2/H4/H5）或按 B 系列补齐复合类型（推荐先 B 系列，
  以便 `examples/main.sim` 尽早回归）。
- **2026-09-26**：**M2-1 完成**（值模型/引用/算术/一元/赋值）——`genAddr`/`genExpr` 双通道
  （可寻址：局部变量、解引用；索引/字段待 B4/B5/B6）、`&T`→`ptr`、算术/位运算/移位按符号性分派
  （浮点 `FRem` 不调 libm）、`!`→`Not`/`Xor true`、赋值与复合赋值（左值地址单次求值）；
  IR 级验证 udiv/ashr/lshr/srem/urem/fadd…frem；旧管线 `examples/main.sim` 回归 `123`；
  对应任务 B3/D1(部分)/D3/D5/D6。
  注意：`*p = x + 1` 需 `let mut p`（分析器 `DeRef.Mutable` 取的是绑定可变性而非引用目标，
  旧后端同样被拒，属前端语义问题，非 llgen bug）。
- **2026-09-26**：**M2-2 完成**（比较/短路/三元/数值转换）——D4 `ICmp`/`FCmp` 谓词映射
  （有/无符号、有序浮点、bool i1、指针 Eq/Neq、`CustomType` 递归解包；聚合相等待 G1~G5 并清晰 panic）、
  D7 `&&`/`||` 与 D8 `?:` 分支块 + 入口块临时 alloca 合流（右侧/未命中分支惰性求值：
  `m2_shortcircuit`→`FTFT`、`m2_sideeffect`→`AABBABB`、`m2_ternary_void`→`AB`，
  IR 中右侧调用只出现在分支块内）、D9 全组合转换（trunc/zext/sext/sitofp/uitofp/fptosi/fptoui/
  fptrunc/fpext）与 D10 标量 typedef 透传；旧管线 `examples/main.sim` 回归 `123`。
  发现前端问题（非本任务范围，旧 C 后端同样复现）：
  (1) parser 缺陷：`parseSuffixExpr` 在二元运算符循环之前消费 `?`，故 `1 < 2 ? a : b` 被解析为
  `1 < (2 ? a : b)`，验收用例需给条件加括号，建议后续单独修 parser；
  (2) `std/buildin` 的 `type bool bool` 使 `bool` 名解析为 `_CustomBooleanType`（`String()` 也是
  "bool"），`let b = mb as bool; cond ? ...` 处 `expectTypeExpr(cond, types.Bool)` 因
  `_CustomBooleanType.Equal(_BooleanType)==false` 报 "expected 'bool' but got 'bool'"。
- **2026-09-26**：**M2-2 审查修复**——浮点 `Neq` 由 `ONE`（有序且不等，NaN 时返回 false）改为
  `UNE`（无序或不等），与 C `a != b`（任一操作数为 NaN 时返回 true）语义一致；
  `m2_nan` 用例修复前输出 `BBB`、修复后 `ABB`，IR 为 `fcmp une`。
  详见 `.superpowers/sdd/task-M2-2-report.md` 的审查修复附录；复审通过（谓词修复无回归，
  其余 O 谓词与惰性求值结构未受影响）。
  M2-2 遗留的「unit 三元值被消费会生成 `ret <br>` 伪值」问题已在最终审查修复中一并解决
  （`genReturn` 按表达式类型是否为 unit 选择 `RetVoid`，见文末最终审查修复条目）。
- **2026-09-26**：**M2-3 完成**（控制流：if/else-if/else、while、嵌套块）——E3 `genIf`：
  条件求值一次 → `CondBr(then, else)` → 各分支跳转**共享**合流块；else-if 递归（条件在 else
  块内求值），无 else 时空 else 块直接跳合流。若所有分支均以终结指令（return）结束，合流块
  无前驱，补 `unreachable` 终结指令保证 IR 合法并保持 `terminated=true`（后续语句经
  `ensureBlock` 进死块）。E4 `genWhile`：`Br(cond)`→`CondBr(body, end)`→body→`Br(cond)`
  块环；循环体终结时不回跳；出口块始终有前驱（条件分支），`MoveToEnd(end)` 后继续发射。
  E6 嵌套块/遮蔽由递归 + `idents` 按 `hir.Ident` 指针索引天然支持。
  验证：`m2_control_for` 输出 `012!@AC`、多出口函数 `m2_multi_exit` 输出 `A.B`、
  全分支终结 `m2_all_exit` 输出 `ABAB`（死代码块不可达且合法）、
  遮蔽/嵌套 `m2_scope` 输出 `BACDACD4`、无条件循环+条件副作用 `m2_loop_edge` 输出 `C...C`；
  同批用例旧 C 后端输出逐字对齐；M1/M2 全部既有用例双后端对照无回归
  （仅 `m2_nan` 的 ABB 与 `m2_ops_float` 的可编译属 M2-2 已记录的既有差异）；
  旧管线 `examples/main.sim` 回归 `123`；build/vet/gofmt 通过。
  附带修复：`let` 的 alloca 移到函数入口块（§4.3 值模型）——原实现放在当前块，
  循环体内 `let` 每轮 alloca，1M 次迭代即栈溢出 SIGSEGV（C 后端 50M 正常）；
  D14 检查：`compiler/hir/locals/expr.go` 无自增表达式节点（`Unary` 仅
  BitsReverse/BooleanReverse/GetRef/DeRef），旧 `cir.UnaryOpEnum.SelfAdd` 只出现在
  for 动作位，属 E5 结构，跳过。
  发现前端问题（旧后端同样复现，非本任务范围）：(a) `while` 不是语言关键字，
  While 语句须写 `for <cond> {}`（parser `parseFor` → `ast.While`）；
  (b) 裸 `{}` 块不创建作用域（`analyzeBlock` 对 `*ast.Block` 直接展开，无 `NewBlockScope`），
  if/while/for/函数体才创建块作用域。
- **2026-09-26**：**最终宽范围审查修复**（range `81becee..8c0ba62`）——三条关键问题：
  (1) `return <unit 表达式>` 生成非法 IR：`genReturn` 改为先求值（保副作用）再按
  `types.GetUnderlying(v.GetType())` 是否为 `UnitType` 选择 `RetVoid`/`Ret`，覆盖 unit call
  伪值与 `genTernaryVoid` 的 `br` 伪值（`fix_unitret`/`fix_unittern` 输出 `G`）；
  `genLocalLet` 对 unit 类型前置检查并给出含变量名/类型的清晰 panic。
  (2) 字符串常量跨包重定义：`@_str.N` 全局设 `LinkagePrivate`（每模块 `@_str.1` 互不冲突），
  并因内部/私有全局在静态重定位模型下被后端用 `R_X86_64_32` 绝对寻址、无法链入 clang 默认
  PIE，`NewContext` 的 TargetMachine 重定位模型由 `RelocDefault` 改为 `RelocPIC`；
  两模块字符串用例链接成功（输出 `main`/`pkg`），IR 为 `@_str.1 = private constant ...`。
  (3) 同包前向引用：`Generate()` 在生成函数体前新增 `genFuncDecls` 函数符号声明子遍
  （统一命名/链接性规则，含 `@extern` 声明），调用点不再依赖声明顺序
  （`fix_forward`：helper 定义在 main 之后 + 互递归，输出 `A1`）；
  `genCall` 未登记符号的 panic 文案改为明确提示。
  回归：M1/M2 全部既有用例双后端对照无新增差异（旧 C 后端在 `rev_forward`/`fix_forward`
  上前向引用直接 SIGSEGV，llgen 现已支持）；旧管线 `examples/main.sim` 输出 `123`；
  build/vet/gofmt 通过。审查报告见 `.superpowers/sdd/final-fix-report.md`。

## 11. 迁移期间发现的前端问题（非后端迁移范围，待单独处理）

1. **parser 三元优先级**：`parseSuffixExpr` 在二元运算符循环之前消费 `?`，`1 < 2 ? a : b`
   被解析为 `1 < (2 ? a : b)`；旧 C 后端同样复现。临时对策：条件加括号。
2. **`DeRef` 可变性检查方向**：`let p = &mut x; *p = v` 被 analyze 以 "must mutable" 拒绝，
   检查的是绑定 `p` 的可变性而非引用目标；`let mut p` 可绕过。旧 C 后端同样拒绝。
3. **`type bool bool` 与内建 `types.Bool` 不 Equal**：`let b = mb as bool; cond ? ...` 报
   "expected 'bool' but got 'bool'"（`_CustomBooleanType.Equal(_BooleanType)==false`）。
4. **`analyzeBinary` 可变性检查漏 `MulAssign`**：`compiler/analyze/expr.go:316-327` 的赋值运算符
   可变性检查列表含 `Assign`/`AddAssign`/`SubAssign`/`QuoAssign`/`RemAssign`/`AndAssign`/`OrAssign`/
   `XorAssign`/`ShlAssign`/`ShrAssign`，唯独漏了 `MulAssign`（`*=`）。因此不可变变量/临时值的
   `x *= v` 不被拒绝（实测：`let x: i32 = 2; x *= 3` 通过 analyze，而 `x += 3` 报 must mutable）；
   旧 C 后端同样复现。
   （原第 4 条 `type R &T` 引用别名经复核属后端 B7 问题，已移入 §6 B7 备注。）
5. **无值 `return` 在非 unit 函数中不被 analyze 拒绝**：`let f = () -> i32 { return; }` 通过 analyze，
   生成 llgen IR 时由 `Module.Verify` 报
   `Function return type does not match operand type of return inst! ret void / i32`（难懂）。
   llgen 已在 `genReturn` 无值分支加防御：当前函数 LLVM 返回类型非 `void` 时 panic
   `llgen: 非 unit 函数 %s 不能使用空 return（前端漏校验）`。旧 C 后端同样坏（生成非法 C）。
   根治应在前端 analyze 的 `analyzeReturn` else 分支校验函数返回类型。
6. **unit 局部变量被 analyze 放行**：`let x = g()`（`g` 返回 unit）通过 analyze，旧 C 后端与
   llgen 都无法为 unit 分配存储。llgen 已在 `genLocalLet` 前置检查并 panic
   `llgen: 不能为 unit 类型的变量 %s 分配存储（类型 %s）`。根治应在前端拒绝 unit 型 `let`。

> **给 H5 的备注（旧 `compile` 缓存/命名隐患，重写时一并处理）**：
> 1. `compiler/codegen/other.go` 的 `stableName` 把绝对 `pkg.Path` 与 `pkg.Name` 一起哈希进符号名，
>    编译产物（`.sim_cache/*.o`）因此绑定绝对路径，换目录/迁移工作区后无法复用，缓存不可迁移；
> 2. `compiler/compile/cache.go` 的 `isCacheValid` 对 `stlerr.ErrorWith(os.Stat(...))` 包装后的
>    not-exist 错误仍用 `os.IsNotExist` 判断（不会解包），缓存文件缺失时按真实错误上抛、最终
>    在 `Compiler.Visit` panic，而非按缓存未命中重编；应改用 `errors.Is(err, fs.ErrNotExist)`；
> 3. `compiler/compile/compiler.go:121` 的 `filepath.Rel(config.IncludePath, config.SimRootPath)`
>    参数顺序反了（应为 `Rel(SimRootPath, IncludePath)`），会生成 `#include "../buildin.h"`
>    而非预期的 `#include "include/buildin.h"`，属潜在隐患。
