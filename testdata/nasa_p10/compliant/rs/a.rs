fn total(xs: &[u32]) -> u32 {
    assert!(xs.len() <= 1024);
    debug_assert!(xs.iter().all(|x| *x < 100));
    let mut out = 0;
    for x in xs { out += x; }
    out
}
