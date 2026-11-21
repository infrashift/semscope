class Spin
{
    void Run(bool flag)
    {
        while (true) { if (flag) { break; } }
        for (;;) { if (flag) { break; } }
        do { if (flag) { break; } } while (true);
        for (int i = 0; i < 3; i++) { }
    }
}
