# go-llvm 使用问题记录

在将 Sim 编译器后端从 C 重写为 LLVM（基于 `github.com/kkkunny/go-llvm`，版本
`v0.0.0-20260926104146-5b02948e8001`）的过程中遇到的问题。按严重程度排序。

> 更新：以下各项的解决方法均已在 go-llvm 本地完成（尚未发布；使用时需把依赖更新为包含
> 修复的版本，或 `replace` 到本地 checkout）。

## 1. `ConstStruct` 不校验操作数是否为常量，错误信息严重误导

构造函数胖值时，如果闭包上下文是 `alloca` 指令却误用了 `Context.ConstStruct`：

```go
pair := ctx.ConstStruct(false, fn, ctxv) // ctxv 是 alloca 指令
```

go-llvm 不会报错，会构造出一个"常量里包含指令"的畸形值。最终 `Module.Verify()`
报错：

```
Use of instruction is not an instruction!
      %5 = alloca { %_Sim_..._S }, align 8
```

错误指向的是那条 `alloca`，而真正的问题在 `ConstStruct` 的调用处，定位成本很高
（LLVM 外部工具 `opt -passes=verify` 反而通过，因为打印出的文本里常量无法携带
指令，打印后看起来是合法的）。建议在 `ConstStruct`/`ConstArray`/
`ConstNamedStruct` 等构造函数中对每个操作数做 `IsConstant()` 校验，panic 时带上
调用方符号；或者在 `llvm.Value` 上提供 `MustBeConstant()`。本次的规避办法：
自己判断 `ctxv.IsConstant()`，非常量时改用 `InsertValue` 指令构造聚合值。

**解决方法（已实现）**：所有接收值的常量构造器都做了恒定（不依赖构建 tag）
的 `IsConstant` 校验：`ConstArray`、`ConstVector`、`ConstStruct`、
`ConstNamedStruct`、`ConstGEP`（base 与各索引）、`ConstIntToPtr`，以及
`IntConst` 的全部折叠运算（`Add`/`Sub`/`NSWAdd`/... 的右操作数）。非常量会当场
`panic(*llvm.Error)`（`ErrInvalidArg`，可经 `llvm.Catch`/`llvm.Try` 捕获），消息
含 `op`、操作数下标、值名与值 IR 文本，例如：

```
llvm.Context.ConstStruct: operand 1 @ctxv is not a constant: %5 = alloca { %_Sim_..._S }, align 8
```

不再出现"常量里包含指令"的畸形值，也就不会再有指向 `alloca` 的误导性 Verify
报错。调用方无需再自己 `IsConstant()` 判断（正常路径照旧用 `InsertValue` 即可）。

## 2. 动态类型代码中的 `Kind` 泛型样板过多

Sim 后端是动态生成器：类型在运行时才确定，几乎所有值都是 `llvm.AnyValue` /
`llvm.Value[DynT]`。但 `Builder` 的多数 API 带编译期 `Kind` 参数：

* `Load[U](p, t TypeRef[U], name)` 必须传入编译期已知的 `Type[U]`；
* `Add/Sub/...` 只接受 `ValueRef[IntT]`；
* `Call[U]` 的 `fn` 必须是 `ValueRef[FnT]`；
* `ExtractValue`/`PHI`/`InsertValue` 结果都是 `DynT`，每个使用点都要
  `MustAs[IntT]()` / `MustAs[PtrT]()` 转换。

结果是后端代码里充斥着 `mustInt`/`mustPtr`/`mustFloat` 包装，且 `MustAs` 失败时
的 panic 只有类型不匹配信息、没有当前操作的上下文。建议：

* 提供 `Value.AsInt()/AsPtr()/AsFloat()` 之类的便捷（或 `Must*`）方法；
* 提供动态重载（如 `LoadAny(p, t AnyType)` 返回 `Value[DynT]`），
  或者允许 `Value[DynT]` 直接作为 `ValueRef[任意 Kind]` 传入并在内部延迟校验；
* 至少在 `MustAs` 的 panic 中带上值字符串，便于定位。

**解决方法（已实现）**：

* `Value.As[U]()` 失败时返回的 `*llvm.Error` 现在附带值的**名字与 IR 文本**，
  `MustAs` 的 panic 自动继承，例如
  `llvm.Value.As: value kind mismatch: have pointer, want int; value @ctxv is %5 = alloca ...`，
  动态代码里可以直接定位到转换点。
* 便捷转换（`MustInt`/`MustPtr` 等薄封装）与 `LoadAny` 等动态重载**有意不做**：
  `MustAs[IntT]()` 已是等价写法，且泛型约束是编译期种类安全的核心，放宽等于破坏
  该设计。动态取类型可传 `t.DynType()` 得到 `Load[DynT]` 等擦除结果。

