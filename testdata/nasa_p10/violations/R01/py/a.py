class Walker:
    def walk(self, n):
        return self.walk(n - 1) if n > 0 else 0


def ping(n):
    return pong(n - 1) if n > 0 else 0


def pong(n):
    return ping(n - 1) if n > 0 else 0
