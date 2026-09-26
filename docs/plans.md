# Sim 未来规划与计划

本文件是**未来要做的事**的唯一清单：功能与语言缺口、性能优化、架构调整、重构、技术债、
测试与工程能力（未修复**缺陷（bug）**见 [`docs/known-issues.md`](known-issues.md)；
go-llvm 库问题见 [`docs/go-llvm-issues.md`](go-llvm-issues.md)）。

**边界**：本文件只收「要做什么」——即使现状是「还不能用」，只要不属于错误行为（功能缺口、
设计取舍、优化、重构、技术债）就归这里；「做错了、要改对」的归 known-issues。

**维护规则**

- **单项完成即从本文件删除**（不保留已完成条目；历史见 git log）。
- 其他任务过程中发现的、未来需要完成的事项，补录到对应分类。
- 条目包含：目标 / 现状与动机 / 验收方向（足够开工即可）。
- 编号一经分配不复用、不重排，新事项顺延。

---

## 一、性能与构建

### P1. 增量编译：缓存命中不再全量 codegen

- **现状**：`compileDepPkg` 即使缓存有效也调用 `gen.Generate()`（为填充共享
  `codegen.Context` 的 idents/typeCache），缓存只省下 `EmitToFile`。
- **目标**：把「Context 填充」与「IR 发射」解耦——持久化接口信息（类型布局/符号表），
  缓存命中时不跑 codegen。
- **验收**：无改动重复编译时依赖包零 codegen；`examples/main.sim` 与大用例结果不变。

### P2. 优化管线与 `-O` 级别

- **现状**：`TargetMachine` 固定 `target.OptNone`，模块无 pass、无 LTO，产物恒为 -O0。
- **目标**：至少 `OptDefault` + 可选 `-O0/-O2`；启用 mem2reg 等必要 function pass。
- **验收**：产物有优化痕迹、语义不变；go-llvm 缺 pass API 时记录到 `go-llvm-issues.md`。

### P3. 包级并行编译

- **现状**：`Dagger.OrderedWalk` 串行访问；DAG 上独立包本可并行，但 `codegen.Context` 的
  `idents/typeCache/eqFuncs` 是共享 map（非线程安全）。
- **目标**：按包隔离 Context（或分片加锁）后并行 EmitToFile。
- **验收**：多包工程编译时间显著下降，产物与串行一致。

### P4. 字符串常量去重

- **现状**：`genString` 每处字面量新建 private global，同模块相同字面量重复占空间。
- **目标**：同模块内按字面量缓存复用同一 global。

### P5. std 的 HIR 序列化缓存（长期）

- **现状**：每次编译都重新解析/分析 `std`；buildin 很小可接受，规模扩大后浪费明显。
- **目标**：HIR 序列化缓存 + 失效校验（接口版本/源码 mtime）。

---

## 二、架构与设计

### P6. 编译根路径不依赖 CWD

- **现状**：`config.SimRootPath = os.Getwd()`——只有从仓库根启动才能找到 `std/buildin`
  （否则报找不到目录），`stableName` 的包路径哈希也随之受 CWD 影响。
- **目标**：根路径来源显式化（环境变量 / `-root` flag / 可执行文件位置），与 CWD 解耦。
- **验收**：从任意目录编译同一个 `.sim` 成功，且符号名/缓存跨目录一致。

### P7. 编译入口显式化

- **现状**：包名 == `"main"` 才触发入口包装与链接；非 main 目录只产 `.o` 不链接也无提示，
  单文件包又恒为 `"main"`。
- **目标**：显式 entry 规则（固定 `main.sim` / `@entry` 属性 / CLI 指定），错误使用给出诊断。
- **验收**：`examples/` 目录名与入口无关；无入口时明确报错。

### P8. bind（方法）在 HIR lowering 统一 desugar

- **现状**：bind 语义散落在 `analyzeMember`、`GetBind.IsStatic`、`codegen.genGetBind` 三处，
  且 `genGetBind` 在 codegen 期动态构造 HIR（`locals.NewFunc/NewCall`）绕过 analyze 校验。
- **目标**：analyze/lowering 阶段统一 desugar 为「函数字面量 + 调用」，codegen 只处理常规闭包。
- **验收**：bind 相关用例（含 `examples/main.sim`）行为不变。

### P9. 后序遍历统一

- **现状**：根驱动 `codegen.go` 的 `collectPkgs` 与 `compile/compiler.go` 的 `collectDeps`
  是两套等价的后序遍历实现。
