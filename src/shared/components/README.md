# COB Aquatics components for Fyne

A native translation of `COB-Aquatics-SPA/frontend/src/general-components`,
independent of lesson data. The tablet uses this theme and its components;
`src/shared/ui` adapts button surfaces for fixed display profiles, long presses,
and lesson progress overlays.

## Run the gallery

From `rec-tablet`:

```sh
make gallery
# Or build once and launch directly:
make gallery-build
./dist/native/ui-gallery
# Narrow viewport:
make gallery ARGS='--width 360 --height 800'
```

The gallery is a separate executable and app identity. It uses local demo state,
opens no lesson database, and does not replace any tablet screen. It includes
buttons, cards, selection, tabs, an editable form, validation, notices, and a modal.

## Paired files

Each component lives in its own package. Filenames identify the component using a
snake_case prefix, as in the action button examples below:

- `action_button_component_ui.go`: Fyne renderers, visual trees, layout, colors, and spacing.
- `action_button_component.go`: public types, constructors, state changes, callbacks, and input handling.
- `action_button_component_test.go`, where present: interaction and rendering checks.

Presentation-only components still have a constructor file and a UI file.
`design` contains shared tokens, font assets, scoped theme, layout helpers, and
canvas-local Escape handling. The gallery follows the same presentation/behavior
split in `src/gallery/gallery_ui.go` and `gallery.go`.

## Source mapping and API

Every package below corresponds to the same-named React component in the web
collection's `components` directory.

| Web component | Go package | Construction / behavior |
| --- | --- | --- |
| ActionButton | `actionbutton` | `New(text, Options, callback)`; primary, secondary (default), outline, danger, ghost; sm/md/lg; `SetText`, `Disable`, `Enable` |
| Card | `card` | `New(content, variant)`; panel (default), subtle, interactive hover, plain; `SetContent` |
| SelectableCard | `selectablecard` | `New(content, Options, callback)`; `SetSelected`, `Selected`, `SetContent`, `Disable`, `Enable` |
| SegmentedTabs | `segmentedtabs` | `New([]Item, Options)`; rounded or pill; `SetActive` and `ActiveKey`; keyed change callback |
| TextInput | `textinput` | `New(Options)`; default/accent fill; native entry APIs; add `Object()` to layouts |
| Textarea | `textarea` | `New(Options)`; native multiline entry APIs; add `Object()` to layouts |
| Select | `select` (package `selectcontrol`) | `New(Options)`; native selection APIs and popup; add `Object()` to layouts |
| Field | `field` | `New(control, Options)`; `SetError`, `SetHelper`; errors replace helper text |
| Notice | `notice` | `New(text, tone)` or `NewContent(content, tone)`; info/success/warning/danger; `SetText`, `SetTone` |
| EmptyState | `emptystate` | `New(Options)` with optional title, description, and content |
| PageShell | `pageshell` | `New(Options, content...)`; centered, maximum width (default 1152), 24-unit section gaps |
| ModalShell | `modalshell` | `New(window, content, Options)`; `Show`, `Hide`, `Visible`, and `OnClose` |

`SetActive` is silent and ignores unknown or disabled keys. User tab changes call
`OnChanged` once. Selectable cards leave exclusive/multiple-selection policy to
the caller. Native entry/select setters retain Fyne's normal callback semantics.
An empty tab key leaves the group unselected until a valid choice is made.

## Example

```go
import (
    "fyne.io/fyne/v2/container"
    "rectablet/src/shared/components/actionbutton"
    "rectablet/src/shared/components/card"
    "rectablet/src/shared/components/field"
    "rectablet/src/shared/components/textinput"
)

name := textinput.New(textinput.Options{Placeholder: "First name"})
nameField := field.New(name.Object(), field.Options{Label: "First name"})
save := actionbutton.New("Save", actionbutton.Options{
    Variant: actionbutton.Primary,
}, func() {
    if name.Text == "" {
        nameField.SetError("Enter a first name.")
        return
    }
    nameField.SetError("")
    // Hand name.Text to the caller's application logic.
})
panel := card.New(container.NewVBox(nameField, save), card.Panel)
```

Use `Object()` for inputs/selects rather than inserting their embedded native
controls directly: it applies the local theme and minimum touch size. Use native
`Entry`/`Select` APIs for binding, validation, selection, and editing configuration.
Avoid changing `MultiLine` on TextInput; use Textarea instead.

Components scope their own themes. `design.Theme{}` can also be installed on a
standalone application's settings, as the gallery does; `design.Scope(content)`
applies it to a subtree. The tablet installs this theme application-wide.
Application-specific labels inside supplied content should use ordinary Fyne
labels to inherit selected foreground colors, or `design.Label` for a fixed tone.

## Tablet integration

The application uses `design.Theme` (also exposed as `ui.AppTheme`), `Card`,
`EmptyState`, and setup `ActionButton`, `Field`, `TextInput`, and `ModalShell`.
Fyne file/folder pickers inherit the application theme. Lesson controls retain
fixed logical sizing, fitted labels, long-press gestures, and progress overlays;
`actionbutton.PaintSurface` shares the library's surface styling with those
controls. Success remains green to distinguish completed work from selection.
PageShell is for flowing pages; lesson screens retain their device-sized layouts.

## Native conventions

Colors come directly from the web `theme.ts` and Tailwind configuration. Cards
use 24-unit corners; controls use 16-unit corners. Buttons have at least 44-unit
interactive targets, textareas at least 120 units. Width/fill is controlled by the
parent Fyne layout instead of CSS `fullWidth` classes. PageShell does not scroll
on its own; wrap it in `container.NewVScroll` where appropriate.

Hover uses color/border feedback rather than moving controls. Keyboard focus has
a visible ring. The fixed light palette matches the source collection even when
the operating system requests a dark theme. Rubik regular/italic and semibold
styles are embedded; see `design/fonts/README.md` for provenance and license.

ModalShell centers a panel, caps its height, and scrolls longer content. `Options.Footer` keeps confirmation actions visible while the body scrolls. Close,
backdrop, and Escape dismissal invoke `OnClose` once per open lifecycle and
restore focus. Library controls forward Escape to their canvas's topmost modal.
Custom third-party focusable controls should call `design.HandleEscape(control,
key)` from their `TypedKey` handler before their normal handler to support Escape
while focused. Closing a parent modal also closes child modals. Call `Hide` before
destroying a window with an open modal. All mutations follow Fyne's UI goroutine
rules; use `fyne.Do` when updating from a background goroutine.

## Verification

```sh
make test
make vet
# Optional software renders, including desktop/narrow examples and a modal:
REC_UI_SCREENSHOTS=/tmp/rec-ui-screens go test -tags ci ./src/gallery -count=1
# Native launch check:
./dist/native/ui-gallery --smoke-test
```
