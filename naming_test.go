package bear

import (
	"reflect"
	"testing"
)

func Test_parseWords(t *testing.T) {
	tests := []struct {
		name      string
		s         string
		wantWords []string
	}{
		{name: "helloWorld", s: "helloWorld", wantWords: []string{"hello", "world"}},
		{name: "HelloWorld", s: "HelloWorld", wantWords: []string{"hello", "world"}},
		{name: "hello_world", s: "hello_world", wantWords: []string{"hello", "world"}},
		{name: "hello__world", s: "hello__world", wantWords: []string{"hello", "world"}},
		{name: "hello world", s: "hello world", wantWords: []string{"hello", "world"}},
		{name: "hello  world", s: "hello  world", wantWords: []string{"hello", "world"}},
		{name: "HELLOWorld", s: "HELLOWorld", wantWords: []string{"hello", "world"}},
		{name: "helloWORLD", s: "helloWORLD", wantWords: []string{"hello", "world"}},
		{name: "helloWORLDHello", s: "helloWORLDHello", wantWords: []string{"hello", "world", "hello"}},
		{name: "hello+world", s: "hello+world", wantWords: []string{"hello+world"}},
		{name: "你好世界", s: "你好世界", wantWords: []string{"你好世界"}},
		{name: "你好_世界", s: "你好_世界", wantWords: []string{"你好", "世界"}},
		{name: "你好 世界", s: "你好 世界", wantWords: []string{"你好", "世界"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotWords := parseWords(tt.s); !reflect.DeepEqual(gotWords, tt.wantWords) {
				t.Errorf("parseWords() = %v, want %v", gotWords, tt.wantWords)
			}
		})
	}
}
