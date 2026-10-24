package geometry

import (
	"strings"

	"rectablet/next/contracts"
	"rectablet/src/shared/ui"
)

// Layout slots use the selected profile's logical surface, scaled uniformly by the root.
type LayoutSlot = contracts.LayoutSlot

const SplitHeaderHeight = contracts.SplitHeaderHeight

const (
	PagePositionTrackWidth  float32 = 12
	PagePositionGutter      float32 = 8
	PagePositionBodyReserve float32 = PagePositionTrackWidth + 2*PagePositionGutter - ui.Gap
)

func SplitHeaderSlot(width float32, index int) (float32, float32, float32, float32) {
	halfWidth := (width - 2*ui.ContentInset - ui.Gap) / 2
	return ui.ContentInset + float32(index)*(halfWidth+ui.Gap), ui.OuterMargin, halfWidth, SplitHeaderHeight
}

func SplitHeaderActionSlot(width float32, actionWidth float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - actionWidth, ui.OuterMargin, actionWidth, SplitHeaderHeight
}

func SplitHeaderHalfWithActionSlot(width float32, index int, actionWidth float32) (float32, float32, float32, float32) {
	available := width - 2*ui.ContentInset - 2*ui.Gap - actionWidth
	halfWidth := available / 2
	return ui.ContentInset + float32(index)*(halfWidth+ui.Gap), ui.OuterMargin, halfWidth, SplitHeaderHeight
}

func BackgroundSlot(width, height float32) (float32, float32, float32, float32) {
	return 0, 0, width, height
}

func RailSlot(height float32, index, count int) (float32, float32, float32, float32) {
	slotHeight := (height - 2*ui.OuterMargin - float32(count-1)*ui.Gap) / float32(count)
	return ui.OuterMargin, ui.OuterMargin + float32(index)*(slotHeight+ui.Gap), ui.RailWidth, slotHeight
}

func HomeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return RailSlot(height, 0, 1)
}
func SplitHomeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return RailSlot(height, 1, 2)
}
func SplitBackRailSlot(_, height float32) (float32, float32, float32, float32) {
	return RailSlot(height, 0, 2)
}
func PreviousPageSlot(width, height float32) (float32, float32, float32, float32) {
	_, y, w, h := RailSlot(height, 0, 2)
	return width - ui.OuterMargin - ui.RailWidth, y, w, h
}
func NextPageSlot(width, height float32) (float32, float32, float32, float32) {
	_, y, w, h := RailSlot(height, 1, 2)
	return width - ui.OuterMargin - ui.RailWidth, y, w, h
}

func PagePositionTrackSlot(state contracts.State) LayoutSlot {
	return func(width, height float32) (float32, float32, float32, float32) {
		_, y, _, bodyHeight := BodySlot(state)(width, height)
		x := width - ui.OuterMargin - ui.RailWidth - PagePositionGutter - PagePositionTrackWidth
		return x, y, PagePositionTrackWidth, bodyHeight
	}
}

func PagePositionThumbSlot(state contracts.State) LayoutSlot {
	return func(width, height float32) (float32, float32, float32, float32) {
		x, y, trackWidth, trackHeight := PagePositionTrackSlot(state)(width, height)
		thumbHeight := trackHeight / float32(state.PageCount)
		page := min(max(0, state.Page), state.PageCount-1)
		progress := float32(page) / float32(max(1, state.PageCount-1))
		return x, y + progress*(trackHeight-thumbHeight), trackWidth, thumbHeight
	}
}

func HeaderActionSlot(width, _ float32, actionWidth float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - actionWidth, ui.OuterMargin, actionWidth, ui.HeaderHeight
}
func HeaderTitleSlot(width, _ float32, left, actionWidth float32) (float32, float32, float32, float32) {
	gap := float32(0)
	if actionWidth > 0 {
		gap = ui.Gap
	}
	return left, ui.OuterMargin, width - ui.ContentInset - left - actionWidth - gap, ui.HeaderHeight
}

func bodyWidth(state contracts.State, width float32) float32 {
	available := width - contentLeft(state) - ui.ContentInset
	if state.PageCount > 1 && state.RightRailActions == nil {
		available -= PagePositionBodyReserve
	}
	return available
}

func contentLeft(state contracts.State) float32 {
	if state.HideLeftRail {
		return ui.OuterMargin
	}
	return ui.ContentInset
}

func HeaderLeft(state contracts.State, width float32) float32 {
	if state.HeaderStartFraction <= 0 {
		return contentLeft(state)
	}
	// Give reclaimed left-rail space to the leading column, preserving the header's start.
	reclaimed := ui.ContentInset - contentLeft(state)
	return ui.ContentInset + (bodyWidth(state, width)-reclaimed-ui.Gap)*state.HeaderStartFraction + ui.Gap
}

func HeaderHeight(state contracts.State) float32 {
	if state.SplitHeader != nil {
		return SplitHeaderHeight
	}
	if state.HeaderTitleFontSize > 25 || state.FitHeader {
		return 56
	}
	return ui.HeaderHeight
}

func PrimaryInfoText(state contracts.State) string {
	parts := []string{state.PrimaryInfo}
	if state.Progress != "" && !state.ProgressInBody {
		parts = append(parts, state.Progress)
	}
	joined := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			joined = append(joined, part)
		}
	}
	return strings.Join(joined, " · ")
}

func PrimaryInfoHeight(state contracts.State, width float32) float32 {
	text := PrimaryInfoText(state)
	if text == "" {
		return 0
	}
	return max(float32(30), ui.WrappedTextHeight(text, width-ui.ContentInset-HeaderLeft(state, width), PrimaryInfoFontSize(state), true)+4)
}

func PrimaryInfoFontSize(state contracts.State) float32 {
	if state.PrimaryInfoFontSize > 0 {
		return state.PrimaryInfoFontSize
	}
	return 20
}

func FooterHeight(state contracts.State) float32 {
	if state.HideFooter {
		return 0
	}
	if state.FooterAction != "" {
		return ui.HeaderHeight + ui.OuterMargin + ui.Gap
	}
	return 0
}
func BodySlot(state contracts.State) ui.Bounds {
	return func(width, height float32) (float32, float32, float32, float32) {
		top := ui.OuterMargin
		if !state.HideHeader {
			top += HeaderHeight(state) + ui.Gap
			infoHeight := PrimaryInfoHeight(state, width)
			if infoHeight > 0 {
				top += infoHeight + ui.Gap
			}
		}
		return contentLeft(state), top, bodyWidth(state, width), height - top - ui.OuterMargin - FooterHeight(state)
	}
}
func SubmitFooterSlot(width, height float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - 208, height - ui.OuterMargin - ui.HeaderHeight, 208, ui.HeaderHeight
}
