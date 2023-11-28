#ifndef SIM_RUNTIME
#define SIM_RUNTIME

#include "types.h"

bool sim_runtime_str_eq_str(str l, str r);
void sim_runtime_debug(str s);
ptr sim_runtime_check_null(ptr s);

#endif