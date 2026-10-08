package realtime

import (
	"strconv"
	"testing"
)

var benchmarkTopicSink string

func BenchmarkConversationTopic(b *testing.B) {
	b.ReportAllocs()

	var result string

	for i := 0; i < b.N; i++ {
		result = ConversationTopic(uint(i%1000 + 1))
	}

	benchmarkTopicSink = result
}

func BenchmarkConversationTopicStrconv(b *testing.B) {
	b.ReportAllocs()

	var result string

	for i := 0; i < b.N; i++ {
		result = conversationTopicPrefix +
			strconv.FormatUint(uint64(i%1000+1), 10)
	}

	benchmarkTopicSink = result
}

func TestConversationTopic(t *testing.T) {
	tests := []struct {
		id   uint
		want string
	}{
		{0, "gohub:conversation:0"},
		{1, "gohub:conversation:1"},
		{4001, "gohub:conversation:4001"},
		{9999, "gohub:conversation:9999"},
	}

	for _, tt := range tests {
		got := ConversationTopic(tt.id)
		if got != tt.want {
			t.Errorf(
				"ConversationTopic(%d) = %q, want %q",
				tt.id,
				got,
				tt.want,
			)
		}
	}
}

func TestConversationTopicRoundTrip(t *testing.T) {
	ids := []uint{1, 1001, 4001, 4003}

	for _, want := range ids {
		topic := ConversationTopic(want)

		got, err := ParseConversationTopic(topic)
		if err != nil {
			t.Fatalf(
				"ParseConversationTopic(%q): %v",
				topic,
				err,
			)
		}

		if got != want {
			t.Errorf(
				"got %d, want %d",
				got,
				want,
			)
		}
	}
}

func TestParseConversationTopicInvalid(t *testing.T) {
	cases := []string{
		"",
		"invalid:conversation:4001",
		"gohub:conversation:",
		"gohub:conversation:abc",
		"gohub:conversation:0",
		"gohub:conversation:-1",
	}

	for _, topic := range cases {
		_, err := ParseConversationTopic(topic)

		if err == nil {
			t.Errorf(
				"expected error for topic %q",
				topic,
			)
		}
	}
}
