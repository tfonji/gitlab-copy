package copy

import "fmt"

// ptrBoolStr safely dereferences a *bool for display, returning "<nil>" if nil.
func ptrBoolStr(b *bool) string {
	if b == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%v", *b)
}
