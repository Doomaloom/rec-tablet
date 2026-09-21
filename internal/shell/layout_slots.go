package shell

import "rectablet/internal/shared/ui"

// Layout slots use the logical 800x480 surface, scaled uniformly by the root.
type LayoutSlot func(width, height float32) (x, y, slotWidth, slotHeight float32)

const SplitHeaderHeight float32 = 32

func splitHeaderSlot(width float32, index int) (float32, float32, float32, float32) {
	halfWidth := (width - 2*ui.ContentInset - ui.Gap) / 2
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
	if state.HeaderTitleFontSize > 25 {
		return 56
	}
	return ui.HeaderHeight
}
func hasContext(state State) bool { return !state.HideContext && state.Subtitle != "" }
func footerHeight(state State) float32 {
	if state.HideFooter {
		return 0
	}
	height := ui.StatusHeight + ui.Gap
	if state.FooterAction != "" {
		height += ui.HeaderHeight + ui.Gap
	}
	return height
}
func bodySlot(state State) ui.Bounds {
	return func(width, height float32) (float32, float32, float32, float32) {
		top := ui.OuterMargin + headerHeight(state) + ui.Gap
		if hasContext(state) {
			top += ui.ContextHeight + ui.Gap
		}
		return ui.ContentInset, top, width - 2*ui.ContentInset, height - top - ui.OuterMargin - footerHeight(state)
	}
}
func submitFooterSlot(width, height float32) (float32, float32, float32, float32) {
	return width - ui.ContentInset - 208, height - ui.OuterMargin - ui.HeaderHeight, 208, ui.HeaderHeight
}
