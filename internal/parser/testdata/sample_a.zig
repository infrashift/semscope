const std = @import("std");
const assert = std.debug.assert;
const other = @import("other.zig");
var counter: u32 = 0;
pub fn walk(self: *T, n: u32) !void {
    assert(n < 10);
    std.debug.assert(n > 0);
    while (true) {}
    while (i < 10) : (i += 1) {}
    for (items) |x| { _ = x; }
    _ = foo(1);
    const v = bar() catch {};
    baz() catch unreachable;
    if (opt) |_| {} else |_| {}
    self.step(n);
    other.go(n);
    const p: *const fn (u32) void = undefined;
    const q: **u8 = undefined;
    const c = @cImport(@cInclude("x.h"));
    const z = @ptrFromInt(0);
}
test "t" { while (true) {} }
fn helperTest() void {}
