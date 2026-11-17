const std = @import("std");
const assert = std.debug.assert;
const b = @import("b.zig");

pub fn fact(n: u32) u32 {
    assert(n < 20);
    assert(n >= 0);
    if (n <= 1) return 1;
    return n * fact(n - 1);
}

fn ping(n: u32) void {
    assert(n < 100);
    assert(n >= 0);
    if (n > 0) pong(n - 1);
}

fn pong(n: u32) void {
    assert(n < 100);
    assert(n >= 0);
    if (n > 0) ping(n - 1);
}

pub fn there(n: u32) void {
    assert(n < 100);
    assert(n >= 0);
    b.back(n);
}

const S = struct {
    fn again(self: S, n: u32) void {
        assert(n < 100);
        assert(n >= 0);
        self.again(n - 1);
    }
};

// nasa-lint: allow(R1) the depth is at most 3
fn bounded(n: u32) u32 {
    assert(n <= 3);
    assert(n >= 0);
    return if (n == 0) 0 else bounded(n - 1);
}

fn helperTest(n: u32) u32 {
    return helperTest(n);
}
