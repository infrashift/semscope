#include <stdlib.h>
static char *buf;
void init_buffers(void) { buf = malloc(64); }
void work(void) { char *p = malloc(8); free(p); }
