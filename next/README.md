# Parallel UI implementation

`next/` is the default app UI. The original implementation stays in `src/` and can be selected at startup:

```sh
make run                             # next
make run ARGS='--ui=legacy'           # original
make run ARGS='--ui=next --database /tmp/comparison.sqlite3'
```

Both implementations use the existing database schema, import/export formats, profile index, and display preference. They cannot open the same database simultaneously because the existing process lock still applies. Use separate database copies to compare them side by side.

## Shift sessions

Every process launch mounts `features/login` through the normal shell. The login keypad verifies the six-digit SPA device export code online. `runtime/session.go` fetches the package asynchronously, imports into the existing instructor/dataset profile, and opens the workflow only after successful activation. Holding the login keypad’s Backspace button for three seconds explicitly restores the last activated local instructor session without redeeming a code. A tap still deletes one digit; an unavailable saved session shows an error and leaves the keypad open.

My Classes exposes a confirmed **Log out** action through `ConfigureSession`. Logout closes the workflow and instructor database, releases its additional process lock, and mounts a fresh login presenter. Local work is retained. The base database lock remains held for the process lifetime so a signed-out instance cannot race another instance over the profile index. Pending requests are cancelled and stale completions ignored; detached presenters, controls, and keyboard handlers are deactivated.

Logged-out Device Setup exposes server and display settings. Signed-in setup adds data export. Manual import and saved-profile switching are blocked in this UI. The legacy implementation keeps its existing development tools. No database-schema, package-format, or server-API changes are required.

## Where changes belong

| Change | Owner |
| --- | --- |
| Persistent screen composition | `app/layout.go` |
| Background, header, rails, content host, footer | Matching directory under `app/` |
| Feature layout and child composition | `features/<feature>/template.go` |
| Feature data loading, actions, and child coordination | `features/<feature>/controller.go` |
| Child presentation or behavior | Child's own `template.go` or `controller.go` |
| Cross-feature navigation and route history | `navigation/` |
| Feature registration and workflow assembly | Root `controller.go` |
| Window/session startup, profiles, imports, capture | `runtime/` |
| Feature-to-host presentation boundary | `contracts/` |

Every UI component directory has a `template.go` and `controller.go`. A stateless component can have a very small controller. Templates construct canvas objects and compose children. Controllers own loaded data, editing state, actions, and dependency wiring. Local `model.go`, `state.go`, and `actions.go` files are used when they keep these two entry points readable.

Each feature root selects compact/large settings in its template. Children receive explicit grid, calendar, sidebar, or panel settings and never inspect display mode. Layout changes that need a different child arrangement belong in the parent template. Shared assessment checklists and paged grids each have one owner under `assessment/internal/`; mode-specific templates compose those controls. Pure geometry and formatting helpers do not require component scaffolding.

Feature children take data, narrow storage operations, navigation callbacks, and action callbacks; parents connect siblings. In particular, feedback's summary accepts a detail-text callback instead of importing the detail component.

`app/` hosts any implementation of `contracts.Presenter`; it does not import the workflow. `runtime/` mounts the workflow into that host and owns window, database, profile, and asynchronous import lifetimes. This keeps the layout independently usable and testable.

## Feature overrides

Each feature builds a fresh `contracts.State`. Set `HeaderActions`, `HeaderSwitch`, or `SplitHeader` for header customization. Set `HomeRailAction` or `BackRailAction` to customize a default left-rail action, or supply `LeftRailActions` / `RightRailActions` to replace a whole rail in top-to-bottom order. A nil rail uses defaults; an explicit empty slice hides its controls. Overrides expire with the active screen, and detached buttons are released.

## Shared code and comparison

The new tree owns its UI, navigation, and controllers. It reuses `src/shared` controls, `src/platform` services, and swimming's domain, curriculum, repository, and transfer packages. It does not depend on the legacy app, shell, feature controllers, or view packages. Architecture tests enforce these boundaries and feature-root display selection.

This is an architectural comparison, so UI/controller duplication between the two implementations is intentional. Changes made to one implementation do not automatically change the other. The large skill picker in `next` exposes Single/Multi/Class mode controls so all assessment modes remain reachable.

```sh
make test                 # both trees, including integration flows
make vet
make race
make smoke-compare        # isolated databases and screenshots for both implementations
```

For representative compact/large screenshots from the new workflow integration tests:

```sh
REC_TABLET_SCREENSHOTS=/tmp/rec-next-screens go test -tags ci ./next/integration
```
