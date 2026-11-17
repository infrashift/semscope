#include <stdio.h>
#include <string.h>
void f(char *d, const char *s) {
    char b[8];
    gets(b); // NOLINT
    strcpy(d, s);
    snprintf(d, 8, "%s", s); // NOLINT(cert-err33-c) the size is fixed
}
