# Rec Tablet

A native **Go + Fyne + SQLite** companion for swim instructors, designed for a five-inch landscape touchscreen. The app stores all lesson work on the device. It uses no Electron, webview, Python runtime, Qt runtime, or lesson-time network connection.

New to the codebase? Open the standalone [developer code guide](docs/code-guide.html) for an architecture tour, data model, core workflows, and pointers for common changes.

## Code organization

The code is grouped by workflow and feature. Under `internal/workflows/swimming`, `classes`, `attendance`, `assessment`, `feedback`, and `setup` each contain their model/repository interface, controller, view, and tests. Swimming's `domain` contains typed entities and rules; its `repository` owns SQL and the unchanged v1 schema. `module.go` constructs the features and registers their routes.

`internal/app` wires the app together. `internal/shell` hosts screens and common navigation controls, `internal/shared/ui` contains touch widgets, and `internal/platform` owns SQLite connection policy, file writing, and process locks. These packages do not import swimming. Future workflows can sit alongside swimming and provide their own routes, content, and data model.

The [code guide](docs/code-guide.html#changes) explains how to add a feature or workflow. Run `make test` to check feature behavior, compatibility, and package boundaries.

## Build and run

Use Go 1.25 or newer and a Linux C compiler, `pkg-config`, OpenGL, and X11 development libraries. On Debian/Ubuntu these are typically `build-essential pkg-config libgl1-mesa-dev xorg-dev`; on Arch they include `base-devel pkgconf mesa libx11 libxcursor libxrandr libxinerama libxi libxxf86vm`.

```bash
make build
./dist/native/rec-tablet
```

Or run directly:

```bash
go run ./cmd/rec-tablet
```

For the dedicated display:

```bash
./dist/native/rec-tablet --fullscreen
```

`F11` toggles fullscreen. `Alt+Left` or Escape goes back. The layout starts at 800 × 480 and scales to larger landscape sizes. The packaged desktop build uses Fyne's Linux graphical driver and requires a working OpenGL/X11 environment (including XWayland on Wayland desktops). The SQLite driver is Go-only; the graphics stack still has native system dependencies.

The database path remains `$XDG_DATA_HOME/rec-tablet/lessons.sqlite3`, falling back to `~/.local/share/rec-tablet/lessons.sqlite3`. Use `--database /path/to/lessons.sqlite3` for an isolated dataset. Close the old Python app before opening the same database in Go. Existing databases are read directly: no re-import or reset is required. A per-database process lock prevents concurrent app instances.

## Try the sample classes

1. Hold **My Classes** for three seconds, or press **Ctrl+Shift+M**.
2. Choose **Import Class Data**, then [examples/classes.json](examples/classes.json).
3. Press Home. Choose Date if needed; the fixture contains September 4, 11, 18, and 25, 2026.
4. Open a class and choose Attendance, Skills, or Feedback.

All fixture names are synthetic. It includes a paginated roster, a long name, two standard classes, and a private class with swimmers at different levels.

The same hidden screen provides **Export Data**. Select a folder, accept or edit the default filename, and confirm replacement if necessary. This lets the storage layer write atomically without a file picker truncating an existing export first. No records are cleared. These temporary file controls are intended for setup outside lessons; they are not network synchronization.

## During a lesson

- **Attendance:** tap a swimmer to toggle Present and Absent. The first tap marks an unmarked swimmer present. When a lesson date is past, any active swimmer without a present mark is automatically recorded absent.
- **Skills:** choose a skill, then choose a swimmer. Each swimmer shows zero to three green progress segments for that skill. Tap a swimmer to open its subskills and mark the criteria that were completed. The checklist stays open while you review the work; use the green checkmark rail button to save the attempt or the red X rail button to discard it.
- **Completion:** all subskills must succeed together for an attempt to count toward progress. Saved incomplete attempts are retained for later feedback but do not add a green progress segment. Three successful attempts complete the skill, including multiple attempts in one lesson. Failed attempts do not reset earned successes.
- **Corrections:** existing attempts remain reviewable through Feedback; corrections still save through the review screen and preserve revision history.
- **Undo:** hold a student button for two seconds to remove that student's newest successful check for the selected skill. The associated observation is removed transactionally.
- **Feedback:** shows skills assessed anywhere in the current enrollment that remain below three successes. It uses the latest saved attempt to identify unsuccessful criteria, or shows the additional successes needed. Unassessed and completed skills are excluded.

Home and Back preserve drafts. Selecting a swimmer for a skill resumes an unfinished observation for that swimmer, skill, and lesson. Existing attempts remain reviewable through Feedback. Tap the title/subtitle to read the full lesson context. Former enrollments remain available in Feedback but cannot receive new checks.

## Curriculum and transfer tools

The embedded [curriculum JSON](internal/workflows/swimming/curriculum/catalog.json) preserves the 23 sister-app templates: 22 assessable levels and one private-lesson reference, with 271 skills and 919 criteria. Unknown levels show Missing curriculum instead of using a fallback level. Existing courses keep their pinned curriculum version after software updates.

Both development tools now run in Go. Flags come **before** positional arguments:

```bash
./dist/native/rec-tablet adapt-rosters --dataset-id demo-autumn-2026 --instructor "Alex Instructor" examples/sister-roster.json /tmp/classes-adapted.json
./dist/native/rec-tablet extract-curriculum --output /tmp/curriculum.json ../COB-Aquatics-SPA/frontend/src/features/attendance-print/templates
```

The extractor preserves the previous JSON content, identifiers, and version hash. Compare a regenerated catalog before replacing the embedded file, then rebuild the app. The sister repository is read only.

The sister app's **Device Exports** page now produces directly importable instructor packages from its selected, loaded session. Review lesson dates and levels there, download the class JSON, then import it on this device. Keep its export ID backup for repeat exports from another browser. Each session has its own dataset and currently requires a separate device database.

See [data and architecture](docs/native-poc.md) for the unchanged import/export contract and identity details.

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
REC_TABLET_SCREENSHOTS=/tmp/rec-tablet-screens go test -tags ci ./internal/workflows/swimming -count=1
```

Tests cover the original POC workflows, recovery after abrupt process exit, failed writes, native widget interaction, curriculum extraction parity, and continuing work in a Python-created database. No Python is needed to run them. Template parity checks use the sibling repository when present; bundled-catalog and source-exception checks always run.

`make package` creates `dist/native/rec-tablet-linux-<architecture>.tar.gz`, containing an executable, desktop entry, icon, and documentation in a `usr/` directory tree. The catalog is embedded in the executable. Install on the same architecture with the required Linux graphics libraries. An x86-64 build does not establish ARM device compatibility.

Follow [the device acceptance trial](docs/device-trial.md) before a pool trial. Actual hardware, enclosure, wet-touch, and battery validation remain outstanding. Login, automated sync, free-text feedback, lesson plans, and timers remain outside this POC.
