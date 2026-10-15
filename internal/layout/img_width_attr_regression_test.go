package layout

import (
	"testing"

	"github.com/starframe-dev/abyss/internal/cellmetrics"
	"github.com/starframe-dev/abyss/internal/css"
	"github.com/starframe-dev/abyss/internal/dom"
	"github.com/stretchr/testify/require"
)

// An <img> without CSS width keeps its width attribute (CSS px) instead of
// stretching over the whole container, so the logo keeps its aspect ratio.
func TestImgWidthAttributeIsUsedWithoutCSSWidth(t *testing.T) {
	doc := dom.Parse(`<body><img alt="Google" width="272" height="92" src="logo.png"></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	img := findTag(box, "img")
	require.NotNil(t, img)
	require.Equal(t, 272.0, img.Content.Width)
}

// The height attribute gives an image its real height, so a 92 px logo gets
// more than the 48 px placeholder minimum and keeps its rows.
func TestImgHeightAttributeIsUsedWithoutCSSHeight(t *testing.T) {
	doc := dom.Parse(`<body><img alt="Google" width="272" height="92" src="logo.png"></body>`)
	box := Build(doc, css.Compute(doc), 200*cellmetrics.WidthPx)
	img := findTag(box, "img")
	require.NotNil(t, img)
	require.Equal(t, 92.0, img.Content.Height)
}
