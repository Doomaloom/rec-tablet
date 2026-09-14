package assessment

import (
	"fmt"
	"image/color"
	"strings"

	"rectablet/internal/shared/ui"
	"rectablet/internal/shell"
	"rectablet/internal/workflows/swimming/curriculum"
	"rectablet/internal/workflows/swimming/domain"
	"rectablet/internal/workflows/swimming/navigation"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

const multiSubskillsPerPage = 5

const (
	multiMatrixLeft   float32 = 100
	multiColumnWidth  float32 = 48
	multiHeaderTop    float32 = 84
	multiHeaderLength         = 350
	multiHeaderFont           = 15
	multiHeaderGap    float32 = 4
	multiRowGap       float32 = 3
)

var multiRowColors = []color.Color{
	color.NRGBA{R: 190, G: 224, B: 250, A: 255},
	color.NRGBA{R: 185, G: 237, B: 193, A: 255},
	color.NRGBA{R: 255, G: 235, B: 157, A: 255},
	color.NRGBA{R: 255, G: 194, B: 194, A: 255},
}

type multiStudentSnapshot struct {
	AttemptID string
	PersonID  string
	Name      string
	Answers   map[string]string
}

type multiAssessmentSnapshot struct {
	Skill    curriculum.Skill
	Students []multiStudentSnapshot
}

func (c *Controller) skillModeSwitch(multi bool) []shell.HeaderSwitchOption {
	return []shell.HeaderSwitchOption{
		{Label: "Single", Selected: !multi, Run: c.EnterSingleMode},
		{Label: "Multi", Selected: multi, Run: c.EnterMultiMode},
	}
}

func (c *Controller) EnterMultiMode() error {
	if c.Frame().Screen != navigation.SkillSwimmers {
		return nil
	}
	return c.Run(func() error {
		c.multiMode = true
		c.multiSelection = nil
		c.Push(navigation.MultiSelection, nil)
		return nil
	})
}

func (c *Controller) EnterSingleMode() error {
	if c.Frame().Screen != navigation.MultiSelection {
		return nil
	}
	return c.Run(func() error {
		c.multiMode = false
		c.multiSelection = nil
		c.Pop()
		return nil
	})
}

func (c *Controller) ConfirmMultiSelection() error {
	if c.Frame().Screen != navigation.MultiSelection || len(c.multiSelection) == 0 {
		return nil
	}
	return c.Run(func() error {
		f := c.Frame()
		attempts, e := c.Store.OpenMultiAttempts(c.multiSelection, f.Lesson, f.Skill)
		if e != nil {
			return nil
		}
		c.multiAttempts = attempts
		c.multiSubskillPage = 0
		c.Push(navigation.MultiAssessment, nil)
		return nil
	})
}

func (c *Controller) DiscardMultiSelection() error {
	if c.Frame().Screen != navigation.MultiSelection {
		return nil
	}
	return c.Run(func() error {
		c.multiMode = false
		c.multiSelection = nil
		c.Pop()
		return nil
	})
}

func (c *Controller) buildMultiAssessmentState(course domain.Course, frame navigation.Frame) (navigation.State, error) {
	state := c.InitialState()
	state.Title = "Multi Mark"
	state.Subtitle = ""
	state.Context = course.Name + " · " + frame.Lesson
	state.HomeRailAction = saveRailAction(c.SaveMultiAssessment)
	state.BackRailAction = discardRailAction(c.DiscardMultiAssessment)
	state.HideFooter = true
	state.Empty = "No students selected."

	skill, e := c.Store.Skill(frame.Version, frame.Skill)
	if e != nil {
		return state, e
	}
	state.Title = skill.Number + ". " + skill.Name
	state.Context += "\n" + skill.Name
	state.Progress = fmt.Sprintf("%d students · %d subskills", len(c.multiAttempts), len(skill.Subskills))
	state.PageSize = multiStudentPageSize(len(c.multiAttempts), len(skill.Subskills))

	snapshot := multiAssessmentSnapshot{Skill: skill, Students: make([]multiStudentSnapshot, 0, len(c.multiAttempts))}
	for _, attempt := range c.multiAttempts {
		answers, e := c.Store.Draft(attempt.AttemptID)
		if e != nil {
			// A missing or unreadable local draft should not strand the
			// instructor in an error screen. Keep the matrix usable and let the
			// eventual batch save apply the normal silent-failure behavior.
			answers = map[string]string{}
		}
		snapshot.Students = append(snapshot.Students, multiStudentSnapshot{AttemptID: attempt.AttemptID, PersonID: attempt.PersonID, Name: attempt.Name, Answers: answers})
		state.Rows = append(state.Rows, navigation.Row{ID: attempt.PersonID, Label: attempt.Name})
	}
	state.ComposeBody = c.composeMultiAssessmentBody(snapshot)
	return state, nil
}

func (c *Controller) ToggleMultiCell(personID, subskillID string) error {
	if c.Frame().Screen != navigation.MultiAssessment {
		return nil
	}
	return c.Run(func() error {
		for _, attempt := range c.multiAttempts {
			if attempt.PersonID != personID {
				continue
			}
			answers, e := c.Store.Draft(attempt.AttemptID)
			if e != nil {
				return nil
			}
			status := "successful"
			if answers[subskillID] == "successful" {
				status = "unmarked"
			}
			if e = c.Store.Mark(attempt.AttemptID, subskillID, status); e != nil {
				return nil
			}
			c.Saved("")
			return nil
		}
		return nil
	})
}

func (c *Controller) SaveMultiAssessment() error {
	if c.Frame().Screen != navigation.MultiAssessment || len(c.multiAttempts) == 0 {
		return nil
	}
	return c.Run(func() error {
		ids := make([]string, 0, len(c.multiAttempts))
		for _, attempt := range c.multiAttempts {
			ids = append(ids, attempt.AttemptID)
		}
		if e := c.Store.SubmitAttempts(ids); e != nil {
			return nil
		}
		c.multiAttempts = nil
		c.multiSelection = nil
		c.multiMode = false
		c.Pop()
		c.Pop()
		return nil
	})
}

func (c *Controller) DiscardMultiAssessment() error {
	if c.Frame().Screen != navigation.MultiAssessment || len(c.multiAttempts) == 0 {
		return nil
	}
	return c.Run(func() error {
		if e := c.Store.DiscardMultiAttempts(c.multiAttempts); e != nil {
			return nil
		}
		c.multiAttempts = nil
		c.multiSelection = nil
		c.multiMode = false
		c.Pop()
		c.Pop()
		return nil
	})
}

func (c *Controller) composeMultiSelectionBody(state shell.State) ui.BodyComposer {
	snapshot := state
	return func(surface *ui.Surface) {
		if snapshot.RowCount == 0 {
			ui.Empty(surface, snapshot.Empty)
			return
		}
		ui.List(surface, snapshot.Rows, func(id string) { _ = c.Controller.Select(id) })
	}
}

func multiHeaderText(_ int, text string) string {
	return strings.Join(strings.Fields(text), " ")
}

func multiStudentPageSize(students, subskills int) int {
	limit := 4
	if subskills > multiSubskillsPerPage {
		limit = 3
	}
	return max(1, min(students, limit))
}

func (c *Controller) pageMultiSubskills(direction, count int) error {
	if c.Frame().Screen != navigation.MultiAssessment {
		return nil
	}
	return c.Run(func() error {
		lastPage := (count - 1) / multiSubskillsPerPage
		c.multiSubskillPage = max(0, min(c.multiSubskillPage+direction, lastPage))
		return nil
	})
}

type multiHeaderPlacement struct {
	Header *ui.AngledLabel
	Width  float32
	Height float32
	Offset float32
}

func multiHeaderPlacements(headers []*ui.AngledLabel, gap float32) ([]multiHeaderPlacement, float32) {
	placements := make([]multiHeaderPlacement, 0, len(headers))
	totalHeight := float32(0)
	for index, header := range headers {
		size := header.MinSize()
		placements = append(placements, multiHeaderPlacement{
			Header: header,
			Width:  size.Width,
			Height: size.Height,
			Offset: totalHeight,
		})
		totalHeight += size.Height
		if index < len(headers)-1 {
			totalHeight += gap
		}
	}
	return placements, totalHeight
}

func multiStudentPage(students []multiStudentSnapshot, page, pageSize int) ([]multiStudentSnapshot, int) {
	start := page * pageSize
	if start >= len(students) {
		start = 0
	}
	end := min(len(students), start+pageSize)
	return students[start:end], start
}

func multiSubskillPage(subskills []curriculum.Subskill, page int) ([]curriculum.Subskill, int, int) {
	if len(subskills) == 0 {
		return nil, 0, 0
	}
	start := min(page*multiSubskillsPerPage, (len(subskills)-1)/multiSubskillsPerPage*multiSubskillsPerPage)
	end := min(len(subskills), start+multiSubskillsPerPage)
	return subskills[start:end], start, end
}

func multiMatrixWidth(width float32) float32 { return width - 200 }

func addMultiHeaders(surface *ui.Surface, headers []multiHeaderPlacement, rowTop float32, matrixLeft, columnWidth float32, nameWidth func(float32) float32) {
	for index, placement := range headers {
		index, placement := index, placement
		if placement.Offset > 0 {
			connector := canvas.NewRectangle(ui.Muted)
			surface.AddNamed(fmt.Sprintf("multi-header-connector-%d", index), connector, func(w, h float32) (float32, float32, float32, float32) {
				columnRight := matrixLeft + nameWidth(w) + float32(index+1)*columnWidth
				return columnRight - 2, rowTop - placement.Offset, 1.5, placement.Offset
			})
		}
		surface.AddNamed(fmt.Sprintf("multi-header-%d", index), placement.Header, func(w, h float32) (float32, float32, float32, float32) {
			columnRight := matrixLeft + nameWidth(w) + float32(index+1)*columnWidth
			return columnRight - placement.Width, rowTop - placement.Offset - placement.Height, placement.Width, placement.Height
		})
	}
}

func newMultiHeaders(subskills []curriculum.Subskill, start int, showDetails func()) []*ui.AngledLabel {
	headers := make([]*ui.AngledLabel, 0, len(subskills))
	for index, subskill := range subskills {
		header := ui.NewAngledLabel(multiHeaderText(start+index, subskill.Text), multiHeaderLength, multiHeaderFont, 0, ui.White, ui.Navy, ui.Navy)
		header.OnTap = showDetails
		headers = append(headers, header)
	}
	return headers
}

func multiMaxHeaderStackHeight(subskills []curriculum.Subskill) float32 {
	maximum := float32(0)
	for start := 0; start < len(subskills); start += multiSubskillsPerPage {
		end := min(len(subskills), start+multiSubskillsPerPage)
		pageHeaders := newMultiHeaders(subskills[start:end], start, nil)
		_, height := multiHeaderPlacements(pageHeaders, multiHeaderGap)
		maximum = max(maximum, height)
	}
	return maximum
}

func addMultiStudentRows(surface *ui.Surface, students []multiStudentSnapshot, studentStart int, subskills []curriculum.Subskill, rowTop, rowHeight, rowGap, matrixLeft, columnWidth float32, matrixWidth, nameWidth func(float32) float32, toggleCell func(string, string)) {
	for rowIndex, student := range students {
		student, rowIndex := student, rowIndex
		rowY := rowTop + float32(rowIndex)*(rowHeight+rowGap)
		rowColor := multiRowColors[(studentStart+rowIndex)%len(multiRowColors)]
		background := canvas.NewRectangle(rowColor)
		background.CornerRadius = 12
		surface.AddNamed("multi-row-bg:"+student.PersonID, background, func(w, h float32) (float32, float32, float32, float32) {
			return matrixLeft, rowY, matrixWidth(w), rowHeight
		})

		name := ui.NewText(student.Name, 18, ui.Navy)
		name.MinimumFont = 15
		name.Bold = true
		name.Center = false
		name.MaxLines = 2
		surface.AddNamed("multi-row-name:"+student.PersonID, name, func(w, h float32) (float32, float32, float32, float32) {
			return matrixLeft + 12, rowY, nameWidth(w) - 24, rowHeight
		})

		for columnIndex, subskill := range subskills {
			columnIndex, subskill := columnIndex, subskill
			complete := student.Answers[subskill.ID] == "successful"
			fill, ink := rowColor, ui.Navy
			if complete {
				fill, ink = ui.Green, ui.White
			}
			cell := surface.Button(fmt.Sprintf("multi-cell:%s:%s", student.PersonID, subskill.ID), "", func() { toggleCell(student.PersonID, subskill.ID) })
			cell.Fill, cell.Ink, cell.FontSize = fill, ink, 22
			cell.Selected = complete
			cell.Outlined = true
			if complete {
				cell.Icon = saveActionIcon
				cell.IconOnly = true
			}
			surface.AddNamed(fmt.Sprintf("multi-cell:%s:%s", student.PersonID, subskill.ID), cell, func(w, h float32) (float32, float32, float32, float32) {
				return matrixLeft + nameWidth(w) + float32(columnIndex)*columnWidth + 1, rowY + 1, columnWidth - 2, rowHeight - 2
			})
		}
	}
}

func addMultiSubskillPager(surface *ui.Surface, controller *Controller, subskillCount, subskillStart, subskillEnd int) {
	if subskillCount <= multiSubskillsPerPage {
		return
	}
	hint := ui.NewText("Tap a subskill for full text", 14, ui.Muted)
	surface.AddNamed("multi-subskill-hint", hint, func(w, h float32) (float32, float32, float32, float32) {
		return multiMatrixLeft + 12, 422, w - 512, 44
	})

	previous := surface.Button("multi-subskills-previous", "‹", func() { _ = controller.pageMultiSubskills(-1, subskillCount) })
	previous.Fill, previous.Ink, previous.FontSize = ui.Rail, ui.Blue, 30
	previous.Disabled = subskillStart == 0
	previous.Icon, previous.IconOnly = theme.NewPrimaryThemedResource(theme.NavigateBackIcon()), true
	if previous.Disabled {
		previous.Icon = theme.NewDisabledResource(theme.NavigateBackIcon())
	}
	surface.AddNamed("multi-subskills-previous", previous, func(w, h float32) (float32, float32, float32, float32) {
		return w - 388, 422, 56, 44
	})

	progress := ui.NewText(fmt.Sprintf("Subskills %d–%d of %d", subskillStart+1, subskillEnd, subskillCount), 16, ui.Navy)
	progress.Center = true
	surface.AddNamed("multi-subskills-progress", progress, func(w, h float32) (float32, float32, float32, float32) {
		return w - 328, 422, 168, 44
	})

	next := surface.Button("multi-subskills-next", "›", func() { _ = controller.pageMultiSubskills(1, subskillCount) })
	next.Fill, next.Ink, next.FontSize = ui.Rail, ui.Blue, 30
	next.Disabled = subskillEnd == subskillCount
	next.Icon, next.IconOnly = theme.NewPrimaryThemedResource(theme.NavigateNextIcon()), true
	if next.Disabled {
		next.Icon = theme.NewDisabledResource(theme.NavigateNextIcon())
	}
	surface.AddNamed("multi-subskills-next", next, func(w, h float32) (float32, float32, float32, float32) {
		return w - 156, 422, 56, 44
	})
}

func showMultiSubskills(window fyne.Window, texts []string) {
	label := widget.NewLabel(strings.Join(texts, "\n\n"))
	label.Wrapping = fyne.TextWrapWord
	details := dialog.NewCustom("Subskills", "Close", container.NewVScroll(label), window)
	details.Resize(fyne.NewSize(620, 360))
	details.Show()
}

func (c *Controller) composeMultiAssessmentBody(snapshot multiAssessmentSnapshot) ui.BodyComposer {
	return func(surface *ui.Surface) {
		if len(snapshot.Students) == 0 || len(snapshot.Skill.Subskills) == 0 {
			ui.Empty(surface, "No students selected.")
			return
		}

		subskillCount := len(snapshot.Skill.Subskills)
		pageSize := multiStudentPageSize(len(snapshot.Students), subskillCount)
		students, studentStart := multiStudentPage(snapshot.Students, c.Frame().Page, pageSize)
		subskills, subskillStart, subskillEnd := multiSubskillPage(snapshot.Skill.Subskills, c.multiSubskillPage)
		// Reserve the full page width even on the final subskill page so the
		// student-name column and existing marks do not move when paging.
		columnCount := min(subskillCount, multiSubskillsPerPage)

		matrixWidth := multiMatrixWidth
		nameWidth := func(width float32) float32 {
			return matrixWidth(width) - float32(columnCount)*multiColumnWidth
		}
		allSubskillTexts := make([]string, 0, subskillCount)
		for index, subskill := range snapshot.Skill.Subskills {
			allSubskillTexts = append(allSubskillTexts, multiHeaderText(index, subskill.Text))
		}
		showDetails := func() { showMultiSubskills(surface.Window, allSubskillTexts) }
		headers := newMultiHeaders(subskills, subskillStart, showDetails)
		placements, _ := multiHeaderPlacements(headers, multiHeaderGap)
		headerStackHeight := multiMaxHeaderStackHeight(snapshot.Skill.Subskills)
		rowTop := multiHeaderTop + headerStackHeight
		rowBottom := float32(466)
		if subskillCount > multiSubskillsPerPage {
			rowBottom = 414
		}
		rowHeight := min(float32(56), (rowBottom-rowTop-float32(pageSize-1)*multiRowGap)/float32(pageSize))

		addMultiHeaders(surface, placements, rowTop, multiMatrixLeft, multiColumnWidth, nameWidth)
		addMultiStudentRows(surface, students, studentStart, subskills, rowTop, rowHeight, multiRowGap, multiMatrixLeft, multiColumnWidth, matrixWidth, nameWidth, func(personID, subskillID string) {
			_ = c.ToggleMultiCell(personID, subskillID)
		})
		addMultiSubskillPager(surface, c, subskillCount, subskillStart, subskillEnd)
	}
}
