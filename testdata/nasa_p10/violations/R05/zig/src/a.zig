const std = @import("std");
const assert = std.debug.assert;

pub fn one(n: u32) u32 {
    assert(n < 100);
    return n;
}

pub fn two(n: u32) u32 {
    assert(n < 100);
    assert(n >= 0);
    return n;
}

pub fn none(n: u32) u32 {
    return n;
}
