package browser

import "testing"

// TestGetActiveTab exercises the public GetActiveTab wrapper.
func TestGetActiveTab(t *testing.T) {
    m := New()
    if m.GetActiveTab() == nil {
        t.Error("expected non-nil active tab from New()")
    }

    m.Tabs = nil
    if m.GetActiveTab() != nil {
        t.Error("expected nil active tab when Tabs is nil")
    }
}
