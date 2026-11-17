void spin(volatile int *flag) {
    while (1) { if (*flag) break; }
    for (;;) { if (*flag) break; }
    do { if (*flag) break; } while (true);
    for (int i = 0; i < 3; i++) {}
}
