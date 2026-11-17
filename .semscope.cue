// The overlay of the semscope self-scan (zmake selfscan).
// ADR-0003 records each change to nasa-p10 in this file.
package overlay

// R3 deviation: Go, cobra, CUE and Tree-sitter allocate memory at run time.
// semscope sizes its pools in engine.New and caps the rest with hard limits.
rules: "NASA-P10-03": enabled: false
