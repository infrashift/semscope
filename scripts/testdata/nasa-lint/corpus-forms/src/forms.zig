//! The forms of the Zig standard library that ISSUE-007 lists. Each form
//! has a comment with the rule and the expected result.
const std = @import("std");
const assert = std.debug.assert;
const other = @import("other.zig");

// R1: two containers have a toFloat. Const.toFloat recurses; Mutable.toFloat
// calls a method of an expression and gives no edge.
const Mutable = struct {
    x: u32,
    fn toConst(self: Mutable) Const {
        assert(self.x < 10);
        assert(self.x >= 0);
        return .{ .x = self.x };
    }
    fn toFloat(self: Mutable) f64 {
        assert(self.x < 10);
        assert(self.x >= 0);
        return self.toConst().toFloat();
    }
};

const Const = struct {
    x: u32,
    fn toFloat(self: Const) f64 {
        assert(self.x < 10);
        assert(self.x >= 0);
        if (self.x == 0) return self.toFloat();
        return @floatFromInt(self.x);
    }
};

// R1: a decl literal and a builtin of the same name are not calls of init.
fn init(n: u32) Mutable {
    assert(n < 10);
    assert(n >= 0);
    const m: Mutable = .init(n);
    _ = @sin(1.0);
    return m;
}

// R1 and R5: the struct that List returns. append and grow recurse. The
// assertions of append count for append only, not for List.
fn List(comptime T: type) type {
    return struct {
        fn append(self: *@This(), x: T) void {
            assert(x != 0);
            assert(x < 100);
            self.grow();
        }
        fn grow(self: *@This()) void {
            self.append(1);
        }
    };
}

// R1: a prefix operator binds to the callee: ?Span(T) and !valid(T) recurse.
fn Span(comptime T: type) type {
    if (T == u8) return u8;
    return ?Span(u8);
}

inline fn valid(comptime T: type) bool {
    if (T == u8) return true;
    return !valid(u8);
}

// R4: a struct parameter ends the first line with `{`. The function has
// 62 lines, and the body starts after `}) void {`.
fn setStruct(n: u32, args: struct {
    a: u32,
    b: u32,
}) void {
    assert(n < 10);
    assert(args.a < 10);
    var sum: u32 = 0;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    sum += 1;
    assert(sum > 0);
}

// R4 and R5: noinline, a quoted name and a one-line function are functions.
noinline fn current() u32 {
    assert(true);
    assert(true);
    return 1;
}

fn @"resume"() void {
    assert(true);
    assert(true);
}

inline fn leave() void {}

// R7: the forms on two lines, with a comment in the block, the destructure
// forms and a call through @This().
fn swallow(r: anyerror!u32) void {
    assert(true);
    assert(true);
    _ = other.f() catch
        unreachable;
    other.g() catch {
        // The error does not matter here.
    };
    if (r) |v| other.h(v) else |_| {
        // Ignore the error.
    }
    const a, _ = other.pair();
    _, const b = other.pair();
    _ = try @This().get();
    const c, _ = other.pair().field;
    _ = a + b + c;
}

fn get() !u32 {
    assert(true);
    assert(true);
    return 1;
}
