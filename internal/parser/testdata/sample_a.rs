static mut COUNT: u32 = 0;
macro_rules! sq { ($x:expr) => { $x * $x }; }
fn walk(n: u32, f: fn(u32) -> u32) -> u32 {
    assert!(n > 0);
    debug_assert_eq!(n, 1);
    loop {}
    while true {}
    for i in 0..n {}
    let _ = foo();
    let v = bar().unwrap();
    let b = Box::new(5);
    let p: &&u32 = &&n;
    let r: *const *const u8 = std::ptr::null();
    unsafe { x(); }
    self.step(n);
    Self::go(n);
    walk(n - 1, f)
}
#[cfg(test)]
mod tests { #[test] fn t() { loop {} } }
