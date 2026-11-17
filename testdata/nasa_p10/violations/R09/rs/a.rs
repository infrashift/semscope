fn apply(f: fn(u32) -> u32, n: &&u32) -> u32 { f(**n) }
