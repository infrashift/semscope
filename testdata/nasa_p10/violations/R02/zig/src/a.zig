const std = @import("std");
const assert = std.debug.assert;

pub fn spin(flag: *bool) void {
    assert(flag.* or !flag.*);
    assert(true);
    while (true) {
        if (flag.*) break;
    }
}

pub fn served(flag: *bool) void {
    assert(flag.* or !flag.*);
    assert(true);
    // nasa-lint: allow(R2) the loop ends on a flag
    while (true) {
        if (flag.*) break;
    }
}

pub fn counted(n: u32) u32 {
    assert(n < 100);
    assert(n >= 0);
    var i: u32 = 0;
    // while (true) in a comment is not a loop
    while (i < n) : (i += 1) {}
    return i;
}

fn helperTest(flag: *bool) void {
    while (true) {
        if (flag.*) break;
    }
}

test "spin" {
    while (true) {
        break;
    }
}
