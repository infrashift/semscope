def f():
    try:
        g()
    except:
        pass
    try:
        g()
    except ValueError:
        pass
