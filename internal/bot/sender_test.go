package bot

import (
	"errors"
	"testing"

	"not-jira/internal/emoji"
)

func TestStripCustomEmojis(t *testing.T) {
	input := `<b>[B0]</b> <tg-emoji emoji-id="5368324170671202286">🐛</tg-emoji> <b>Баг: Что-то сломалось</b>`
	expected := `<b>[B0]</b> 🐛 <b>Баг: Что-то сломалось</b>`

	result := emoji.StripCustomEmojis(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}

	multiple := `<tg-emoji emoji-id="1">⬜️</tg-emoji> Задача 1
<tg-emoji emoji-id="2">☑️</tg-emoji> Задача 2`
	expectedMultiple := `⬜️ Задача 1
☑️ Задача 2`
	if emoji.StripCustomEmojis(multiple) != expectedMultiple {
		t.Errorf("expected %q, got %q", expectedMultiple, emoji.StripCustomEmojis(multiple))
	}
}

func TestIsCustomEmojiError(t *testing.T) {
	err1 := errors.New("Bad Request: can't use custom emoji in message")
	if !emoji.IsCustomEmojiError(err1) {
		t.Errorf("expected true for %v", err1)
	}

	err2 := errors.New("Bad Request: ENTITY_CUSTOM_EMOJI_INVALID")
	if !emoji.IsCustomEmojiError(err2) {
		t.Errorf("expected true for %v", err2)
	}

	err3 := errors.New("Bad Request: chat not found")
	if emoji.IsCustomEmojiError(err3) {
		t.Errorf("expected false for %v", err3)
	}
}
