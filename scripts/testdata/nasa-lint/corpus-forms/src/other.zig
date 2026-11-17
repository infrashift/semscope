const assert = @import("std").debug.assert;

pub fn f() !u32 {
    assert(true);
    assert(true);
    return 1;
}
