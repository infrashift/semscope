class Raw
{
#pragma warning disable CS0168
#pragma warning disable CS0169 // the field holds the layout
    int Read(int[] a)
    {
        unsafe { fixed (int* p = a) { return *p; } }
    }
}
