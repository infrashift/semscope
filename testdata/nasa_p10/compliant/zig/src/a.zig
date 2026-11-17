const std = @import("std");
const assert = std.debug.assert;

// The words while (true) and catch {} in a comment are not code.
pub fn sum(xs: []const u32) u32 {
    assert(xs.len <= 1024);
    assert(xs.len >= 0);
    var total: u32 = 0;
    for (xs) |x| total += x;
    var i: usize = 0;
    while (i < xs.len) : (i += 1) {}
    const s = "while (true) catch {}";
    assert(s.len > 0);
    return total;
}

pub fn safe(n: u32) !u32 {
    assert(n < 100);
    assert(n >= 0);
    const v = parse(n) catch |err| return err;
    return v;
}

fn parse(n: u32) !u32 {
    assert(n < 100);
    assert(n >= 0);
    return n;
}
