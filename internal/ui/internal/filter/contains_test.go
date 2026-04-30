package filter

import (
	"reflect"
	"testing"
)

func TestContainsFold(t *testing.T) {
	id := func(s string) string { return s }
	type tc struct {
		name  string
		items []string
		q     string
		want  []string
	}
	cases := []tc{
		{name: "nil input + empty query", items: nil, q: "", want: nil},
		{name: "empty query returns input header", items: []string{"a", "b"}, q: "", want: []string{"a", "b"}},
		{name: "all match", items: []string{"alpha", "beta"}, q: "a", want: []string{"alpha", "beta"}},
		{name: "no match", items: []string{"alpha", "beta"}, q: "zz", want: []string{}},
		{name: "case fold", items: []string{"Alpha", "BETA"}, q: "be", want: []string{"BETA"}},
		{name: "preserves order", items: []string{"x1", "x2", "x3"}, q: "x", want: []string{"x1", "x2", "x3"}},
		{name: "mixed match/no-match", items: []string{"keep", "skip", "Keeper"}, q: "keep", want: []string{"keep", "Keeper"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := ContainsFold(c.items, c.q, id)
			if !reflect.DeepEqual(got, c.want) {
				t.Fatalf("got %v, want %v", got, c.want)
			}
		})
	}
}

func TestContainsFold_EmptyQueryReturnsSameSliceHeader(t *testing.T) {
	src := []string{"a", "b", "c"}
	got := ContainsFold(src, "", func(s string) string { return s })
	if &got[0] != &src[0] {
		t.Fatalf("empty query should return input slice header without allocation")
	}
}

type modelLike struct{ Name, Path string }

func TestContainsFold_GenericKey(t *testing.T) {
	src := []modelLike{{Name: "Llama-3"}, {Name: "Mistral"}, {Name: "lLAMA-2"}}
	got := ContainsFold(src, "lla", func(m modelLike) string { return m.Name })
	want := []modelLike{{Name: "Llama-3"}, {Name: "lLAMA-2"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v, want %v", got, want)
	}
}
