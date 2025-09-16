package logging

// Filter defines a transformation or condition on log events
type Filter interface {
	Apply(event LogEvent) (*LogEvent, bool)
}

// --------------------
// NamespaceFilter
// --------------------
type NamespaceFilter struct {
	Excluded []string
}

func (f *NamespaceFilter) Apply(event LogEvent) (*LogEvent, bool) {
	for _, ns := range f.Excluded {
		if event.Namespace == ns {
			return nil, false // drop the event
		}
	}
	return &event, true
}

// --------------------
// KeywordFilter
// --------------------
type KeywordFilter struct {
	Keyword string
}

func (f *KeywordFilter) Apply(event LogEvent) (*LogEvent, bool) {
	if f.Keyword != "" && !contains(event.Line, f.Keyword) {
		return nil, false
	}
	return &event, true
}

func contains(line, substr string) bool {
	return len(substr) > 0 && (len(line) >= len(substr)) && (stringContains(line, substr))
}

// Simple substring check
func stringContains(s, sub string) bool {
	return len(sub) <= len(s) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
