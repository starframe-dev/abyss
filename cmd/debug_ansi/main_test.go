package main

import (
    "bytes"
    "testing"
)

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

func TestMin(t *testing.T) {
    cases := []struct {
        a, b, want int
    }{
        {1, 2, 1},
        {5, 3, 3},
        {0, 0, 0},
        {-1, 1, -1},
        {100, 100, 100},
    }
    for _, c := range cases {
        got := min(c.a, c.b)
        if got != c.want {
            t.Errorf("min(%d, %d) = %d, want %d", c.a, c.b, got, c.want)
        }
    }
}

func TestContainsAByss(t *testing.T) {
    log := "Hello Abyss, this is a test"
    if !bytes.Contains([]byte(log), []byte("Abyss")) {
        t.Error("expected 'Abyss' substring to be found")
    }
}
