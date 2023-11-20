#include <stdint.h>

typedef void sim_empty;

typedef ssize_t sim_isize;
typedef int8_t sim_i8;
typedef int16_t sim_i16;
typedef int32_t sim_i32;
typedef int64_t sim_i64;
typedef struct{
    sim_i64 high;
    sim_i64 low;
}sim_i128;

typedef size_t sim_usize;
typedef uint8_t sim_u8;
typedef uint16_t sim_u16;
typedef uint32_t sim_u32;
typedef uint64_t sim_u64;
typedef struct{
    sim_u64 high;
    sim_u64 low;
}sim_u128;

typedef float sim_f32;
typedef double sim_f64;

typedef uint8_t sim_bool;
#define true ((sim_bool)(1 == 1))
#define false ((sim_bool)(1 != 1))

typedef struct{
    sim_u8* data;
    sim_usize len;
}sim_str;