- **目标**：合并为一个共享工具（`compiler/compile` 或 `util`）。

### P10. HIR 打印顺序稳定

- **现状**：`locals.Struct.Fields` 是 map，HIR 打印字段顺序随 map 迭代漂移。
- **目标**：打印按字段名/声明序排序；为后续 golden 测试提供稳定输出。
  （codegen 已按类型字段序处理，无正确性问题。）

---

## 三、工程与质量

### P11. 测试体系

- **现状**：仓库无 `*_test.go`；验证依赖手工编译 `.sim` 片段，known-issues 的复现用例易失守。
- **目标**：先为 lex/parse/report 建 golden 测试；把 known-issues 各条最小复现收进回归脚本
  （可在 CI 用一个脚本批量编译/运行并断言输出）。

### P12. 代码卫生清理

- `compiler/parse/expr.go` 死类型 `type A struct { a int }` 删除。
- `codegen.Ident.ExternalFunc` 死字段删除（判定已改用 `FuncSymbol`）。
- `codegen.Context.idents` 跨函数局部符号残留（按 `hir.Ident` 指针唯一，无正确性问题，
  随程序规模增长；可随包隔离 Context 一并处理）。
- `hir/types/*.go` 的 `WriteFormat(t.String())` 改 `WriteString`（`go vet` 告警）。
- README 的 `make run` 用法与 `Makefile` 已过时（引用不存在的 `cmd/`、`runtime/`、`tests/`），
  同步或删除。
- `compiler/compile/compiler.go` 清理旧 C 后端头文件的 `os.Remove(*.h)` 可在下次 ABI 升级时移除。

---

## 四、语言与运行时（长期）

### P13. 逃逸闭包支持（堆分配 ctx）

- **现状**：闭包 ctx 在定义点栈分配（入口块 alloca），捕获局部变量的闭包逃逸出定义帧后再调用
  属未定义行为（实测返回悬垂垃圾）；同一闭包字面量在循环内多次求值会复用同一 ctx 槽位，
  跨迭代保留多个实例会互相串写。安全范围：定义帧内使用（GetBind、闭包作实参、返回无捕获闭包
  `{fn,null}`）。
- **目标**：ctx 堆分配 + 生命周期管理；语言层需要所有权/GC 设计（自动内存管理是 README
  宣称的目标之一）。
- **验收**：闭包作为返回值逃逸后调用结果正确，无悬垂。

### P14. 注释语法

- **现状**：词法器 `skipWhite` 只跳空格/制表/回车；`//`、`/* */` 都会被当作 token 导致解析错误，
  `.sim` 源码无法写注释（原 known-issues F20）。
- **目标**：支持行注释 `//` 与块注释 `/* */`（是否嵌套待定）；std/examples 补注释。
- **验收**：注释可出现在任意空白位置，诊断位置/行列不受影响。

### P15. 浮点字面量语法

- **现状**：词法器 `scanInteger` 只扫描十进制数字，token/AST 均无浮点字面量；
  `1.5` 被切成 `1` `.` `5` 产生解析错误（清理 F15 时发现）。f32/f64 目前只能由整数字面量转换生成。
- **目标**：支持 `1.5`、`1e10` 等浮点字面量，与整数字面量同走期望类型适配（含负数折叠路径）。
- **验收**：`let x: f64 = 1.5` 等用例通过，std 与大用例结果不变。

### P16. 全局常量折叠与初始化器覆盖面

- **现状**：`genConstExpr` 只支持字面量与其常量嵌套；`let g: i32 = 1 + 1`、
  `let g: i64 = 65 as i64`、union 注入等全局初始化器不被支持（对应 ICE 消息不可读的**缺陷**
  见 known-issues B2）。旧 C 后端把这些交给 C 编译器可编。
- **目标**：支持常见常量表达式（算术/比较/转换/聚合嵌套）的折叠，优先在 analyze/lowering 期
  折叠为字面量。
- **验收**：上述形态可初始化全局；`examples/main.sim` 与大用例结果不变。

### P17. for-in 快照语义明确化

- **现状**：`for x in *getp()` 这类以指针解引用为 range 的循环，codegen 只求一次指针、逐轮读活内存；
  旧后端因 `Temporary()` 会把数组值拷进临时变量（快照）。语言语义目前未规定该情形
  （值 range 本身已在入口块物化，不受影响）。
- **目标**：明确「值快照 / 内存视图」语义，并决定是否对齐旧行为。
- **验收**：语义结论固定并有对应用例（循环体内修改同一内存时的行为可预期）。
