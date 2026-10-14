package layout

import (
    "testing"

    "github.com/starframe/abyss/internal/css"
    "github.com/starframe/abyss/internal/dom"
    "github.com/stretchr/testify/assert"
)

func TestFlexAvailableSpace_Row(t *testing.T) {
    contentWidth := 800.0
    totalGap := 400.0

    result := flexAvailableSpace(css.FlexDirectionRow, contentWidth, totalGap)

    assert.Equal(t, float64(400), result, "Row direction should subtract total gap")
}

func TestFlexAvailableSpace_Column(t *testing.T) {
    contentWidth := 800.0
    totalGap := 400.0

    result := flexAvailableSpace(css.FlexDirectionColumn, contentWidth, totalGap)

    assert.InDelta(t, 100000, result, 0.0, "Column direction should return large number")
}

func TestFlexTotals(t *testing.T) {
    box := &LayoutBox{
        Children: []*LayoutBox{
            {
                Style: css.ComputedStyle{
                    Width:      css.Px(100),
                    FlexGrow:    2,
                    FlexShrink:  1,
                },
            },
            {
                Style: css.ComputedStyle{
                    Width: css.Px(150),
                },
            },
        },
    }

    basis, grow, shrink := flexTotals(box, 800.0)

    assert.Equal(t, float64(250), basis, "Total basis should be sum of widths")
    assert.Equal(t, float64(2), grow, "Total grow should be sum of flex grow values")
    assert.Equal(t, float64(1), shrink, "Total shrink should be sum of flex shrink values")
}

func TestFlexRemaining_Row(t *testing.T) {
    availableSpace := 500.0
    totalBasis := 300.0

    result := flexRemaining(css.FlexDirectionRow, availableSpace, totalBasis)

    assert.Equal(t, float64(200), result, "Remaining should be available minus basis")
}

func TestFlexRemaining_Column(t *testing.T) {
    availableSpace := 500.0
    totalBasis := 300.0

    result := flexRemaining(css.FlexDirectionColumn, availableSpace, totalBasis)

    assert.Equal(t, float64(0), result, "Column direction should return 0 remaining")
}

func TestFlexContainerSize_Row(t *testing.T) {
    children := []*LayoutBox{
        {
            Style: css.ComputedStyle{
                Width: css.Px(100),
                Height: css.Px(50),
            },
            Margin: Rect{
                Height: 60,
            },
        },
        {
            Style: css.ComputedStyle{
                Width: css.Px(150),
                Height: css.Px(80),
            },
            Margin: Rect{
                Height: 90,
            },
        },
    }

    box := &LayoutBox{
        Children: children,
    }

    result := flexContainerSize(box, css.FlexDirectionRow, 200.0, 10.0)

    assert.Equal(t, float64(90), result, "Container size should be max child height")
}

func TestFlexContainerSize_Column(t *testing.T) {
    children := []*LayoutBox{
        {
            Style: css.ComputedStyle{
                Width: css.Px(100),
                Height: css.Px(50),
            },
            Margin: Rect{
                Height: 60,
            },
        },
        {
            Style: css.ComputedStyle{
                Width: css.Px(150),
                Height: css.Px(80),
            },
            Margin: Rect{
                Height: 90,
            },
        },
    }

    box := &LayoutBox{
        Children: children,
    }

    result := flexContainerSize(box, css.FlexDirectionColumn, 200.0, 10.0)

    assert.Equal(t, float64(190), result, "Column size should be pos minus gap")
}

func TestFlexWithSingleChild(t *testing.T) {
    html := `<div style="display: flex; width: 300px;">
        <div style="width: 100px;"></div>
        <div style="width: 100px;"></div>
    </div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 300.0)

    assert.NotNil(t, box)
    assert.Equal(t, 300.0, box.Content.Width)
}

func TestFlexWithMultipleChildren(t *testing.T) {
    html := `<div style="display: flex; width: 400px;">
        <div style="width: 100px;"></div>
        <div style="width: 100px;"></div>
        <div style="width: 100px;"></div>
    </div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 400.0)

    assert.NotNil(t, box)
    assert.Equal(t, 400.0, box.Content.Width)
}

func TestFlexEmptyBox(t *testing.T) {
    html := `<div style="display: flex; width: 200px;"></div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 200.0)

    assert.NotNil(t, box)
    assert.Equal(t, 200.0, box.Content.Width)
}

func TestFlexWithFlexGrow(t *testing.T) {
    html := `<div style="display: flex; width: 300px;">
        <div style="width: 100px; flex-grow: 2;"></div>
        <div style="width: 100px;"></div>
    </div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 300.0)

    assert.NotNil(t, box)
    assert.Equal(t, 300.0, box.Content.Width)
}

func TestFlexColumnDirection(t *testing.T) {
    html := `<div style="display: flex; flex-direction: column; width: 300px;">
        <div style="height: 100px;"></div>
        <div style="height: 100px;"></div>
    </div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 300.0)

    assert.NotNil(t, box)
    assert.Equal(t, 300.0, box.Content.Width)
}

func TestFlexWithGap(t *testing.T) {
    html := `<div style="display: flex; gap: 10px; width: 300px;">
        <div style="width: 100px;"></div>
        <div style="width: 100px;"></div>
    </div>`
    doc := dom.Parse(html)
    styles := css.Compute(doc)
    box := Build(doc, styles, 300.0)

    assert.NotNil(t, box)
    assert.Equal(t, 300.0, box.Content.Width)
}
