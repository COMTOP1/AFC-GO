package sanitize_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/COMTOP1/AFC-GO/server/internal/sanitize"
)

func TestHTML(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"script removed":       {`<p>hi</p><script>alert(1)</script>`, `<p>hi</p>`},
		"heading class kept":   {`<h2 class="title">T</h2>`, `<h2 class="title">T</h2>`},
		"mailto kept":          {`<a href="mailto:a@b.c">x</a>`, `<a href="mailto:a@b.c">x</a>`},
		"javascript href gone": {`<a href="javascript:alert(1)">x</a>`, `x`},
		"alignment style kept": {`<div style="text-align: center">c</div>`, `<div style="text-align: center">c</div>`},
		"img not allowed":      {`<img src="x" onerror="alert(1)">`, ``},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, sanitize.HTML(tc.in))
		})
	}
}
