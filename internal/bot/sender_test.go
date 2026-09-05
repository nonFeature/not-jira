package bot

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"not-jira/internal/emoji"

	"github.com/mymmrac/telego"
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

func TestMakeInlineButtonAndStrip(t *testing.T) {
	btn := emoji.MakeInlineButton("Саб-таск", "call_1", "", emoji.ID_PLUS, "➕", "primary")
	if emoji.IsCustomEmojiEnabled() {
		if btn.Text != "Саб-таск" {
			t.Errorf("expected text 'Саб-таск', got %q", btn.Text)
		}
		if btn.IconCustomEmojiID != emoji.ID_PLUS {
			t.Errorf("expected IconCustomEmojiID %q, got %q", emoji.ID_PLUS, btn.IconCustomEmojiID)
		}
		data, err := json.Marshal(btn)
		if err != nil {
			t.Fatalf("json marshal failed: %v", err)
		}
		jsonStr := string(data)
		if !strings.Contains(jsonStr, `"icon_custom_emoji_id":"5251309384419548860"`) {
			t.Errorf("expected json to contain icon_custom_emoji_id, got %s", jsonStr)
		}
		if !strings.Contains(jsonStr, `"text":"Саб-таск"`) {
			t.Errorf("expected json to contain text 'Саб-таск', got %s", jsonStr)
		}
	}

	ikm := &telego.InlineKeyboardMarkup{
		InlineKeyboard: [][]telego.InlineKeyboardButton{{btn}},
	}
	StripKeyboardCustomEmojis(ikm)
	strippedBtn := ikm.InlineKeyboard[0][0]
	if strippedBtn.IconCustomEmojiID != "" {
		t.Errorf("expected empty IconCustomEmojiID, got %q", strippedBtn.IconCustomEmojiID)
	}
	if strippedBtn.Text != "➕ Саб-таск" {
		t.Errorf("expected text '➕ Саб-таск', got %q", strippedBtn.Text)
	}
}

func TestCleanAlertText(t *testing.T) {
	input := `<tg-emoji emoji-id="5251750898467649056">✅</tg-emoji> @alice принял задачу <b>[B0]</b>.`
	expected := `✅ @alice принял задачу [B0].`
	result := cleanAlertText(input)
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

