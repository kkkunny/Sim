# Sim

Sim是一门简洁的、强类型的编译型语言

## Features：

+ 语法简单，关键字尽可能的少，向C与Go语言看齐

+ 面向对象，Go+Rust

+ 自动内存管理

## TODO List

+ [x] 基础语法（基础运算 / 流程控制 / 函数 / 变量）

+ [x] 基本类型（int / uint / float / bool / pointer with null / pointer not with null / function / array / tuple / struct / union）

+ [x] 函数/变量导出 && 函数/变量链接

+ [x] 类型定义 && 类型别名

+ [x] 方法定义与调用

+ [x] 泛型（泛型函数 / 泛型结构体 / 泛型方法）

+ [ ] trait

+ [ ] 泛型约束

+ [ ] defer

+ [x] 异常处理

+ [ ] 垃圾回收

+ [ ] 闭包

## Dependences

+ linux / windows

+ llvm(version==17 or 18)

+ golang

+ c lib

## Hello World

compiler/examples/hello_world.sim

```go
import std::io

func main(){
    io::println("hello world")
}
```

```shell
> make run TEST_FILE=$PWD/compiler/examples/hello_world.sim
hello world
```