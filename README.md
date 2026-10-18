# Rec Tablet

A native **Go + Fyne + SQLite** companion for swim instructors, designed for a five-inch landscape touchscreen. The app stores all lesson work on the device. It uses no Electron, webview, Python runtime, Qt runtime, or lesson-time network connection.

New to the codebase? Open the standalone [developer code guide](docs/code-guide.html) for an architecture tour, data model, core workflows, and pointers for common changes.

## Code organization

The code is grouped by workflow and feature. Under `src/workflows/swimming`, `classes`, `attendance`, `assessment`, `feedback`, and `setup` each contain their model/repository interface, controller, view, and tests. Swimming's `domain` contains typed entities and rules; its `repository` owns SQL and the unchanged v1 schema. `module.go` constructs the features and registers their routes.

Feature UI uses `view.go` for screen selection, `view_large.go` and `view_small.go` where layouts differ, and `view_shared.go` for reusable rendering. Size-specific detail and timed views follow the same naming. Setup keeps a single view because its layout is shared across display sizes. Assembled workflow interaction tests live in `src/workflows/swimming/integration`; feature unit tests stay with their feature.

Classes is a parent component with named navigation and repository dependencies. It has `classlist`, `datepicker`, and `lessonmenu` child components, each receiving rendering data and callbacks. The date picker owns calendar browsing state; the parent handles data loading, pagination, and navigation. Children do not depend on navigation or storage.

Attendance is a parent component with named navigation and repository dependencies and a `roster` child. The parent loads lesson data, finalizes past attendance, saves toggles, and coordinates pagination. Roster builds active-swimmer rows, status labels and colors, and small-list or large-grid layouts using a selection callback.

Assessment is a parent component with named navigation and repository dependencies and five mode components: `SkillPicker`, `SingleAssessment`, `ClassAssessment`, `MultiAssessment`, and `TimedHeat`. Each mode owns its state, local pagination, and screen construction. The parent supplies lesson snapshots, read callbacks, persistence actions, and navigation callbacks. Rendering helpers in `assessment/view/large`, `assessment/view/small`, and `assessment/view/shared` receive explicit actions; they do not embed the navigation controller. TimedHeat owns ticker lifecycle, and MultiAssessment owns group assignment and local checklist edits. Parent persistence actions are grouped in `actions_*.go`.

Feedback is a parent component with named navigation and repository dependencies and `SwimmerPicker`, `SkillSummary`, and `SkillDetail` children. The parent loads data and preserves swimmer, skill, and curriculum-version navigation context. Children receive snapshots and callbacks; the large view combines the swimmer picker with scrollable skill summaries, using SkillDetail for criteria and timed-result text.

Setup is a parent component with `FileTransfer`, `SessionSharing`, `DisplayPicker`, and `ProfilePicker` children. Children own their menu items, dialogs, and input validation and emit callbacks. The parent coordinates file persistence, session downloads and status, server settings, display persistence, profile switching, and navigation. Child components have no repository or navigation-controller dependencies.

The swimming workflow has named navigation and feature dependencies. `module.go` constructs and registers the five features, and `navigation.go` deliberately forwards the workflow actions used by the app and shell. Consumers read screen state through `Snapshot()` and stage lesson dates through `SetDate()`. Navigation has a named `History` dependency and explicit history operations; features access that history directly where they already manage route frames.

`src/app` wires the app together. `src/shell` hosts screens and common navigation controls, `src/shared/components` provides the COB Aquatics design library, `src/shared/ui` adapts it for scaled lesson controls and long-press gestures, and `src/platform` owns SQLite connection policy, file writing, and process locks. These packages do not import swimming. Future workflows can sit alongside swimming and provide their own routes, content, and data model.

