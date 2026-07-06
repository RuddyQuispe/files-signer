package cli

import (
	"flag"
	"strings"
)

// reorder moves flags before positional arguments so the user can write them in
// any order (Go's flag package otherwise stops at the first positional). It uses
// the FlagSet to know which flags are booleans and thus take no value.
func reorder(fs *flag.FlagSet, args []string) []string {
	bools := map[string]bool{}
	fs.VisitAll(func(f *flag.Flag) {
		if bf, ok := f.Value.(interface{ IsBoolFlag() bool }); ok && bf.IsBoolFlag() {
			bools[f.Name] = true
		}
	})

	var flags, positional []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--":
			positional = append(positional, args[i+1:]...)
			i = len(args)
		case a != "-" && strings.HasPrefix(a, "-"):
			name := strings.TrimLeft(a, "-")
			hasInlineValue := strings.ContainsRune(name, '=')
			name, _, _ = strings.Cut(name, "=")
			flags = append(flags, a)
			if !hasInlineValue && !bools[name] && i+1 < len(args) {
				i++
				flags = append(flags, args[i]) // consume the flag's value
			}
		default:
			positional = append(positional, a)
		}
	}
	return append(flags, positional...)
}
