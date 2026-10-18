package shell

import (
	"strings"

	"rectablet/src/shared/ui"
)

// Layout slots use the selected profile's logical surface, scaled uniformly by the root.
type LayoutSlot func(width, height float32) (x, y, slotWidth, slotHeight float32)

const SplitHeaderHeight float32 = 52

const (
	pagePositionTrackWidth  float32 = 12
	pagePositionGutter      float32 = 8
	pagePositionBodyReserve float32 = pagePositionTrackWidth + 2*pagePositionGutter - ui.Gap
)

func splitHeaderSlot(width float32, index int) (float32, float32, float32, float32) {
	halfWidth := (width - 2*ui.ContentInset - ui.Gap) / 2
	return ui.ContentInset + float32(index)*(halfWidth+ui.Gap), ui.OuterMargin, halfWidth, SplitHeaderHeight
}

func splitHeaderActionSlot(width float32, actionWidth float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - actionWidth, ui.OuterMargin, actionWidth, SplitHeaderHeight
}

func splitHeaderHalfWithActionSlot(width float32, index int, actionWidth float32) (float32, float32, float32, float32) {
	available := width - 2*ui.ContentInset - 2*ui.Gap - actionWidth
	halfWidth := available / 2
	return ui.ContentInset + float32(index)*(halfWidth+ui.Gap), ui.OuterMargin, halfWidth, SplitHeaderHeight
}

func backgroundSlot(width, height float32) (float32, float32, float32, float32) {
	return 0, 0, width, height
}

func railSlot(height float32, index, count int) (float32, float32, float32, float32) {
	slotHeight := (height - 2*ui.OuterMargin - float32(count-1)*ui.Gap) / float32(count)
	return ui.OuterMargin, ui.OuterMargin + float32(index)*(slotHeight+ui.Gap), ui.RailWidth, slotHeight
}

func homeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return railSlot(height, 0, 1)
}
func splitHomeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return railSlot(height, 1, 2)
}
func splitBackRailSlot(_, height float32) (float32, float32, float32, float32) {
	return railSlot(height, 0, 2)
}
func previousPageSlot(width, height float32) (float32, float32, float32, float32) {
	_, y, w, h := railSlot(height, 0, 2)
	return width - ui.OuterMargin - ui.RailWidth, y, w, h
}
func nextPageSlot(width, height float32) (float32, float32, float32, float32) {
	_, y, w, h := railSlot(height, 1, 2)
	return width - ui.OuterMargin - ui.RailWidth, y, w, h
}

func pagePositionTrackSlot(state State) LayoutSlot {
	return func(width, height float32) (float32, float32, float32, float32) {
		_, y, _, bodyHeight := bodySlot(state)(width, height)
		x := width - ui.OuterMargin - ui.RailWidth - pagePositionGutter - pagePositionTrackWidth
		return x, y, pagePositionTrackWidth, bodyHeight
	}
}

func pagePositionThumbSlot(state State) LayoutSlot {
	return func(width, height float32) (float32, float32, float32, float32) {
		x, y, trackWidth, trackHeight := pagePositionTrackSlot(state)(width, height)
		thumbHeight := trackHeight / float32(state.PageCount)
		page := min(max(0, state.Page), state.PageCount-1)
		progress := float32(page) / float32(max(1, state.PageCount-1))
		return x, y + progress*(trackHeight-thumbHeight), trackWidth, thumbHeight
	}
}

func headerActionSlot(width, _ float32, actionWidth float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - actionWidth, ui.OuterMargin, actionWidth, ui.HeaderHeight
}
func headerTitleSlot(width, _ float32, left, actionWidth float32) (float32, float32, float32, float32) {
	gap := float32(0)
	if actionWidth > 0 {
		gap = ui.Gap
	}
	return left, ui.OuterMargin, width - ui.ContentInset - left - actionWidth - gap, ui.HeaderHeight
}

func headerHeight(state State) float32 {
	if state.SplitHeader != nil {
		return SplitHeaderHeight
	}
	if state.HeaderTitleFontSize > 25 || state.FitHeader {
		return 56
	}
	return ui.HeaderHeight
}

func primaryInfoText(state State) string {
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

func primaryInfoHeight(state State, width float32) float32 {
	text := primaryInfoText(state)
	if text == "" {
		return 0
	}
	return max(float32(30), ui.WrappedTextHeight(text, width-2*ui.ContentInset, primaryInfoFontSize(state), true)+4)
}

func primaryInfoFontSize(state State) float32 {
	if state.PrimaryInfoFontSize > 0 {
		return state.PrimaryInfoFontSize
	}
	return 20
}

func footerHeight(state State) float32 {
	if state.HideFooter {
		return 0
	}
	if state.FooterAction != "" {
		return ui.HeaderHeight + ui.OuterMargin + ui.Gap
	}
	return 0
}
func bodySlot(state State) ui.Bounds {
	return func(width, height float32) (float32, float32, float32, float32) {
		top := ui.OuterMargin + headerHeight(state) + ui.Gap
		infoHeight := primaryInfoHeight(state, width)
		if infoHeight > 0 {
			top += infoHeight + ui.Gap
		}
		bodyWidth := width - 2*ui.ContentInset
		if state.PageCount > 1 {
			bodyWidth -= pagePositionBodyReserve
		}
		return ui.ContentInset, top, bodyWidth, height - top - ui.OuterMargin - footerHeight(state)
	}
}
func submitFooterSlot(width, height float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - 208, height - ui.OuterMargin - ui.HeaderHeight, 208, ui.HeaderHeight
}
