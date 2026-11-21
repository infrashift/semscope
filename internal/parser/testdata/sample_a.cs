using System.Diagnostics;

class Walker
{
    static int count;
    static readonly int Limit = 3;

    int Walk(int n, delegate*<int, int> f)
    {
        Debug.Assert(n > 0);
        while (true) { }
        for (;;) { }
        _ = Step(n);
        int** p = null;
        unsafe { Step(n); }
        string s = @"// Walk(n)";
        string r = """goto x""";
        try { Step(n); } catch { }
        this.Walk(n - 1, f);
        return Walker.Step(n);
    }

    static int Step(int n) => n;

    public int Count { get { return count; } }
}

[TestClass]
class WalkerTests
{
    [TestMethod]
    void T() { while (true) { } }
}
#define X
