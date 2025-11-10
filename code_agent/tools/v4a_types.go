package tools

// V4APatch represents a parsed V4A format patch.
// V4A is a semantic patch format that uses context markers (class/function names)
// instead of line numbers, making it more resilient to code changes.
//
// Example V4A format:
//
//	*** Update File: src/models/user.py
//	@@ class User
//	@@     def validate():
//	-          return True
//	+          if not self.email:
//	+              raise ValueError("Email required")
//	+          return True
type V4APatch struct {
	// FilePath is the target file to patch (extracted from "*** Update File:" header)
	FilePath string

	// Hunks are the individual changes to apply
	Hunks []V4AHunk
}

// V4AHunk represents a single change block in a V4A patch.
type V4AHunk struct {
	// ContextMarkers define the semantic location (e.g., ["class User", "def validate()"])
	// These are extracted from @@ prefix lines and used to find the change location
	ContextMarkers []string

	// Removals are lines to remove (originally prefixed with -)
	Removals []string

	// Additions are lines to add (originally prefixed with +)
	Additions []string

	// BaseIndentation is the indentation level from the deepest @@ context marker
	// Used to preserve indentation when applying changes
	BaseIndentation int
}
