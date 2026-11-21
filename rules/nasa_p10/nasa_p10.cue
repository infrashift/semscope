// The built-in nasa-p10 ruleset: the NASA/JPL Power of 10 rules of
// Gerard J. Holzmann ("The Power of 10: Rules for Developing
// Safety-Critical Code", 2006), adapted to seven languages.
//
// Each tunable value is a CUE default, so an overlay or --set can replace it.
// The checks of NASA-P10-01, -02, -04, -05, -07 and -09 match
// scripts/nasa-lint.py (Zig) and scripts/nasa-lint-go.py (Go). ADR-0005
// defines the parity.
package nasap10

name:        "nasa-p10"
description: "The NASA/JPL Power of 10 rules for safety-critical code, adapted to Go, C, C++, Zig, Python, Rust and C#."

rules: {
	"NASA-P10-01": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Simple control flow"
		message:   "Use simple control flow: no goto, no setjmp or longjmp, and no direct or indirect recursion."
		check: {
			kind: "no_goto_or_recursion"
			banned_calls: {
				c: *["setjmp", "longjmp", "sigsetjmp", "siglongjmp", "_setjmp", "_longjmp"] | [...string]
				cpp: *["setjmp", "longjmp", "sigsetjmp", "siglongjmp", "_setjmp", "_longjmp"] | [...string]
			}
		}
	}
	"NASA-P10-02": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Fixed loop bounds"
		message:   "Give each loop a fixed upper bound. This loop has no condition that can end it."
		check: kind: "bounded_loops"
	}
	"NASA-P10-03": {
		// Python and C# are not in the list: their objects come from the
		// runtime, so most code would break the rule (ADR-0016).
		languages: *["go", "c", "cpp", "zig", "rust"] | [...]
		// The check uses call names only, so it reports a warning.
		severity: *"warning" | "error" | "note"
		title:    "No dynamic allocation after initialization"
		message:  "Do not allocate memory after initialization. Move the allocation into an init function."
		check: {
			kind: "no_dynamic_alloc"
			alloc_calls: {
				go: *["make", "new", "append"] | [...string]
				c: *["malloc", "calloc", "realloc", "aligned_alloc", "strdup", "strndup"] | [...string]
				cpp: *["malloc", "calloc", "realloc", "aligned_alloc", "strdup", "strndup", "new"] | [...string]
				zig: *["alloc", "create", "dupe", "dupeZ", "realloc", "alignedAlloc", "allocSentinel"] | [...string]
				rust: *["Box::new", "Vec::new", "Vec::with_capacity", "String::new", "String::from", "vec!", "Rc::new", "Arc::new"] | [...string]
			}
			init_functions: *["main", "init", "init*", "Init*", "new", "New*", "setup*", "Setup*"] | [...string]
		}
	}
	"NASA-P10-04": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Short functions"
		message:   "Keep each function to max_lines lines or fewer, from the signature to the closing brace."
		params: max_lines: *60 | int
		check: kind: "max_function_lines"
	}
	"NASA-P10-05": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Assertion density"
		message:   "Use min_assertions or more assertions for each function. Assertions check programmer errors only."
		params: {
			min_assertions: *2 | int
			mode:           *"average" | "per_function"
		}
		check: {
			kind: "min_assertions"
			assert_calls: {
				go: *["assert.*"] | [...string]
				c: *["assert", "static_assert", "_Static_assert"] | [...string]
				cpp: *["assert", "static_assert"] | [...string]
				zig: *["assert"] | [...string]
				python: *["assert"] | [...string]
				rust: *["assert!", "assert_eq!", "assert_ne!", "debug_assert!", "debug_assert_eq!", "debug_assert_ne!"] | [...string]
				csharp: *["Debug.Assert", "Trace.Assert", "*.Debug.Assert", "*.Trace.Assert"] | [...string]
			}
		}
	}
	"NASA-P10-06": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Smallest scope"
		message:   "Declare data in the smallest possible scope. Do not add mutable global state."
		check: kind: "narrowest_scope"
	}
	"NASA-P10-07": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Check return values"
		message:   "Check the return value of each call. Do not discard or swallow an error."
		// go_types loads the Go packages with type information. It needs the
		// go command and a go.mod. nasa-lint-go.py has no such check, so the
		// default is false (ADR-0009).
		params: {
			go_types: *false | bool
			go_ignore_calls: *[
				"(*bytes.Buffer).Write", "(*bytes.Buffer).WriteByte", "(*bytes.Buffer).WriteRune", "(*bytes.Buffer).WriteString",
				"(*strings.Builder).Write", "(*strings.Builder).WriteByte", "(*strings.Builder).WriteRune", "(*strings.Builder).WriteString",
				"fmt.Fprint(*bytes.Buffer)", "fmt.Fprintf(*bytes.Buffer)", "fmt.Fprintln(*bytes.Buffer)",
				"fmt.Fprint(*strings.Builder)", "fmt.Fprintf(*strings.Builder)", "fmt.Fprintln(*strings.Builder)",
			] | [...string]
		}
		check: {
			kind: "checked_returns"
			must_check_calls: {
				c: *["malloc", "calloc", "realloc", "fopen", "fclose", "fread", "fwrite", "fgets", "read", "write", "close", "scanf", "sscanf", "fscanf", "snprintf", "pthread_create", "pthread_join", "pthread_mutex_lock"] | [...string]
				cpp: *["malloc", "calloc", "realloc", "fopen", "fclose", "fread", "fwrite", "fgets", "read", "write", "close", "scanf", "sscanf", "fscanf", "snprintf", "pthread_create", "pthread_join", "pthread_mutex_lock"] | [...string]
			}
		}
	}
	"NASA-P10-08": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Limited preprocessor"
		message:   "Limit the preprocessor and code generation to includes and simple constants."
		params: max_conditional_depth: *2 | int
		check: {
			kind: "restricted_preprocessor"
			banned_directives: {
				go: *["go:linkname"] | [...string]
				c: *["token_paste", "stringize", "variadic_macro", "undef"] | [...string]
				cpp: *["token_paste", "stringize", "variadic_macro", "undef"] | [...string]
				zig: *["@cImport"] | [...string]
				python: *["exec", "eval", "compile"] | [...string]
				rust: *["macro_rules!"] | [...string]
				csharp: *["define", "undef"] | [...string]
			}
		}
	}
	"NASA-P10-09": {
		languages: *["go", "c", "cpp", "zig", "rust", "csharp"] | [...]
		title:     "Restricted pointers"
		message:   "Use one level of pointer indirection at most. Do not use function pointers."
		params: {
			max_pointer_depth:       *1 | int
			allow_function_pointers: *false | bool
		}
		check: kind: "max_pointer_indirection"
	}
	"NASA-P10-10": {
		languages: *["go", "c", "cpp", "zig", "python", "rust", "csharp"] | [...]
		title:     "Zero warnings and no unsafe constructs"
		message:   "Do not use unsafe constructs. Do not suppress a compiler or linter warning without a reason."
		check: {
			kind: "ast_query"
			query: {
				go: *"""
					(import_spec path: (interpreted_string_literal) @violation (#eq? @violation "\\"unsafe\\""))
					((comment) @violation (#match? @violation "^//\\\\s*nolint(:[A-Za-z0-9_,-]+)?\\\\s*$"))
					""" | string
				c: *"""
					(call_expression function: (identifier) @violation (#any-of? @violation "gets" "strcpy" "strcat" "sprintf" "vsprintf"))
					((comment) @violation (#match? @violation "NOLINT(\\\\([^)]*\\\\))?\\\\s*(\\\\*/)?$"))
					""" | string
				cpp: *"""
					(call_expression function: (identifier) @violation (#any-of? @violation "gets" "strcpy" "strcat" "sprintf" "vsprintf"))
					((comment) @violation (#match? @violation "NOLINT(\\\\([^)]*\\\\))?\\\\s*(\\\\*/)?$"))
					""" | string
				zig: *"""
					((builtin_identifier) @violation (#any-of? @violation "@ptrFromInt" "@intFromPtr" "@ptrCast" "@alignCast"))
					""" | string
				python: *"""
					(import_statement name: (dotted_name) @violation (#eq? @violation "ctypes"))
					((comment) @violation (#match? @violation "#\\\\s*(noqa|type:\\\\s*ignore)\\\\s*$"))
					""" | string
				rust: *"""
					(unsafe_block) @violation
					""" | string
				csharp: *"""
					(unsafe_statement) @violation
					((modifier) @violation (#eq? @violation "unsafe"))
					((preproc_pragma) @violation (#match? @violation "^#\\\\s*pragma\\\\s+warning\\\\s+disable[^/]*$"))
					""" | string
			}
		}
	}
}
