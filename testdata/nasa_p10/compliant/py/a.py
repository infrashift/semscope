def total(xs):
    assert len(xs) <= 1024
    assert all(x >= 0 for x in xs)
    out = 0
    for x in xs:
        out += x
    return out
