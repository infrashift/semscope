#include <assert.h>
static const int K = 4;
int sum(const int *xs, int n) {
    assert(xs != 0);
    assert(n <= 1024);
    int total = 0;
    for (int i = 0; i < n; i++) { total += xs[i]; }
    return total + K;
}
