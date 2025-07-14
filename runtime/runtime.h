#include <stdint.h>
#include <setjmp.h>
#include <stdio.h>

#ifdef _WIN64
#include <string.h>
#include <windows.h>
#endif

void sim_runtime_gc_init(void* stack_pos);
void* sim_runtime_gc_alloc(size_t size);
void sim_runtime_gc_gc();