## 3. 缺少 Builder 插入点的保存/恢复 API

`Builder.CurrentBlock()` 只能拿到当前块，没有公开的"取出插入点句柄再恢复"的
能力。生成嵌套函数（闭包、相等性辅助函数、全局构造器）时必须手工保存
`CurrentBlock()`、切到新函数入口、生成完再 `MoveToEnd(旧块)`；如果进入嵌套前
构建器尚未定位（比如正在生成全局量），还没有"恢复成未定位"的表达。
建议提供 `Builder.SaveInsertPoint() / RestoreInsertPoint()`（或返回不透明句柄）。

**解决方法（已实现）**：新增 `ir.Builder.SaveInsertPoint() ir.InsertPoint` 与
`ir.Builder.RestoreInsertPoint(p ir.InsertPoint)`，快照三态齐全：

* 块末尾（`MoveToEnd` 后）：恢复时重新 `PositionBuilderAtEnd`；
* 指令之前（`MoveBefore` 后）：Builder 内部新增跟踪 before 指令，恢复时精确
  `PositionBuilderBefore`，不会退化成"块末尾"；
* 未定位（新 Builder 尚未定位，如正在生成全局量时）：恢复为未定位，底层新增
  `LLVMClearInsertionPosition` 封装。

快照跨 Context、或指令已不在快照的块中时会 panic（`ErrCrossContext` /
`ErrInvalidArg`），不会把非法位置交给 LLVM。嵌套闭包/相等性辅助/全局构造器的
生成可以直接 `p := b.SaveInsertPoint(); ...; b.RestoreInsertPoint(p)`。

## 4. 同名 `NewFunction` / `NewGlobal` 静默改名，没有 GetOrCreate

`Module.NewFunction("foo", ...)` 在 `@foo` 已存在时会悄悄创建 `@foo.1`；
`NewGlobal` 同理。跨模块/前向引用场景里必须自己先 `GetFunction`/`GetGlobal`
再决定创建。建议提供 `GetOrCreateFunction`/`GetOrCreateGlobal` 或在重名时给出
可捕获的错误（尤其是调试构建）。

**解决方法（已实现）**：新增

```go
func (m *Module) GetOrCreateFunction(name string, t llvm.FnType) (Function, bool)
func (m *Module) GetOrCreateGlobal(name string, t llvm.AnyType) (Global, bool)
```

bool 表示是否新建；已存在但签名/内容类型不一致时 panic `ErrTypeMismatch`（可
`Catch`），不会静默复用错误签名的符号。`NewFunction`/`NewGlobal` 的文档已明确
写出"同名静默改名 `name.1`"的行为（实测 LLVM 22 也确实改名，含不同签名）。
本次 LLVM 22 实测：同签名重名得到 `@f.1`，不同签名得到 `@f.2`。

## 5. `StructType.SetBody` 二次调用直接触发 LLVM abort

命名结构体只能 `SetBody` 一次，二次调用是断言/abort（无 Go error）。当多个包
共享同一个 `llvm.Context`（本项目的 `llgen.Context` 有意跨包共享类型）时，
类型定义去重完全靠调用方。建议 `SetBody` 在已定义时可返回错误或幂等（相同
body 时）。

**解决方法（已实现）**：`StructType.SetBody` 在结构体已定义（非 opaque）时直接
`panic(*llvm.Error)`（`ErrInvalidArg`，"struct X body is already set"，可经
`llvm.Catch` 捕获），不再触发 LLVM 的 `report_fatal_error`（该 abort 在 release
构建下也会杀死进程）。语义取严格版：即使 body 相同也 panic，跨包类型去重仍由
调用方负责（先 `StructType.IsOpaque()` 判断，或在自己的类型注册表里去重）。

## 6. 常量指针/位转换 API 未导出

`internal/binding` 里已有 `LLVMConstPointerCast`、`LLVMConstBitCast`，但
`llvm` 包没有对应的公开封装（`Context.ConstIntToPtr` 有但不够）。本次因为
`Function.Type()` 返回 `ptr`、函数值可直接放进 `ptr` 字段而绕开了，但做全局
常量重解释（如常量联合体）时会需要。

**解决方法（已实现）**：`llvm` 包新增两个公开封装（均带常量操作数校验）：

```go
func (ctx *Context) ConstBitCast[U Kind](v AnyValue, to TypeRef[U]) Value[U]
func (ctx *Context) ConstPointerCast(v ValueRef[PtrT], to PtrType) Value[PtrT]
```

