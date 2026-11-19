// The vocabulary of the semscope repository: technical terms of this project
// that the lexiscope rules STE-100-06 and STE-100-07 do not report. The
// docs-lexiscope gate loads this file over the built-in ste-100 ruleset of
// lexiscope:
//
//	lexiscope scan --ruleset ste-100 --ruleset vocabulary.cue ...
//
// Add a term only when it names a concept of the project, a tool or a
// standard. Rewrite other prose.
vocabulary: {
	// Project process: PLAN.md, ADRs and issues.
	acceptance:  true // acceptance criteria of an issue
	decision:    true // architecture decision record
	development: true // the development workflow
	resolution:  true // the Resolution section of an issue

	// The Power of 10 rules and the nasa-p10 ruleset.
	allocation:     true // Power of 10 rule R3
	initialization: true // Power of 10 rule R3
	generation:     true // code generation, Power of 10 rule R8
	compilation:    true // conditional compilation, Power of 10 rule R8
	indirection:    true // Power of 10 rule R9
	deviation:      true // a Power of 10 rule that semscope does not follow
	suppression:    true // a semscope or nasa-lint allow comment
	violation:      true // a finding, and the violation fixtures

	// Parity with the oracles (ADR-0005).
	divergence: true // a record in parity-divergences.json
	detection:  true // detection parity with the oracles

	// Source code that the checks read.
	enclosing: true // the enclosing function or container
	opening:   true // the opening line or brace of a function
	closing:   true // the closing brace of a function

	// CUE and the rulesets.
	unification: true // CUE unification
	composition: true // the composition of rulesets
	disjunction: true // a CUE disjunction
	embedding:   true // a CUE embedding

	// SARIF, the CLI and the docs site.
	validation:  true // SARIF schema validation
	scanning:    true // code scanning tools, and the Scanning guide
	integration: true // CI integration, and its guide
	completion:  true // the cobra completion command
	landing:     true // the landing page of the docs site
	marketing:   true // the marketing words of the style rules
}
