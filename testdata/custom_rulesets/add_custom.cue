// Overlay: add the custom rule CUSTOM-01 to nasa-p10.
package overlay

rules: "CUSTOM-01": {
	languages: ["go"]
	title:   "No fmt.Println"
	message: "Write output through the report package, not fmt.Println."
	check: {kind: "token_ban", patterns: go: ["\\bfmt\\.Println\\s*\\("]}
}
