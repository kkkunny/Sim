#ifndef _SIM_BUILDIN
#define _SIM_BUILDIN	1

#include <math.h>

// 工具
#ifndef NULL
#define NULL ((void *)0)
#endif
typedef struct{} ZERO_TYPE;
ZERO_TYPE ZERO_TYPE_VALUE = {};

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

#define bool _Bool
const bool true = 1==1;
const bool false = 1!=1;

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
    f32: fmodf(x, y), \
    f64: fmod(x, y)

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
    f64: ~x, \
    _Bool: !x)

#define SELFADD(x) _Generic((x), \
    i8: x++, \
    i16: x++, \
    i32: x++, \
    i64: x++, \
    u8: x++, \
    u16: x++, \
    u32: x++, \
    u64: x++, \
    f32: x++, \
    f64: x++)

#define EQ(x, y) _Generic((x), \
    i8: x == y, \
    i16: x == y, \
    i32: x == y, \
    i64: x == y, \
    u8: x == y, \
    u16: x == y, \
    u32: x == y, \
    u64: x == y, \
    f32: x == y, \
    f64: x == y, \
    bool: x == y, \
    default: x == y)

#define NEQ(x, y) _Generic((x), \
    i8: x != y, \
    i16: x != y, \
    i32: x != y, \
    i64: x != y, \
    u8: x != y, \
    u16: x != y, \
    u32: x != y, \
    u64: x != y, \
    f32: x != y, \
    f64: x != y, \
    bool: x != y, \
    default: x != y)

#define LT(x, y) _Generic((x), \
    i8: x < y, \
    i16: x < y, \
    i32: x < y, \
    i64: x < y, \
    u8: x < y, \
    u16: x < y, \
    u32: x < y, \
    u64: x < y, \
    f32: x < y, \
    f64: x < y)

#define LTE(x, y) _Generic((x), \
    i8: x <= y, \
    i16: x <= y, \
    i32: x <= y, \
    i64: x <= y, \
    u8: x <= y, \
    u16: x <= y, \
    u32: x <= y, \
    u64: x <= y, \
    f32: x <= y, \
    f64: x <= y)

#define GT(x, y) _Generic((x), \
    i8: x > y, \
    i16: x > y, \
    i32: x > y, \
    i64: x > y, \
    u8: x > y, \
    u16: x > y, \
    u32: x > y, \
    u64: x > y, \
    f32: x > y, \
    f64: x > y)

#define GTE(x, y) _Generic((x), \
    i8: x >= y, \
    i16: x >= y, \
    i32: x >= y, \
    i64: x >= y, \
    u8: x >= y, \
    u16: x >= y, \
    u32: x >= y, \
    u64: x >= y, \
    f32: x >= y, \
    f64: x >= y)

// 函数胖指针与闭包
#define FUNCTYPE(ft, ct) struct{union{ft f; ct c;} func; void* ctx;}
#define FUNCEXPR_F(expr) {.func.f=expr, .ctx=NULL}
#define FUNCEXPR_C(expr, ctxv) {.func.c=expr, .ctx=ctxv}
#define FUNCCALL(expr, ...) expr.ctx==NULL?expr.func.f(__VA_ARGS__):expr.func.c(expr.ctx, __VA_ARGS__)

// 主函数
static void sim_main();
int main(){
    sim_main();
    return 0;
}

#endif
