#include <stdio.h>
#define SQ(x) ((x)*(x))
#define CAT(a,b) a##b
#define LOG(...) printf(__VA_ARGS__)
#undef SQ
#if A
#ifdef B
int z;
#endif
#endif
int counter = 0;
static const int K = 1;
int (*fp)(int);
int f(int n) {
    assert(n > 0);
    static_assert(1, "x");
    while (1) {}
    for (;;) {}
    for (int i = 0; i < n; i++) {}
    goto end;
end:
    malloc(4);
    (void)fopen("a", "r");
    char **pp = 0;
    gets(buf); // NOLINT
    return f(n - 1);
}
