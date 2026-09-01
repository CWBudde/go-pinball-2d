package platform

import (
	"errors"
	"testing"
	"unicode/utf8"
)

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		value string
		limit int
		want  string
	}{
		{name: "unchanged", value: "short", limit: 5, want: "short"},
		{name: "newline", value: "one\ntwo", limit: 20, want: "one two"},
		{name: "ASCII", value: "abcdef", limit: 5, want: "ab..."},
		{name: "UTF-8", value: "éclair", limit: 5, want: "éc..."},
		{name: "three", value: "abcdef", limit: 3, want: "..."},
		{name: "two", value: "abcdef", limit: 2, want: ".."},
		{name: "one", value: "abcdef", limit: 1, want: "."},
		{name: "zero", value: "abcdef", limit: 0, want: ""},
		{name: "negative", value: "abcdef", limit: -1, want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := truncate(test.value, test.limit)
			if got != test.want {
				t.Fatalf("truncate(%q, %d) = %q, want %q", test.value, test.limit, got, test.want)
			}
			if !utf8.ValidString(got) {
				t.Fatalf("truncate(%q, %d) returned invalid UTF-8 %q", test.value, test.limit, got)
			}
		})
	}
}

func TestAppClearsStatusError(t *testing.T) {
	a := app{statusError: errors.New("transient")}
	a.clearStatusError()
	if a.statusError != nil {
		t.Fatalf("status error was not cleared: %v", a.statusError)
	}
}

func TestAppAttemptsIconOnlyOnce(t *testing.T) {
	a := app{}
	attempts := 0
	wantErr := errors.New("icon unavailable")
	setIcon := func(path string) error {
		attempts++
		if path != "assets/images/favicon.png" {
			t.Fatalf("unexpected icon path %q", path)
		}
		return wantErr
	}

	if err := a.setIconOnce(setIcon); !errors.Is(err, wantErr) {
		t.Fatalf("first attempt error = %v, want %v", err, wantErr)
	}
	if err := a.setIconOnce(setIcon); err != nil {
		t.Fatalf("second attempt error = %v, want nil", err)
	}
	if attempts != 1 {
		t.Fatalf("SetIcon called %d times, want 1", attempts)
	}
}

func TestRecoverJSValue(t *testing.T) {
	if got := recoverJSValue(func() int { return 42 }); got != 42 {
		t.Fatalf("successful operation returned %d, want 42", got)
	}
	if got := recoverJSValue(func() int { panic("storage denied") }); got != 0 {
		t.Fatalf("panicking operation returned %d, want zero value", got)
	}
}
