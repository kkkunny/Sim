#ifndef _SIM_BUILDIN
#define _SIM_BUILDIN	1

#include <math.h>

// 工具
#ifndef NULL
#define NULL ((void *)0)
#endif
typedef struct{} ZERO_TYPE;
const ZERO_TYPE ZERO_TYPE_VALUE = {};

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

#define SHL(x, y) _Generic((x), \
    i8: x << y, \
    i16: x << y, \
    i32: x << y, \
    i64: x << y, \
    u8: x << y, \
    u16: x << y, \
    u32: x << y, \
    u64: x << y)

#define SHR(x, y) _Generic((x), \
    i8: x >> y, \
    i16: x >> y, \
    i32: x >> y, \
    i64: x >> y, \
    u8: x >> y, \
    u16: x >> y, \
    u32: x >> y, \
    u64: x >> y)

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

#define LOGIC_AND(x, y) x && y

#define LOGIC_OR(x, y) x || y

// 函数胖指针与闭包
#define FUNC_TYPE(ret, ...) struct { \
    union { \
        ret (*f)(__VA_ARGS__);\
        ret (*c)(void*, ##__VA_ARGS__);\
    } func;\
    void* ctx;\
}
#define FUNC_EXPR_F(expr) {.func.f=expr, .ctx=NULL}
#define FUNC_EXPR_C(expr, ctxv) {.func.c=expr, .ctx=ctxv}
#define FUNC_CALL(expr, ...) expr.ctx==NULL?expr.func.f(__VA_ARGS__):expr.func.c(expr.ctx, __VA_ARGS__)
#define FUNC_EQ(x, y) x.ctx == y.ctx && (x.ctx == NULL ? x.func.f == y.func.f : x.func.c == y.func.c)
#define FUNC_NEQ(x, y) x.ctx != y.ctx || (x.ctx == NULL ? x.func.f != y.func.f : x.func.c != y.func.c)

// 数组
#define ARRAY_TYPE(elem, size) struct{elem array[size];}
#define ARRAY_INDEX(a, i) a.array[i]

// 元组
#define TUPLE_INDEX(t, i) t.e##i

// 联合
#define UNION_TYPE_INDEX(v) v.t
#define UNION_VALUE_INDEX(v, i) v.v.t##i

// 主函数
static void sim_main();
int main(){
    sim_main();
    return 0;
}

#endif
