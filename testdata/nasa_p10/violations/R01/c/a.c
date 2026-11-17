#include <setjmp.h>
static jmp_buf env;
int fact(int n) {
    return n <= 1 ? 1 : n * fact(n - 1);
}
int jump(int n) {
start:
    if (n > 0) { n--; goto start; }
    if (setjmp(env)) { return 0; }
    return n;
}
