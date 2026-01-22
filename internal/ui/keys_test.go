package ui

import (
	"testing"
)

func TestDefaultKeyMapHasQuit(t *testing.T) {
	km := DefaultKeyMap()
	if len(km.Quit.Keys()) == 0 {
		t.Fatal("KeyMap should have quit key")
	}
}

func TestDefaultKeyMapHasNavigation(t *testing.T) {
	km := DefaultKeyMap()
	if len(km.Up.Keys()) == 0 {
		t.Fatal("KeyMap should have up key")
	}
	if len(km.Down.Keys()) == 0 {
		t.Fatal("KeyMap should have down key")
	}
}

func TestKeyMapShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	help := km.ShortHelp()
	if len(help) == 0 {
		t.Fatal("ShortHelp should return bindings")
	}
}

func TestDefaultKeyMapSearchNavigation(t *testing.T) {
	km := DefaultKeyMap()

	if len(km.NextMatch.Keys()) == 0 {
		t.Error("NextMatch keybinding should be defined")
	}
	if km.NextMatch.Keys()[0] != "n" {
		t.Errorf("NextMatch should be 'n', got %q", km.NextMatch.Keys()[0])
	}

	if len(km.PrevMatch.Keys()) == 0 {
		t.Error("PrevMatch keybinding should be defined")
	}
	if km.PrevMatch.Keys()[0] != "N" {
		t.Errorf("PrevMatch should be 'N', got %q", km.PrevMatch.Keys()[0])
	}
}

func TestKeyMapFullHelp(t *testing.T) {
	km := DefaultKeyMap()

	fullHelp := km.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp() should return non-empty help")
	}

	// Should contain key bindings
	// Format is [][]key.Binding
	for i, section := range fullHelp {
		if len(section) == 0 {
			t.Errorf("FullHelp() section %d is empty", i)
		}
	}
}
