const names = blk: {
    var local: [2]u8 = undefined;
    break :blk local;
};
var counter: u32 = 0;
const S = struct {
    var hits: u32 = 0;
};
