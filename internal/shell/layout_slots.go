package shell

// Layout slots describe the logical 800x480 design surface. The root layout
// scales these slots uniformly when the window is larger or smaller.
type LayoutSlot func(width, height float32) (x, y, slotWidth, slotHeight float32)

func backgroundSlot(width, height float32) (float32, float32, float32, float32) {
	return 0, 0, width, height
}

func homeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return 12, 14, 76, height - 28
}

func splitHomeRailSlot(_, height float32) (float32, float32, float32, float32) {
	return 12, 26 + (height-40)/2, 76, (height - 40) / 2
}

func splitBackRailSlot(_, height float32) (float32, float32, float32, float32) {
	return 12, 14, 76, (height - 40) / 2
}

func previousPageSlot(width, height float32) (float32, float32, float32, float32) {
	return width - 88, 14, 76, (height - 40) / 2
}

func nextPageSlot(width, height float32) (float32, float32, float32, float32) {
	return width - 88, 26 + (height-40)/2, 76, (height - 40) / 2
}

func headerActionSlot(width, _ float32, actionWidth float32) (float32, float32, float32, float32) {
	return width - 100 - actionWidth, 14, actionWidth, 58
}

func headerTitleSlot(width, _ float32, left, actionWidth float32) (float32, float32, float32, float32) {
	gap := float32(0)
	if actionWidth > 0 {
		gap = 8
	}
	return left, 14, width - 100 - left - actionWidth - gap, 58
}

func contextSlot(width, _ float32) (float32, float32, float32, float32) {
	return 100, 75, width - 200, 36
}

func progressFooterSlot(width, height float32) (float32, float32, float32, float32) {
	return 100, height - 71, width - 424, 57
}

func submitFooterSlot(width, height float32) (float32, float32, float32, float32) {
	return width - 308, height - 71, 208, 57
}

func deviceStatusSlot(width, height float32) (float32, float32, float32, float32) {
	return width - 230, height - 71, 130, 57
}
