package ruleset

rules: "DIR-01": {languages: ["rust"], title: "Bounded loops", message: "loop {} has no bound", check: kind: "bounded_loops"}
