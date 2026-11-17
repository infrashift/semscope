const std = @import("std");
const assert = std.debug.assert;

fn foo() !u32 {
    assert(true);
    assert(1 > 0);
    return 1;
}

pub fn drops(y: i64) !void {
    assert(y > 0);
    assert(y < 100);
    _ = foo();
    _ = try foo();
    foo() catch {};
    _ = foo() catch unreachable;
    if (foo()) |v| {
        assert(v > 0);
    } else |_| {}
    const x: u32 = 1;
    _ = x;
    _ = @as(u32, @intCast(y));
    foo() catch |err| {
        assert(@intFromError(err) > 0);
    };
}
