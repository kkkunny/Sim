# Sim

Sim是一门简洁的、强类型的编译型语言

## Features：

+ 语法简单，关键字尽可能的少，向C与Go语言看齐

+ 面向对象，Go+Rust

+ 自动内存管理

## TODO List

+ [x] 基础语法（基础运算 / 流程控制 / 函数 / 变量）

+ [x] 基本类型（int / uint / float / bool / reference / function / array / tuple / struct / union）

+ [x] 函数/变量导出 && 函数/变量链接

+ [x] 类型定义 && 类型别名

+ [x] 方法定义与调用

+ [x] 泛型（泛型函数 / 泛型结构体 / 泛型方法）

+ [x] trait

+ [x] 运算符重载

+ [x] 泛型约束

+ [x] 异常处理

+ [x] 垃圾回收

+ [x] 闭包

## Dependences

+ linux

+ llvm(version==23，go-llvm 绑定)

+ clang（用作链接驱动，gcc 作为回退）

+ golang(>=1.27)

## Hello World

`examples/main.sim`

```sim
import std::c

type S struct {
	name: str
}

let getname | S = (self: &Self) -> str {
	return self.name
}

let main = () {
	let s = S{name: "123"}
	let ss = &s
	c::puts((*ss).getname())
}
```

```shell
> go run -tags compile . examples/main.sim
> ./main.out
123
```

## Debug stages

```shell
go run -tags lex . <file.sim>       # 打印token
go run -tags parse . <file.sim>     # 打印AST
go run -tags analyze . <file.sim>   # 打印HIR
go run -tags codegen . <file.sim>   # 打印LLVM IR
go run -tags compile . <file.sim>   # 编译生成main.out
go run -tags debug .                # 编译并运行examples/main.sim
```