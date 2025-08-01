package template

import (
	"testing"

	"github.com/jandedobbeleer/oh-my-posh/src/cache"
	"github.com/jandedobbeleer/oh-my-posh/src/maps"
	"github.com/jandedobbeleer/oh-my-posh/src/runtime/mock"

	"github.com/stretchr/testify/assert"
)

func TestRenderTemplate(t *testing.T) {
	type Me struct {
		Name string
	}

	cases := []struct {
		Context     any
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{
			Case:     "dot literal",
			Expected: "Hello .NET \uE77F",
			Template: "{{ .Text }} .NET \uE77F",
			Context:  struct{ Text string }{Text: "Hello"},
		},
		{
			Case:     "color override with dots",
			Expected: "😺💬<#FF8000> Meow! What should I do next? ...</>",
			Template: "😺💬<#FF8000> Meow! What should I do next? ...</>",
		},
		{
			Case:     "tillig's regex",
			Expected: " ⎈ hello :: world ",
			Template: " ⎈ {{ replaceP \"([a-f0-9]{2})[a-f0-9]{6}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{10}([a-f0-9]{2})\" .Context \"$1..$2\" }}{{ if .Namespace }} :: {{ .Namespace }}{{ end }} ", //nolint:lll
			Context: struct {
				Context   string
				Namespace string
			}{
				Context:   "hello",
				Namespace: "world",
			},
		},
		{
			Case:     "Env like property name",
			Expected: "hello world",
			Template: "{{.EnvLike}} {{.Text2}}",
			Context: struct {
				EnvLike string
				Text2   string
			}{
				EnvLike: "hello",
				Text2:   "world",
			},
		},
		{
			Case:     "single property with a dot literal",
			Expected: "hello world",
			Template: "{{ if eq .Text \".Net\" }}hello world{{ end }}",
			Context:  struct{ Text string }{Text: ".Net"},
		},
		{
			Case:     "single property",
			Expected: "hello world",
			Template: "{{.Text}} world",
			Context:  struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "duplicate property",
			Expected: "hello jan posh",
			Template: "hello {{ .Me.Name }} {{ .Name }}",
			Context: struct {
				Name string
				Me   Me
			}{
				Name: "posh",
				Me: Me{
					Name: "jan",
				},
			},
		},
		{
			Case:        "invalid property",
			ShouldError: true,
			Template:    "{{.Durp}} world",
			Context:     struct{ Text string }{Text: "hello"},
		},
		{
			Case:        "invalid template",
			ShouldError: true,
			Template:    "{{ if .Text }} world",
			Context:     struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "if statement true",
			Expected: "hello world",
			Template: "{{ if .Text }}{{.Text}} world{{end}}",
			Context:  struct{ Text string }{Text: "hello"},
		},
		{
			Case:     "if statement false",
			Expected: "world",
			Template: "{{ if .Text }}{{.Text}} {{end}}world",
			Context:  struct{ Text string }{Text: ""},
		},
		{
			Case:     "if statement true with 2 properties",
			Expected: "hello world",
			Template: "{{.Text}}{{ if .Text2 }} {{.Text2}}{{end}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
		{
			Case:     "if statement false with 2 properties",
			Expected: "hello",
			Template: "{{.Text}}{{ if .Text2 }} {{.Text2}}{{end}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text: "hello",
			},
		},
		{
			Case:     "double property template",
			Expected: "hello world",
			Template: "{{.Text}} {{.Text2}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
		{
			Case:     "sprig - contains",
			Expected: "hello world",
			Template: "{{ if contains \"hell\" .Text }}{{.Text}} {{end}}{{.Text2}}",
			Context: struct {
				Text  string
				Text2 string
			}{
				Text:  "hello",
				Text2: "world",
			},
		},
	}

	for _, tc := range cases {
		env := new(mock.Environment)
		env.On("Shell").Return("foo")
		Cache = new(cache.Template)
		Init(env, nil, nil)

		text, err := Render(tc.Template, tc.Context)
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}

		assert.NoError(t, err)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestRenderTemplateEnvVar(t *testing.T) {
	cases := []struct {
		Context     any
		Env         map[string]string
		Case        string
		Expected    string
		Template    string
		ShouldError bool
	}{
		{
			Case:        "nil struct with env var",
			ShouldError: true,
			Template:    "{{.Env.HELLO }} world{{ .Text}}",
			Context:     nil,
			Env:         map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "map with env var",
			Expected: "hello world",
			Template: "{{.Env.HELLO}} {{.World}}",
			Context:  map[string]any{"World": "world"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "struct with env var",
			Expected: "hello world posh",
			Template: "{{.Env.HELLO}} world {{ .Text }}",
			Context:  struct{ Text string }{Text: "posh"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{Case: "no env var", Expected: "hello world", Template: "{{.Text}} world", Context: struct{ Text string }{Text: "hello"}},
		{Case: "map", Expected: "hello world", Template: "{{.Text}} world", Context: map[string]any{"Text": "hello"}},
		{Case: "empty map", Expected: " world", Template: "{{.Text}} world", Context: map[string]string{}, ShouldError: true},
		{
			Case:     "Struct with duplicate property",
			Expected: "posh",
			Template: "{{ .OS }}",
			Context:  struct{ OS string }{OS: "posh"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "Struct with duplicate property, but global override",
			Expected: "darwin",
			Template: "{{ .$.OS }}",
			Context:  struct{ OS string }{OS: "posh"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "Map with duplicate property",
			Expected: "posh",
			Template: "{{ .OS }}",
			Context:  map[string]any{"OS": "posh"},
			Env:      map[string]string{"HELLO": "hello"},
		},
		{
			Case:     "Non-supported map",
			Expected: "darwin",
			Template: "{{ .OS }}",
			Context:  map[int]any{},
			Env:      map[string]string{"HELLO": "hello"},
		},
	}
	for _, tc := range cases {
		env := &mock.Environment{}
		env.On("Shell").Return("foo")

		for k, v := range tc.Env {
			env.On("Getenv", k).Return(v)
		}

		Cache = &cache.Template{
			OS: "darwin",
		}
		Init(env, nil, nil)

		text, err := Render(tc.Template, tc.Context)
		if tc.ShouldError {
			assert.Error(t, err)
			continue
		}

		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}

func TestPatchTemplate(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{
			Case:     "Literal dots",
			Expected: " ... ",
			Template: " ... ",
		},
		{
			Case:     "Literal dot",
			Expected: "hello . what's up",
			Template: "hello . what's up",
		},
		{
			Case:     "Variable",
			Expected: "{{range $cpu := .Data.CPU}}{{round $cpu.Mhz 2 }} {{end}}",
			Template: "{{range $cpu := .CPU}}{{round $cpu.Mhz 2 }} {{end}}",
		},
		{
			Case:     "Same prefix",
			Expected: "{{ (call .Getenv \"HELLO\") }} {{ .Data.World }} {{ .Data.WorldTrend }}",
			Template: "{{ .Env.HELLO }} {{ .World }} {{ .WorldTrend }}",
		},
		{
			Case:     "Double use of property with different child",
			Expected: "{{ (call .Getenv \"HELLO\") }} {{ .Data.World.Trend }} {{ .Data.World.Hello }} {{ .Data.World }}",
			Template: "{{ .Env.HELLO }} {{ .World.Trend }} {{ .World.Hello }} {{ .World }}",
		},
		{
			Case:     "Hello world",
			Expected: "{{(call .Getenv \"HELLO\")}} {{.Data.World}}",
			Template: "{{.Env.HELLO}} {{.World}}",
		},
		{
			Case:     "Multiple vars",
			Expected: "{{(call .Getenv \"HELLO\")}} {{.Data.World}} {{.Data.World}}",
			Template: "{{.Env.HELLO}} {{.World}} {{.World}}",
		},
		{
			Case:     "Multiple vars with spaces",
			Expected: "{{ (call .Getenv \"HELLO\") }} {{ .Data.World }} {{ .Data.World }}",
			Template: "{{ .Env.HELLO }} {{ .World }} {{ .World }}",
		},
		{
			Case:     "Braces",
			Expected: "{{ if or (.Data.Working.Changed) (.Data.Staging.Changed) }}#FF9248{{ end }}",
			Template: "{{ if or (.Working.Changed) (.Staging.Changed) }}#FF9248{{ end }}",
		},
		{
			Case:     "Global property override",
			Expected: "{{.OS}}",
			Template: "{{.$.OS}}",
		},
		{
			Case:     "Local property override",
			Expected: "{{.Data.OS}}",
			Template: "{{.OS}}",
		},
		{
			Case:     "Keep .Contains intact for Segments",
			Expected: `{{.Segments.Contains "Git"}}`,
			Template: `{{.Segments.Contains "Git"}}`,
		},
		{
			Case:     "Replace a direct call to .Segments with .Segments.List",
			Expected: `{{(.Segments.MustGet "Git").Repo}}`,
			Template: `{{.Segments.Git.Repo}}`,
		},
	}

	env := new(mock.Environment)
	env.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(env, nil, nil)

	for _, tc := range cases {
		context := map[string]any{
			"OS":         true,
			"World":      true,
			"WorldTrend": "chaos",
			"Working":    true,
			"Staging":    true,
			"CPU":        true,
		}

		tmpl := Text{
			template: tc.Template,
			context:  context,
		}

		tmpl.patchTemplate()
		assert.Equal(t, tc.Expected, tmpl.template, tc.Case)
	}
}

type Foo struct{}

func (f *Foo) Hello() string {
	return "hello"
}

func TestPatchTemplateStruct(t *testing.T) {
	env := new(mock.Environment)
	env.On("Shell").Return("foo")
	Cache = new(cache.Template)
	Init(env, nil, nil)

	tmpl := Text{
		template: "{{ .Hello }}",
		context:  Foo{},
	}

	tmpl.patchTemplate()
	assert.Equal(t, "{{ .Data.Hello }}", tmpl.template)
}

func TestSegmentContains(t *testing.T) {
	cases := []struct {
		Case     string
		Expected string
		Template string
	}{
		{Case: "match", Expected: "hello", Template: `{{ if .Segments.Contains "Git" }}hello{{ end }}`},
		{Case: "match", Expected: "world", Template: `{{ if .Segments.Contains "Path" }}hello{{ else }}world{{ end }}`},
	}

	env := &mock.Environment{}
	segments := maps.NewConcurrent[any]()
	segments.Set("Git", "foo")
	env.On("Shell").Return("foo")

	Cache = &cache.Template{
		Segments: segments,
	}
	Init(env, nil, nil)

	for _, tc := range cases {
		text, _ := Render(tc.Template, nil)
		assert.Equal(t, tc.Expected, text, tc.Case)
	}
}
