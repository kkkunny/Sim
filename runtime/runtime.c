#include "runtime.h"

extern void sim_runtime_gc__init(void* stack_begin, void* data_begin, void* data_end);
extern void* sim_runtime_gc__alloc(size_t size, size_t stack_pos);
extern void sim_runtime_gc__gc(size_t stack_pos);

#ifdef _WIN64
typedef struct{
    void* start;
    void* end;
}ptrs;

static ptrs get_section_range(const char* section_name) {
    HMODULE hModule = GetModuleHandle(NULL);
    if (!hModule) {
        return (ptrs){NULL, NULL};
    }

    PIMAGE_DOS_HEADER dos = (PIMAGE_DOS_HEADER)hModule;
    PIMAGE_NT_HEADERS nt = (PIMAGE_NT_HEADERS)((char*)hModule + dos->e_lfanew);
    PIMAGE_SECTION_HEADER section = IMAGE_FIRST_SECTION(nt);

    for (int i = 0; i < nt->FileHeader.NumberOfSections; i++, section++) {
        if (strncmp((const char*)section->Name, section_name, 8) == 0) {
            char* start = (char*)hModule + section->VirtualAddress;
            char* end = start + section->Misc.VirtualSize;
            return (ptrs){start, end};
        }
    }
    return (ptrs){NULL, NULL};
}
#else
extern char __data_start, _edata;
#endif

void sim_runtime_gc_init(void* stack_begin){
#ifdef _WIN64
    ptrs data_ptrs = get_section_range(".data");
    void* data_begin = data_ptrs.start;
    void* data_end = data_ptrs.end;
#else
    void* data_begin = &__data_start;
    void* data_end = &_edata;
#endif
	return sim_runtime_gc__init(stack_begin, data_begin, data_end);
}

void* sim_runtime_gc_alloc(size_t size){
    jmp_buf env;
    setjmp(env);
	size_t temp;
	return sim_runtime_gc__alloc(size, (size_t)(&temp));
}

void sim_runtime_gc_gc(){
    jmp_buf env;
    setjmp(env);
	size_t temp;
	sim_runtime_gc__gc((size_t)(&temp));
}
