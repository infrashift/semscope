package ruleset

name: "test-base"
rules: {
	"TEST-01": {
		languages: ["go", "zig"]
		title:   "Short functions"
		message: "the function has too many lines"
		params: max_lines: *60 | int
		check: kind: "max_function_lines"
	}
	"TEST-02": {
		languages: ["go"]
		title:   "Assertion density"
		message: "too few assertions"
		params: {
			min_assertions: *2 | int
			mode:           *"average" | "per_function"
			strict:         *false | bool
			allow:          *["init"] | [...string]
		}
		check: {kind: "min_assertions", assert_calls: go: ["assert.That"]}
	}
	"TEST-03": {
		languages: ["c"]
		severity: "warning"
		title:   "No gets"
		message: "gets has no bound"
		params: label: "fixed"
		check: {kind: "token_ban", patterns: c: ["\\bgets\\s*\\("]}
	}
}
