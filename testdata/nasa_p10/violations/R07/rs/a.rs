fn f() {
    let _ = g();
    let v = h().unwrap();
    let w = h().expect("h");
}
