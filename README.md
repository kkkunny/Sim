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

+ [ ] 泛型（泛型函数 / 泛型结构体 / 泛型方法）

+ [x] trait

+ [x] 运算符重载

+ [ ] 泛型约束

+ [ ] defer

+ [ ] 异常处理

+ [ ] 垃圾回收

+ [x] 闭包

## Dependences

+ linux / windows

+ llvm(version==18)

+ golang

+ c lib

## Hello World

compiler/examples/hello_world.sim

```go
func main(){
    debug("Hello World")
}
```

```shell
> make run TEST_FILE=$PWD/compiler/examples/hello_world.sim
Hello World
```