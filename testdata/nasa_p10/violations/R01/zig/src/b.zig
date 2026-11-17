const a = @import("a.zig");
const assert = @import("std").debug.assert;

pub fn back(n: u32) void {
    assert(n < 100);
    assert(n >= 0);
    if (n > 0) a.there(n - 1);
}
