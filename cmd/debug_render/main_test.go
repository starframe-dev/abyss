package main

import "testing"

func TestStripANSI(t *testing.T) {
    cases := []struct {
        name string
        in   string
        want string
    }{
        {"empty", "", ""},
        {"plain", "hello world", "hello world"},
        {"color", "\033[31mred\033[0m", "red"},
        {"bold", "\033[1;33mbold yellow\033[0m", "bold yellow"},
        {"multiple", "\033[31m\033[1mhi\033[0m there", "hi there"},
    }
    for _, c := range cases {
        t.Run(c.name, func(t *testing.T) {
            got := stripANSI(c.in)
            if got != c.want {
                t.Errorf("stripANSI(%q) = %q, want %q", c.in, got, c.want)
            }
        })
    }
}
