package emoji

import (
	"errors"
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
