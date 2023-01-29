# Sim

Sim是一门简洁的、强类型的编译型语言

## Snake

```shell
> sim run examples/snake.sim
```

<div><img src=./examples/snake.gif width=40%/></div>

## Features：

+ 语法简单

  + c语言风格的字符串，没有内置string类型
  
  + 简单（残缺）的面向对象，类似于go

+ 无运行时开销（依赖c语言运行时）
  
+ 手动内存管理（malloc / free）

## TODO List

+ [x] 基础语法（基础运算 / 流程控制 / 函数 / 全局变量）

+ [x] 基本类型（int / uint / float / bool / pointer / function / array / tuple / struct）

+ [x] C标准库

+ [x] 类型定义

+ [x] 方法定义与调用

+ [x] 接口

+ [ ] defer

+ [ ] 异常处理

+ [ ] 垃圾回收

+ [ ] 泛型

## Dependences

+ linux

+ llvm14

+ golang(version>=1.18)

+ clang / gcc

## Install

```shell
> git clone https://github.com/kkkunny/Sim.git
> cd Sim
> go mod download
> make build
```

### Docker

```shell
> make docker
> docker run -it --name sim sim
```

## Hello World

examples/hello_world.sim

```go
import std.io
import std.container.string

@extern(main)
func main()u8{
    io::println(string::new("Hello World"))
    return 0
}
```

```shell
> sim run examples/hello_world.sim
Hello World
```