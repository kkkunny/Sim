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

+ linux（当前仅在 linux/x86-64 验证）

+ llvm(version==22) 开发头文件（`github.com/kkkunny/go-llvm` 经 cgo 链接 libLLVM）

+ clang（链接驱动）

+ golang(>=1.27)

## Hello World

```sim
import std::c

let main = () {
	c::puts("Hello World")
}
```

```shell
> go run -tags compile . hello_world.sim
> ./main.out
Hello World
```