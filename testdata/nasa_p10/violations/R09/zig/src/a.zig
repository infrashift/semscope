const std = @import("std");
const assert = std.debug.assert;

pub const Handler = *const fn (u32) void;

pub const Ops = struct {
    run: *const fn () void,
};

pub fn apply(comptime f: fn (u32) u32, n: u32) u32 {
    assert(n < 100);
    assert(n >= 0);
    const p: **u8 = undefined;
    _ = p;
    return f(n);
}
