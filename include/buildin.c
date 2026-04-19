#ifndef _SIM_BUILDIN
#define _SIM_BUILDIN	1

#define NULL ((void *)0)

// 基础类型
#define i8 signed char
#define i16 signed short
#define i32 signed int
#define i64 signed long long

#define u8 unsigned char
#define u16 unsigned short
#define u32 unsigned int
#define u64 unsigned long long

#define f32 float
#define f64 double

#define ADD(x, y) _Generic((x), \
    i8: x + y, \
    i16: x + y, \
    i32: x + y, \
    i64: x + y, \
    u8: x + y, \
    u16: x + y, \
    u32: x + y, \
    u64: x + y, \
    f32: x + y, \
    f64: x + y)

#define SUB(x, y) _Generic((x), \
    i8: x - y, \
    i16: x - y, \
    i32: x - y, \
    i64: x - y, \
    u8: x - y, \
    u16: x - y, \
    u32: x - y, \
    u64: x - y, \
    f32: x - y, \
    f64: x - y)

#define MUL(x, y) _Generic((x), \
    i8: x * y, \
    i16: x * y, \
    i32: x * y, \
    i64: x * y, \
    u8: x * y, \
    u16: x * y, \
    u32: x * y, \
    u64: x * y, \
    f32: x * y, \
    f64: x * y)

#define QUO(x, y) _Generic((x), \
    i8: x / y, \
    i16: x / y, \
    i32: x / y, \
    i64: x / y, \
    u8: x / y, \
    u16: x / y, \
    u32: x / y, \
    u64: x / y, \
    f32: x / y, \
    f64: x / y)

#define REM(x, y) _Generic((x), \
    i8: x % y, \
    i16: x % y, \
    i32: x % y, \
    i64: x % y, \
    u8: x % y, \
    u16: x % y, \
    u32: x % y, \
    u64: x % y, \
    f32: x % y, \
    f64: x % y)

#define AND(x, y) _Generic((x), \
    i8: x & y, \
    i16: x & y, \
    i32: x & y, \
    i64: x & y, \
    u8: x & y, \
    u16: x & y, \
    u32: x & y, \
    u64: x & y, \
    f32: x & y, \
    f64: x & y)

#define OR(x, y) _Generic((x), \
    i8: x | y, \
    i16: x | y, \
    i32: x | y, \
    i64: x | y, \
    u8: x | y, \
    u16: x | y, \
    u32: x | y, \
    u64: x | y, \
    f32: x | y, \
    f64: x | y)

#define XOR(x, y) _Generic((x), \
    i8: x ^ y, \
    i16: x ^ y, \
    i32: x ^ y, \
    i64: x ^ y, \
    u8: x ^ y, \
    u16: x ^ y, \
    u32: x ^ y, \
    u64: x ^ y, \
    f32: x ^ y, \
    f64: x ^ y)

#define NOT(x) _Generic((x), \
    i8: ~x, \
    i16: ~x, \
    i32: ~x, \
    i64: ~x, \
    u8: ~x, \
    u16: ~x, \
    u32: ~x, \
    u64: ~x, \
    f32: ~x, \
    f64: ~x)

// 函数胖指针与闭包
#define FUNCTYPE(ft, ct) struct{union{ft f; ct c;} func; void* env;}
#define FUNCEXPR_F(v) {.func.f=v, .env=NULL}
#define FUNCEXPR_C(e, v) {.func.f=v, .env=e}
#define FUNCCALL(v, ...) v.env==NULL?v.func.f(__VA_ARGS__):v.func.c(v.env, __VA_ARGS__)

#endif
