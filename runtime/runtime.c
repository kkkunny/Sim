#include "runtime.h"

extern void* sim_runtime_gc__alloc(size_t size, size_t stack_pos);
extern void sim_runtime_gc__gc(size_t stack_pos);

void* sim_runtime_gc_alloc(size_t size){
	char temp;
	return sim_runtime_gc__alloc(size, (size_t)(&temp));
}

void sim_runtime_gc_gc(){
	char temp;
	sim_runtime_gc__gc((size_t)(&temp));
}
