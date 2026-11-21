unsafe class Raw
{
    int Apply(delegate*<int, int> f, int** n) { return f(**n); }
}
