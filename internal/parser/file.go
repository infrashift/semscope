// Package parser reads source files with Tree-sitter and gives a flat record
// of each file to the engine.
//
// The walker visits each syntax tree with a Tree-sitter cursor, without
// recursion. One adapter for each language turns the nodes into functions,
// calls and sites. The engine checks the record. It does not read the tree,
// except for ast_query rules.
package parser

// Pos is a 1-based source range. A column counts bytes.
type Pos struct {
	Line    int
	Col     int
	EndLine int
	EndCol  int
}

// Function is one function or method with a body.
type Function struct {
	// Name is the plain name. Key is the call graph name: "Type.name" for a
	// Go method, Scope + "." + Name for a Zig function in a container, else
	// Name.
	Name string
	Key  string
	// Scope is the container path of a Zig function, outermost first and
	// separated by ".". A function of `const Mutable = struct {...}` has the
	// scope "Mutable". A function of the struct that the function ArrayList
	// returns has the scope "ArrayList()". A function of the file has "".
	Scope string
	// Recv is the receiver variable of a Go method, else "".
	Recv string
	// Pos runs from the first line of the signature to the closing brace.
	Pos Pos
	// Exempt marks a test function or a test helper. It is exempt from the
	// recursion, function length and assertion checks.
	Exempt bool
}

// Call is one call, macro invocation or assertion statement.
type Call struct {
	// Fn is the index of the enclosing function in File.Functions, or -1.
	Fn int
	// Callee is the text of the called expression, for example
	// "assert.That", "std.debug.assert", "Box::new" or "assert!".
	Callee string
	// Name is the last segment of Callee. Qual is the segment before it.
	Name string
	Qual string
	// Unused marks a call that is a whole expression statement.
	Unused bool
	Pos    Pos
}

// SiteKind is the kind of a construct that a check can report.
type SiteKind int

// The site kinds.
const (
	SiteGoto SiteKind = iota + 1
	SiteUnboundedLoop
	SiteDiscard
	SiteGlobal
	SiteDirective
	SiteConditional
	SitePointer
	SiteFuncPointer
)

// Site is one construct that a check can report.
type Site struct {
	Kind SiteKind
	// Detail describes the construct, for example "`for {}`" or a directive name.
	Detail string
	// Depth is the pointer depth or the conditional compilation depth.
	Depth int
	Fn    int
	Pos   Pos
}

// File is the flat record of one source file.
type File struct {
	Path     string
	Lang     Language
	Package  string
	Exempt   bool
	Lines    []string
	Stripped []string
	// Functions, Calls and Sites follow the source sequence.
	Functions []Function
	Calls     []Call
	Sites     []Site
	// Imports maps a Zig import name to the imported path, as written.
	Imports map[string]string
}
