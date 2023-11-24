#include "runtime.h"
#include <string.h>

bool sim_runtime_str_eq_str(str s1, str s2){
    if (s1.len != s2.len){
        return false;
    }
    return memcmp(s1.data, s2.data, s1.len) == 0;
}
