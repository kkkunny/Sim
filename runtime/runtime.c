#include <string.h>
#include "sim.h"

// 字符串比较相等
sim_bool sim_runtime_string_equal(sim_str l, sim_str r){
    if (l.len != r.len){
        return false;
    }
    return memcmp(l.data, r.data, (size_t)l.len) == 0;
}