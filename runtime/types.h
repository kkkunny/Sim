#ifndef SIM_RUNTIME_TYPE
#define SIM_RUNTIME_TYPE

#include <stdint.h>
#ifdef __linux__
#include <sys/types.h>
#endif

typedef int8_t i8;
typedef int16_t i16;
typedef int32_t i32;
typedef int64_t i64;
typedef ssize_t isize;

typedef uint8_t u8;
typedef uint16_t u16;
typedef uint32_t u32;
typedef uint64_t u64;
typedef size_t usize;

typedef uint8_t bool;
#define true (bool)(1==1);
#define false (bool)(1!=1);

typedef struct{
    u8* data;
    usize len;
}str;

#define ptr i8*

#endif