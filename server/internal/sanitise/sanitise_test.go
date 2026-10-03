package sanitise_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitise"
)

func TestHTML(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"script removed":       {`<p>hi</p><script>alert(1)</script>`, `<p>hi</p>`},
		"heading class kept":   {`<h2 class="title">T</h2>`, `<h2 class="title">T</h2>`},
		"mailto kept":          {`<a href="mailto:a@b.c">x</a>`, `<a href="mailto:a@b.c">x</a>`},
		"javascript href gone": {`<a href="javascript:alert(1)">x</a>`, `x`},
		"alignment style kept": {`<div style="text-align: center">c</div>`, `<div style="text-align: center">c</div>`},
		"editor marks kept": {
			`<p><strong>b</strong><em>i</em><u>u</u><s>s</s></p>`,
			`<p><strong>b</strong><em>i</em><u>u</u><s>s</s></p>`,
		},
		"h3 alignment kept": {`<h3 style="text-align: right;">h</h3>`, `<h3 style="text-align: right;">h</h3>`},
		"img not allowed":   {`<img src="x" onerror="alert(1)">`, ``},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitise.HTML(tc.in))
		})
	}
}
