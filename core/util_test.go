package core

import (
	"testing"
)

func TestHash5Sum(t *testing.T) {
	for _, tt := range []struct {
		src  interface{}
		want string
	}{
		{"123456", "e10adc3949ba59abbe56e057f20f883e"},
		{[]byte("123456"), "e10adc3949ba59abbe56e057f20f883e"},
		{123456, "e10adc3949ba59abbe56e057f20f883e"},
	} {
		if got := md5sum(tt.src); got != tt.want {
			t.Errorf("want %s, got %s", tt.want, got)
		}
	}

	for _, tt := range []struct {
		src  interface{}
		want string
	}{
		{"123456", "7c4a8d09ca3762af61e59520943dc26494f8941b"},
		{[]byte("123456"), "7c4a8d09ca3762af61e59520943dc26494f8941b"},
		{123456, "7c4a8d09ca3762af61e59520943dc26494f8941b"},
	} {
		if got := sha1sum(tt.src); got != tt.want {
			t.Errorf("want %s, got %s", tt.want, got)
		}
	}
}