The [code guide](docs/code-guide.html#changes) explains how to add a feature or workflow. Run `make test` to check feature behavior, compatibility, and package boundaries.

## Build and run

Use Go 1.25 or newer and a Linux C compiler, `pkg-config`, OpenGL, and X11 development libraries. On Debian/Ubuntu these are typically `build-essential pkg-config libgl1-mesa-dev xorg-dev`; on Arch they include `base-devel pkgconf mesa libx11 libxcursor libxrandr libxinerama libxi libxxf86vm`.

```bash
make build
./dist/native/rec-tablet
```

Run the built executable with optional arguments:

```bash
make run ARGS="--fullscreen"
```

The first native build can take several minutes before a window opens while Go compiles the graphics dependencies; later builds use the cache. To start an existing build immediately, run `./dist/native/rec-tablet` directly.

For the dedicated display:

```bash
./dist/native/rec-tablet --fullscreen
```

To QA the native app from iPhone Safari while it runs in an isolated 800 × 480
virtual display on this PC, follow [iPhone QA over noVNC](docs/iphone-qa.md).
The short form is `make qa-setup`, `make qa-start`, and `make qa-stop`.

`F11` toggles fullscreen. `Alt+Left` or Escape goes back. The compact display profile uses an 800 × 480 logical layout; the large profile uses 1024 × 600. Each profile scales to the window while preserving its chosen layout. The packaged desktop build uses Fyne's Linux graphical driver and requires a working OpenGL/X11 environment (including XWayland on Wayland desktops). The SQLite driver is Go-only; the graphics stack still has native system dependencies.

## Display profiles

Compact is the default profile. To select a profile, hold **My Classes** for three seconds to open the hidden Device Setup screen, then choose **Display**. The preference is saved on the device independently of the active lesson dataset. Rec Tablet does not switch profiles automatically when the monitor or window size changes; choose the profile that matches the device.

The large profile provides six class cards in a two-column grid, a Sunday-first calendar with scheduled lesson dates, three side-by-side lesson choices, and attendance pages with twelve swimmers in two columns. Skills appear beside their swimmers; Single and Class marking keep a roster beside the complete checklist. Multi group assignment shows groups and swimmers together, while group marking uses a full-width table. Timed heat setup combines swimmer selection and lap entry. Feedback keeps swimmers beside their scrollable assessment details. Attempt history and Device Setup show six rows per page.

Private-level redesign is deferred. Private-lesson swimmers assigned a standard curriculum level continue using that level's assessment flow; unassigned swimmers cannot be assessed. All standard checklists fit without scrolling in the large profile.

The database path remains `$XDG_DATA_HOME/rec-tablet/lessons.sqlite3`, falling back to `~/.local/share/rec-tablet/lessons.sqlite3`. Use `--database /path/to/lessons.sqlite3` for an isolated dataset. Close the old Python app before opening the same database in Go. Existing databases are read directly: no re-import or reset is required. A per-database process lock prevents concurrent app instances.

## Try the sample classes

1. Hold **My Classes** for three seconds, or press **Ctrl+Shift+M**.
2. Choose **Import Class Data**, then [examples/classes.json](examples/classes.json).
3. Press Home. Choose Date if needed; the fixture contains September 4, 11, 18, and 25, 2026.
4. Open a class and choose Attendance, Skills, or Feedback.

All fixture names are synthetic. It includes a paginated roster, a long name, two standard classes, and a private class with swimmers at different levels.

The same hidden screen provides **Export Data**. Select a folder, accept or edit the default filename, and confirm replacement if necessary. This lets the storage layer write atomically without a file picker truncating an existing export first. No records are cleared. These file controls are intended for setup outside lessons.

## During a lesson

- **Attendance:** tap a swimmer to toggle Present and Absent. The first tap marks an unmarked swimmer present. When a lesson date is past, any active swimmer without a present mark is automatically recorded absent.
- **Skills:** choose a skill, then choose a swimmer. Each swimmer shows zero to three green progress segments for that skill. Tap a swimmer to open its subskills and mark the criteria that were completed. The checklist stays open while you review the work; use the green checkmark rail button to save the attempt or the red X rail button to discard it. Timed benchmark skills instead let you select a heat, enter the required laps, count laps with the shared timer, and save or discard the entire heat.
- **Evaluation modes:** Single marks one swimmer; Multi assesses instructor-assigned groups of present students; Class opens every active swimmer at the skill's level, one swimmer per page. Stroke skills and skills with more than six subskills offer Single and Class only. In Class, the header shows both skill and student, all subskills fit on the page, and the right arrows move between students without losing draft marks. The left rail, top to bottom, is ✓ Save current, X Discard current attempt, XX Discard all pending attempts, ✓✓ Save all pending attempts. Saving or discarding the current attempt advances to the next pending student. Batch actions finish the session without changing students already saved individually; saving an untouched or partially marked student records an incomplete attempt. Discard deletes pending attempts, including drafts resumed from an earlier visit.
- **Multi groups:** mark attendance first. Multi includes only present, active swimmers at the selected skill’s level. Use − and + to choose the number of groups and review the live balanced-size summary, including any larger or smaller groups, then tap Continue. Assign every student exactly once; use Previous/Next group to revisit assignments and the sidebar arrows to page through students. Start marking becomes available when every group is full. Changing the group count clears assignments. Each marking page shows the whole group, with sidebar arrows switching between pending groups. Tap a student’s ✓ to pass every subskill, or ✕ to select the subskills that passed. Checklist Save updates that student’s draft; Discard or Escape returns to the group without those edits. The left rail is ✓ Save group, X Discard group, XX Discard all pending groups, ✓✓ Save all pending groups. Saving includes untouched students as incomplete attempts. Discard restores drafts resumed at session entry and deletes newly created drafts; it never changes groups already saved. Groups last only for this marking session. Leaving preserves saved student drafts, but reopening Multi requires assigning groups again.
- **Completion:** all subskills must succeed together for an attempt to count toward progress. Saved incomplete attempts are retained for later feedback but do not add a green progress segment. Three successful attempts complete the skill, including multiple attempts in one lesson. Failed attempts do not reset earned successes.
- **Corrections:** existing attempts remain reviewable through Feedback; corrections still save through the review screen and preserve revision history.
- **Undo:** hold a student button for two seconds to remove that student's newest successful check for the selected skill. The associated observation is removed transactionally.
- **Feedback:** shows skills assessed anywhere in the current enrollment that remain below three successes. It uses the latest saved attempt to identify unsuccessful criteria, or shows the additional successes needed. Unassessed and completed skills are excluded.

Home and Back preserve drafts. Selecting a swimmer for a skill resumes an unfinished observation for that swimmer, skill, and lesson. Existing attempts remain reviewable through Feedback until a curriculum reset. Lesson, skill, student, and progress information appears in the screen’s main header area. Former enrollments remain available in Feedback but cannot receive new checks.

My Classes shows the selected lesson’s position in the shared schedule, such as **Class 1/12**, once in the split main header. Course rows stay focused on the course name and time. The available dates remain accessible from the Date button. Instructor names and calendar dates are not repeated in lesson subtitles.

Class marking uses a split header with the skill number and short title on the left and the full student name on the right. The main information line shows the course time, full skill title, and student position. Subskill buttons remain equal-sized and use one shared font size.

## Curriculum and transfer tools

The embedded [curriculum JSON](src/workflows/swimming/curriculum/catalog.json) preserves the 23 sister-app templates: 22 assessable levels and one private-lesson reference, with 290 skills and 937 criteria. Grouped multi-component skills are separate entries with the same skill number so each component can be marked independently. Unknown levels show Missing curriculum instead of using a fallback level. When the bundled curriculum version changes, assessment marking rows are reset and existing courses move to the new version while roster and attendance data are retained.

Edit Class-mode short titles in [compact_titles.json](src/workflows/swimming/curriculum/compact_titles.json), using the existing skill IDs as keys and omitting the skill number from the values. These are concise header labels; the full assessment instructions remain visible in the main information line. They are embedded display metadata, separate from the versioned curriculum: editing them requires a rebuild but does not reset markings, change exports, or need a curriculum version bump. Skills without an override use their original title.

Both development tools now run in Go. Flags come **before** positional arguments:

```bash
./dist/native/rec-tablet adapt-rosters --dataset-id demo-autumn-2026 --instructor "Alex Instructor" examples/sister-roster.json /tmp/classes-adapted.json
./dist/native/rec-tablet extract-curriculum --output /tmp/curriculum.json ../COB-Aquatics-SPA/frontend/src/features/attendance-print/templates
```

The extractor preserves the previous JSON content, identifiers, and version hash. Compare a regenerated catalog before replacing the embedded file, then rebuild the app. The bundled file also contains app-specific multi-component splits; regenerating it from the sister repository produces the unsplit source catalog and those overrides must be reapplied. The sister repository is read only.

The sister app's **Device Exports** page can publish a temporary share for its selected, loaded session. Review each instructor's roster, levels, and lesson dates, start sharing, then enter that instructor's six digit code from **Import from Session Code** in the hidden Device Setup menu. The tablet downloads and imports the package, then opens that instructor's saved session automatically. Codes work while the host page is open and expire after a minute without a host heartbeat. The host can end sharing at any time; roster changes require ending and starting a new share.

Each dataset and instructor has a separate SQLite database under `<database-directory>/<database-name>.sessions/`. The app records saved profiles and the active profile in `<database-path>.profiles.json`; the original database remains in place and is registered as a profile when it already contains an instructor dataset. Use the hidden **Switch Session** rows during development to change profiles. The active profile opens automatically on the next launch. Use **Configure Share Server** in the code keypad for development servers; the default is the hosted sister-app server. `REC_TABLET_SHARE_SERVER` or `--share-server URL` sets the initial server. HTTP is accepted only for localhost; remote servers must use HTTPS. Import and export files continue to operate on the active profile.

See [data and architecture](docs/native-poc.md) for the unchanged import/export contract and identity details.

## UI component gallery

Run `make gallery` to open the separate COB Aquatics style gallery. It demonstrates
the same component library used by the tablet without opening lesson storage. Each component has its own folder with paired UI and behavior Go files.
See [the component guide](src/shared/components/README.md) for APIs, styling
conventions, and verification commands.

## Tests and packaging

```bash
make test
make race
make vet
make smoke
make package
```

Headless tests and `make smoke` use Fyne's software test driver (`-tags ci`). The resulting `rec-tablet-headless` binary is a verification tool, not the desktop app. For native graphics verification on a graphical Linux session:

```bash
./dist/native/rec-tablet --database /tmp/rec-tablet-native-check.sqlite3 --smoke-test
```

Native launch and visible rendering were checked from the extracted Linux archive on the development desktop. Optional `--screenshot FILE.png` uses the graphics driver's capture API; that API returns a blank buffer on the tested Hyprland/XWayland setup, even though the window renders correctly. Blank captures report an error. Use the software screenshots below or capture the visible window for visual review.

To capture representative test screens:

```bash
REC_TABLET_SCREENSHOTS=/tmp/rec-tablet-screens go test -tags ci ./src/workflows/swimming/integration -count=1
REC_TABLET_SCREENSHOTS=/tmp/rec-tablet-large-screens go test -tags ci ./src/workflows/swimming/integration ./src/workflows/swimming/assessment -run TestLarge -count=1
```

Tests cover the original POC workflows, recovery after abrupt process exit, failed writes, native widget interaction, curriculum extraction parity, and continuing work in a Python-created database. No Python is needed to run them. Template parity checks use the sibling repository when present; bundled-catalog and source-exception checks always run.

`make package` creates `dist/native/rec-tablet-linux-<architecture>.tar.gz`, containing an executable, desktop entry, icon, and documentation in a `usr/` directory tree. The catalog is embedded in the executable. Install on the same architecture with the required Linux graphics libraries. An x86-64 build does not establish ARM device compatibility.

Follow [the device acceptance trial](docs/device-trial.md) before a pool trial. Actual hardware, enclosure, wet-touch, and battery validation remain outstanding. Login, automated sync, free-text feedback, and lesson plans remain outside this POC.
