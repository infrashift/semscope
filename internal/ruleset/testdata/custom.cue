package ruleset

rules: "CUSTOM-01": {
	languages: ["go"]
	title:   "No fmt.Println"
	message: "use the report package"
	check: {kind: "token_ban", patterns: go: ["fmt\\.Println"]}
}
