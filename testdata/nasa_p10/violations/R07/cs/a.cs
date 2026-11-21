using System;

class Calls
{
    int G() { return 1; }

    void F()
    {
        _ = G();
        try { G(); } catch (Exception) { }
        try { G(); } catch (ArgumentException) { }
    }
}