`ConstBitCast` 的返回种类随目标类型推断（如 `ctx.ConstBitCast(x, ptrTy)` 得
`Value[PtrT]`）。其余 cast（addrspacecast、ptrtoint 等）按"按需添加"策略暂未
导出，需要时再补。

## 7. `pass.RunPasses` 不接受 `TargetMachine`

`pass.RunPasses`/`AutoOpt` 内部固定传空的 `LLVMTargetMachineRef`，无法让
`default<O2>` 管线使用目标机器（这会损失目标相关优化，也影响
`opt` 的 `-mcpu` 行为）。建议增加带 `*target.TargetMachine` 的重载。

**解决方法（已实现）**：`pass` 新增选项

```go
func pass.WithTargetMachine(tm *target.TargetMachine) Option
```

`RunPasses`/`RunPassesOnFunction`/`AutoOpt` 全部支持，内部把机器传给
`LLVMRunPasses` 的 `tm` 参数（不再固定空句柄）。已关闭的机器会 panic
`ErrUseAfterFree`；传 `nil` 等价于不传。`pass` 包现在依赖 `llvm/target`
（依赖方向更新为 `llvm → ir → target → {jit, pass}` 的直连形式）。`Option`
的底层类型改成了包内私有配置闭包，但外部模块本来就无法 import
`internal/binding`、写不出自定义 Option 字面量，因此对调用方无破坏性影响。

## 8. `target.TargetMachine` 与应用模块的 DataLayout 绑定时序容易被搞错

`TargetMachine.ApplyTo(module)` 必须**在生成 IR 之前**调用，否则 Builder 生成的
load/store 对齐信息来自默认 DataLayout（i64 对齐 4），而不是目标布局（i64 对齐
8）；而 `EmitToFile` 只在结束时代码生成时才使用模块 DL。附带的坑：
`llvm.NewDataLayout("")` 查询出来的 i64 ABI 对齐是 4，与宿主不一致，想拿宿主的
布局只能通过 `target.TargetMachine.DataLayout()`。建议在 package doc 或
`NewModule` 的文档里显著提示先 `ApplyTo` 再建 IR；或提供
`target.TargetMachine.EnsureApplied(module)` 之类更直白的 API。

**解决方法（已实现，未加 `EnsureApplied`）**：

* 文档：`target` 包文档新增 "Data layout ordering" 一节，`ApplyTo`、`ir.NewModule`
  文档都显著写明"必须在生成 IR 之前调用"；`llvm.NewDataLayout("")` 文档写明
  空串得到 LLVM 默认布局（i64 对齐 4）而非宿主布局，要宿主布局用
  `TargetMachine.DataLayout()`。
* 运行时兜底（仅调试构建）：`EmitToFile`/`Emit` 会校验模块数据布局与目标机器
  一致，不符则 panic `ErrInvalidArg` 并提示先 `ApplyTo`——"忘记 ApplyTo"从
  静默生成错对齐代码变成当场报错。
* 不加 `EnsureApplied`：`ApplyTo` 本身就是幂等的纯 setter（重复调用安全），
  别名只会增加 API 面积，文档才是这个坑的正解。

## 9. `Context.Bool()` 等类型/常量返回的具体角色不一致（轻微）

`Context.Bool()` 返回 `IntType`（i1），`Context.ConstBool()` 返回 `Value[IntT]`，
两者配合没问题，但对动态代码来说 `bool` 与 `i1` 的概念容易混。属于命名/直觉
问题。

**解决方法（已实现，仅文档）**：`Context.Bool()` 文档写明"LLVM 没有独立 bool
类型，bool 即 i1 整数类型"，`ConstBool` 文档写明返回 `Value[IntT]`。不改名
（改名是破坏性变更，且 i1 语义本身没错）。

## 10. 其它可用性观察

* `ir.Uses`/`OpOf` 对"用户不是指令"的畸形 IR 无法给出比 LLVM 更友好的诊断
  （见问题 1）。
* `Module.String()` 不 Verify，非法 IR 可能打印出"看起来合法"的文本（问题 1
  的误导来源之一），文档已有说明，但对编译器调试不够友好；可考虑提供
  `StringChecked()`。

**解决方法（已实现）**：问题 1 修复后，"常量里包含指令"这类畸形值在构造时即被
拦截，`Uses`/`OpOf` 不再会遇到该来源的畸形输入，无需额外诊断层。`Module.String()`
文档补充了推荐模式（要合法文本先 `m.Verify()` 再 `m.String()`）；不新增
`StringChecked()`——它只是这两行的组合，不值得增加 API 面积。
