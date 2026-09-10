package bot

import (
	"testing"
	"time"
)

func TestMemberCacheGetSet(t *testing.T) {
	c := newMemberCache()

	if _, ok := c.get(1); ok {
		t.Fatalf("expected empty cache miss for user 1")
	}

	c.set(1, true)
	c.set(2, false)

	if isMember, ok := c.get(1); !ok || !isMember {
		t.Errorf("expected cached member for user 1, got member=%v ok=%v", isMember, ok)
	}
	if isMember, ok := c.get(2); !ok || isMember {
		t.Errorf("expected cached non-member for user 2, got member=%v ok=%v", isMember, ok)
	}
}

func TestMemberCacheExpiry(t *testing.T) {
	c := newMemberCache()
	c.ttl = -time.Second
	c.set(1, true)

	if _, ok := c.get(1); ok {
		t.Errorf("expected expired cache entry to be a miss")
	}

	if cleaned := c.cleanup(time.Millisecond); cleaned != 1 {
		t.Errorf("expected 1 expired entry cleaned, got %d", cleaned)
	}
	if _, ok := c.get(1); ok {
		t.Errorf("expected cleaned cache entry to be a miss")
	}
}
