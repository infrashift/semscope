const std = @import("std");
pub fn init(a: std.mem.Allocator) ![]u8 {
    return a.alloc(u8, 64);
}
pub fn work(a: std.mem.Allocator) ![]u8 {
    return a.alloc(u8, 8);
}
