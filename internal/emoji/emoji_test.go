package emoji

import (
	"errors"
	"strings"
	"testing"

	"github.com/mymmrac/telego"
)

func TestE(t *testing.T) {
	SetCustomEmojiEnabled(true)
	defer SetCustomEmojiEnabled(true)

	// When empty ID, returns fallback
	if res := E("", "✅"); res != "✅" {
		t.Errorf("expected %q, got %q", "✅", res)
	}

	// When ID is set, returns tag
	if res := E("12345", "✅"); res != `<tg-emoji emoji-id="12345">✅</tg-emoji>` {
		t.Errorf("expected tag, got %q", res)
	}

	// When disabled, returns fallback
	SetCustomEmojiEnabled(false)
	if res := E("12345", "✅"); res != "✅" {
		t.Errorf("expected fallback when disabled, got %q", res)
	}
}

func TestCheckAndUpdateFromMessage(t *testing.T) {
	SetCustomEmojiEnabled(true)
	defer SetCustomEmojiEnabled(true)

	// Message without custom emoji when hadCustomEmoji was true -> should disable
	msgNoCustom := &telego.Message{
		Entities: []telego.MessageEntity{
			{Type: "bold"},
		},
	}
	CheckAndUpdateFromMessage(msgNoCustom, true)
	if IsCustomEmojiEnabled() {
		t.Errorf("expected custom emojis to be disabled")
	}

	// Reset
	SetCustomEmojiEnabled(true)

	// Message WITH custom_emoji entity -> should stay enabled
	msgWithCustom := &telego.Message{
		Entities: []telego.MessageEntity{
			{Type: "custom_emoji"},
		},
	}
	CheckAndUpdateFromMessage(msgWithCustom, true)
	if !IsCustomEmojiEnabled() {
		t.Errorf("expected custom emojis to remain enabled")
	}
}

func TestStripCustomEmojis(t *testing.T) {
	input := `<tg-emoji emoji-id="123">🐛</tg-emoji> hello <tg-emoji emoji-id="456">💡</tg-emoji>`
	expected := `🐛 hello 💡`
	if res := StripCustomEmojis(input); res != expected {
		t.Errorf("expected %q, got %q", expected, res)
	}
}

func TestIsCustomEmojiError(t *testing.T) {
	err := errors.New("Bad Request: can't use custom emoji")
	if !IsCustomEmojiError(err) {
		t.Errorf("expected true")
	}
}

func TestBoxAndUser(t *testing.T) {
	SetCustomEmojiEnabled(false)
	if Box() != "📦" {
		t.Errorf("expected 📦 when disabled, got %q", Box())
	}
	if User() != "👤" {
		t.Errorf("expected 👤 when disabled, got %q", User())
	}
	if Tag() != "🏷️" {
		t.Errorf("expected 🏷️ when disabled, got %q", Tag())
	}
	if Priority() != "⚡️" {
		t.Errorf("expected ⚡️ when disabled, got %q", Priority())
	}

	SetCustomEmojiEnabled(true)
	defer SetCustomEmojiEnabled(true)

	if !strings.Contains(Box(), "5251483382134646909") {
		t.Errorf("expected custom emoji for Box, got %q", Box())
	}
	if !strings.Contains(User(), "5253956862390345938") {
		t.Errorf("expected custom emoji for User, got %q", User())
	}
	if !strings.Contains(Tag(), "5253876499257271563") {
		t.Errorf("expected custom emoji for Tag, got %q", Tag())
	}
	if !strings.Contains(Priority(), "5253449291745241072") {
		t.Errorf("expected custom emoji for Priority, got %q", Priority())
	}
	if !strings.Contains(P0(), "5251665716381262011") {
		t.Errorf("expected custom emoji for P0, got %q", P0())
	}

	// Test FallbackForID
	if FallbackForID(ID_CHECK) != "✅" {
		t.Errorf("expected ✅ for ID_CHECK, got %q", FallbackForID(ID_CHECK))
	}
	// Test FallbackForID with configured ID
	ID_TRASH = "trash_123"
	defer func() { ID_TRASH = "" }()
	if FallbackForID("trash_123") != "🗑" {
		t.Errorf("expected 🗑 for ID_TRASH, got %q", FallbackForID("trash_123"))
	}
	ID_WRITE = "write_123"
	defer func() { ID_WRITE = "" }()
	if FallbackForID("write_123") != "✍️" {
		t.Errorf("expected ✍️ for ID_WRITE, got %q", FallbackForID("write_123"))
	}
	if FallbackForID("") != "" {
		t.Errorf("expected empty string for empty ID, got %q", FallbackForID(""))
	}
	if FallbackForID("unknown_id") != "" {
		t.Errorf("expected empty string for unknown ID, got %q", FallbackForID("unknown_id"))
	}

	ID_TRASH = ""
	ID_WRITE = ""

	// Test new accessors when ID is empty (fallback to unicode)
	if Trash() != "🗑" {
		t.Errorf("expected 🗑, got %q", Trash())
	}
	if Write() != "✍️" {
		t.Errorf("expected ✍️, got %q", Write())
	}
}

