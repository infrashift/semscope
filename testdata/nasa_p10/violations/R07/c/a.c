#include <stdio.h>
void f(void) {
    fopen("a", "r");
    (void)fclose(stdin);
    FILE *g = fopen("b", "r");
}
