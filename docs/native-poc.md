# Native POC architecture and transfer contract

## Boundaries

`internal/workflows/swimming` owns swimming features, typed domain models, curriculum, transfer contracts, and SQLite schema. Each feature keeps its model/interface, controller, view, and tests together. Its repository applies domain rules within transactions. `internal/shell` hosts feature content and common navigation controls; `internal/shared/ui` supplies touch primitives. `internal/platform` owns database connection policy, atomic file writes, and process locks. `internal/app` assembles the desktop; `cmd/rec-tablet` handles CLI arguments and transfer tools. The shell and platform do not import swimming. See [the developer guide](code-guide.html) for the feature map and extension recipe. No HTTP service or browser renderer is started.

The Go implementation opens existing Python POC databases directly, using the same default path and schema version 1. Tests restore a Python-created SQL fixture and compare every exported table, then continue its unfinished assessment and correction. Import/export formats, source IDs, curriculum IDs, and the catalog version hash are unchanged. Close the previous app before opening its database; the Go process lock also respects a live legacy Qt lock.

The schema starts at version 1, migrated transactionally via `PRAGMA user_version`. SQLite uses foreign keys, WAL, FULL synchronous writes, and a busy timeout. Committed results and persistent draft edits are separate; replacing a result and adding its revision happen in one transaction. UI save errors remain visible until dismissed. A brief input guard prevents a tap carrying across a screen transition.

Courses pin a curriculum version. Progress belongs to an enrollment + skill + curriculum version, so students in different classes never share progress accidentally. Attendance and observations also reference a dated lesson. Feedback orders observations by lesson date, then creation time; correcting an older observation does not make it today's observation.

Program identity (`aquatics`) and curriculum definitions live in data, not UI conditionals. Other program areas can supply a curriculum later. Accounts, synchronization, configurable completion thresholds, and general-purpose plugin machinery are deferred.

## Import package v1

```json
{
  "schemaVersion": 1,
  "kind": "rec-tablet-classes",
  "datasetId": "durable-device-course-dataset",
  "instructor": "Alex Instructor",
  "courses": [{
    "sourceId": "course-12345-fall-2026",
    "name": "Splash 1",
    "level": "Splash1",
    "startTime": "16:00",
    "location": "Training Pool",
    "lessonDates": ["2026-09-04", "2026-09-11"],
    "swimmers": [{
      "sourceId": "durable-registration-person-id",
      "name": "Harrold",
      "level": "Splash1"
    }]
  }]
}
```

`sourceId` is required on every course and swimmer. Course IDs identify a specific course offering, not a recurring course name or reused course code. Swimmer IDs must remain stable within the dataset. Device UUIDs are assigned on first import, and the source references are retained. This POC does not infer IDs from a name, row index, or the sister app's generated roster IDs.

Use the same `datasetId` for subsequent imports. Another dataset requires a separate database (`--database`). Changing a name behind an existing swimmer source ID is rejected for manual resolution, including intentional spelling corrections; name-edit reconciliation is deferred. Different swimmers may have identical names if their durable source IDs differ.

Each listed course is a roster snapshot. Missing swimmers become inactive; their historical data remains. Reintroducing the same swimmer ID reactivates the same enrollment. Courses and dates omitted from a package remain on the device. Re-importing updates course display metadata and current enrollment levels; it never overwrites attempts or attendance. Use a new course ID for a new enrollment period.

Names and IDs are bounded nonempty strings. Times use `HH:MM`; dates use `YYYY-MM-DD`. Class files are limited to 10 MB in the UI, 1–1000 courses, up to 500 swimmers per course, and 1–366 dates per course. Whole-package validation precedes an atomic transaction. Empty rosters are valid. Swimmer level defaults to course level; private classes require an actual per-swimmer level. Unknown levels import for attendance but show Missing curriculum for assessment.

No phone numbers, email addresses, or authentication credentials are needed in the device package. Local roster identity and dates are authoritative for the POC; there is no concurrent server reconciliation.

