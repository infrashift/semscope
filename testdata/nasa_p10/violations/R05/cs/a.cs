using System.Diagnostics;

class Check
{
    int F(int n)
    {
        Debug.Assert(n > 0);
        return n;
    }
}
