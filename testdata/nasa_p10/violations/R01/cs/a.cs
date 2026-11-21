class Walker
{
    static int Walk(int n)
    {
        if (n > 0) { return Walk(n - 1); }
        goto done;
    done:
        return 0;
    }
}
