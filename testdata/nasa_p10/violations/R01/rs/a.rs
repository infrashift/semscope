struct Walker;
impl Walker {
    fn walk(n: u32) -> u32 {
        if n > 0 { Self::walk(n - 1) } else { 0 }
    }
}
