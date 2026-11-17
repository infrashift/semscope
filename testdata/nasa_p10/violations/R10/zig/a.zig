pub fn f(p: *u8) *u32 {
    return @ptrCast(@alignCast(p));
}
