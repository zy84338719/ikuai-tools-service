package job

import (
	"reflect"
	"testing"
)

func TestSanitizeTag(t *testing.T) {
	tests := []struct{ in, want string }{
		{"cn", "cn"},
		{"my-tag!", "mytag"},
		{"国内", "国内"},
		{"a b_c", "abc"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := sanitizeTag(tt.in); got != tt.want {
			t.Errorf("sanitizeTag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestBuildTagName(t *testing.T) {
	if got := buildTagName("cn"); got != "IKBcn" {
		t.Errorf("buildTagName(cn) = %q, want IKBcn", got)
	}
	if got := buildTagName("国内"); got != "IKB国内" {
		t.Errorf("buildTagName(国内) = %q, want IKB国内", got)
	}
}

func TestChunkCommentRoundTrip(t *testing.T) {
	for _, idx := range []int{1, 2, 42, 999} {
		c := buildChunkComment(idx)
		got := parseChunkComment(c)
		if got != idx {
			t.Errorf("parseChunkComment(buildChunkComment(%d)=%q) = %d, want %d", idx, c, got, idx)
		}
	}
}

func TestParseChunkCommentInvalid(t *testing.T) {
	for _, c := range []string{"", "ikb:0", "ikb:-1", "ikb:abc", "other:1"} {
		if got := parseChunkComment(c); got != 0 {
			t.Errorf("parseChunkComment(%q) = %d, want 0", c, got)
		}
	}
}

func TestStreamDomainCommentRoundTrip(t *testing.T) {
	for _, idx := range []int{1, 5, 100} {
		c := buildStreamDomainComment("proxy", idx)
		got := parseStreamDomainComment(c, "proxy")
		if got != idx {
			t.Errorf("parseStreamDomainComment(%q, proxy) = %d, want %d", c, got, idx)
		}
	}
	// Wrong tag must not match.
	c := buildStreamDomainComment("proxy", 1)
	if got := parseStreamDomainComment(c, "other"); got != 0 {
		t.Errorf("parseStreamDomainComment with mismatched tag = %d, want 0", got)
	}
}

func TestSplitChunks(t *testing.T) {
	// Empty input.
	if got := splitChunks(nil, 3); got != nil {
		t.Errorf("splitChunks(nil,3) = %v, want nil", got)
	}
	// Even division.
	items := []string{"a", "b", "c", "d", "e", "f"}
	got := splitChunks(items, 2)
	want := [][]string{{"a", "b"}, {"c", "d"}, {"e", "f"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitChunks even = %v, want %v", got, want)
	}
	// Remainder in the last chunk.
	got = splitChunks(items, 4)
	want = [][]string{{"a", "b", "c", "d"}, {"e", "f"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("splitChunks remainder = %v, want %v", got, want)
	}
	// chunkSize <= 0 → default 5000.
	got = splitChunks(items, 0)
	if len(got) != 1 || len(got[0]) != 6 {
		t.Errorf("splitChunks(,0) = %v, want single chunk of 6", got)
	}
	// Chunks must not alias the input slice (mutations stay independent).
	chunks := splitChunks(items, 2)
	chunks[0][0] = "X"
	if items[0] != "a" {
		t.Errorf("splitChunks chunk aliases input: items[0]=%q, want a", items[0])
	}
}

func TestDedupe(t *testing.T) {
	in := []string{"a", "b", "a", "c", "b", "d"}
	got := dedupe(in)
	want := []string{"a", "b", "c", "d"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("dedupe = %v, want %v", got, want)
	}
	// Empty / single.
	if got := dedupe(nil); len(got) != 0 {
		t.Errorf("dedupe(nil) = %v, want empty", got)
	}
}