## Temporary sister-app adapter

DeckSupervisor's `ClassRoster` supplies `serviceName`, `location`, and `students` with names, levels, and waitlist flags. Its shape lacks durable swimmer IDs and exact dated lesson occurrences. Annotate a roster export with course/swimmer `sourceId`, course `level`, `startTime24`, and explicit `lessonDates`, as shown in `examples/sister-roster.json`. Preserve those assigned IDs in future exports.

```bash
./dist/native/rec-tablet adapt-rosters --dataset-id demo-autumn-2026 --instructor "Alex Instructor" examples/sister-roster.json /tmp/classes-adapted.json
```

The adapter excludes waitlisted swimmers and contact fields. It does not parse ambiguous display schedule strings or automatically make IDs. Its output passes through the same import validation as any hand-authored package. No modifications to DeckSupervisor are required for this temporary path.

The sister app now provides **Device Exports** (`/device-exports`). Select and load a session, choose an instructor, review their class dates and swimmer levels, then download a package for this importer. It uses the existing loaded data and saved assignment/level APIs; no new export endpoint is needed. The standalone adapter remains available for annotated raw rosters.

The export page persists UUIDs in a browser registry and provides an ID backup/restore flow. Save that backup and restore it before exporting from another browser. Each session has one dataset ID shared by its instructor packages. A different session currently needs a separate device database. Swimmer IDs are retained per source course using an exact source signature, not the sister app's generated UI row IDs. Source identity changes require review; the registry is a POC substitute for upstream registration IDs. Reviewed lesson dates are retained in the backup. Automatic synchronization and receiving device results still require a separate design.

## Export snapshot v1

Exports have `kind: "rec-tablet-results"`, `schemaVersion: 1`, `databaseVersion: 1`, `exportedAt`, and a `tables` object containing:

- Device metadata and all versioned curriculum bodies.
- Courses, swimmers, enrollments (including inactive ones), and dated lessons.
- Attendance, attempts, submitted subskill results, persistent drafts, and prior submitted revisions.

Rows retain device and source IDs. Curriculum `body` and draft/revision `answers` columns are JSON strings, matching their SQLite representation. Submitted attempts have a non-null `submitted_at`; unfinished attempts do not count toward progress. Draft edits coexist with the previous submitted result until saved. Revision rows capture the prior result before each correction.

The database snapshot is read in one transaction, serialized to a temporary file, flushed, and atomically renamed. A failed export cannot replace a previous export with partial content. Export never deletes data. Results snapshots are not accepted by the class importer; restore and upstream synchronization are not implemented yet.

## Curriculum extraction review

Each level includes a source filename/hash; each skill preserves its source number, text lines, and assessable leaf criteria with group paths. The catalog version is a hash of extracted content. Re-extraction is deterministic and tested against the sister templates when available.

The extractor accounts for these source conventions and exceptions:

- Back-page headings are the canonical skill units; front-page Water Smart a/b/c/d columns summarize grouped criteria.
- Splash 2B has two distinct skills numbered 8; their IDs are `Splash2B:8` and `Splash2B:8-2`. The original display numbering is retained.
- Splash 8 places skills 12a, 12b, and 12c inside one paragraph; they are separate assessable skills.
- Splash 7 skill 16 splits its heading across two bold elements; the heading is rejoined.
- Stroke names, workout phases, and Water Smart headings remain context on leaf criteria. HTML source wrapping does not split a criterion; deliberate line continuations are rejoined.
- Participation requirements remain assessable criteria alongside individual Water Smart messages.
- SplashPrivate contains a reference list, not an independent curriculum. Actual assessment uses each swimmer's level, including Splash 2A or 2B rather than inferring from the private sheet's generic Splash 2 label.

Automated checks verify all 23 source hashes, unique identifiers, populated criteria, preservation of source words, known extraction exceptions, and deterministic regeneration. Source wording has not been rewritten or independently certified as current curriculum; instructors should review program content before operational adoption.
