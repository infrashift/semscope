/// Doc line one.
/// Doc line two.
pub inline fn f() void {}
export fn g() void {}
const S = struct {
    fn m(self: S) void {
        self.m();
        Self.m();
        _ = try h();
        x = y;
    }
};
const names = blk: {
    var local: [2]u8 = undefined;
    break :blk local;
};
const Counter = struct {
    var hits: u32 = 0;
};
