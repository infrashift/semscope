import re


def f(code, obj):
    re.compile(code)
    obj.exec(code)
    exec(code)
    return eval(code)
