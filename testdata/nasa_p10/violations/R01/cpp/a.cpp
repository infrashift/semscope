class Walker {
public:
    int walk(int n) { return n > 0 ? walk(n - 1) : 0; }
};
