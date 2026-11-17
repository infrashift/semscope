// The schema of a semscope ruleset. The loader unifies each ruleset source
// with #Ruleset. ADR-0002 and the spec "CUE ruleset schema" describe it.
package ruleset

// #Ruleset is one ruleset source: a base ruleset or an overlay.
#Ruleset: {
	name?:        string
	description?: string
	rules: [ID=string]: #Rule & {id: ID}
}

// #Language names a language that the parser reads.
#Language: "go" | "c" | "cpp" | "zig" | "python" | "rust"

// #PerLanguage maps a language to a list of names or patterns.
#PerLanguage: [#Language]: [...string]

// #Rule is one rule. A tunable value needs a CUE default in the base
// ruleset, for example `max_lines: *60 | int`, so that an overlay can
// replace it. The schema gives no defaults: two different defaults cancel
// each other in CUE. The loader sets enabled to true and severity to
// "error" when a rule omits them.
#Rule: {
	id:        =~"^[A-Za-z][A-Za-z0-9_]*(-[A-Za-z0-9_]+)*$"
	enabled?:  bool
	severity?: "error" | "warning" | "note"
	languages: [#Language, ...#Language]
	title:   string & !=""
	message: string & !=""
	params: [string]: int | string | bool | [...string]
	check: #Check

	if check.kind == "max_function_lines" {
		params: max_lines: int & >0
	}
	if check.kind == "min_assertions" {
		params: {
			min_assertions: int & >=0
			mode:           "average" | "per_function"
		}
	}
	if check.kind == "max_pointer_indirection" {
		params: {
			max_pointer_depth:       int & >=0
			allow_function_pointers: bool
		}
	}
	if check.kind == "restricted_preprocessor" {
		params: max_conditional_depth: int & >=0
	}
}

// #Check selects the checker of a rule through the field kind.
#Check: #NoGotoOrRecursion | #BoundedLoops | #NoDynamicAlloc | #MaxFunctionLines |
	#MinAssertions | #NarrowestScope | #CheckedReturns | #RestrictedPreprocessor |
	#MaxPointerIndirection | #ASTQuery | #TokenBan

#NoGotoOrRecursion: {kind: "no_goto_or_recursion", banned_calls?: #PerLanguage}
#BoundedLoops: {kind: "bounded_loops"}
#NoDynamicAlloc: {kind: "no_dynamic_alloc", alloc_calls?: #PerLanguage, init_functions?: [...string]}
#MaxFunctionLines: {kind: "max_function_lines"}
#MinAssertions: {kind: "min_assertions", assert_calls?: #PerLanguage}
#NarrowestScope: {kind: "narrowest_scope"}
// must_check_calls names calls whose result the code must use. The parser
// has no type information, so the list stands in for "non-void" in C and C++.
#CheckedReturns: {kind: "checked_returns", must_check_calls?: #PerLanguage}
#RestrictedPreprocessor: {kind: "restricted_preprocessor", banned_directives?: #PerLanguage}
#MaxPointerIndirection: {kind: "max_pointer_indirection"}

// #ASTQuery runs one Tree-sitter query for each language. Each capture
// named @violation gives one finding.
#ASTQuery: {kind: "ast_query", query: [#Language]: string}

// #TokenBan bans regular expressions in the source without comments and strings.
#TokenBan: {kind: "token_ban", patterns: #PerLanguage}
