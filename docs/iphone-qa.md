# iPhone QA over TigerVNC and noVNC

This workflow runs the real native Go/Fyne app on the Arch Linux PC inside a
separate TigerVNC X server. noVNC exposes that virtual display to Safari over
the local network. SQLite and the application process stay on the PC.

The default framebuffer is fixed at 800 × 480. noVNC is forced to use local
browser scaling, and Xvnc rejects requests to change the remote display size.
The normal desktop launch remains unchanged.

## Prerequisites

- The normal Rec Tablet build requirements from the README: Go 1.25 or newer,
  a C toolchain, `pkgconf`, Mesa/OpenGL, and the X11 development libraries.
- An Arch/Omarchy user session with `systemd --user`, plus `bash`, `curl`,
  `python`, `openssl`, `iproute2`, `xorg-xrandr`, and `xorg-xprop`.
- `pacman`, `pacman-key`, `bsdtar`, and the Arch keyring when TigerVNC needs to
  be installed locally. Installing the system `tigervnc` package skips this
  local extraction.
- ImageMagick is optional; `make qa-verify` uses its `import` command to save a
  screenshot when available.
- The PC and iPhone must share a LAN that permits client-to-client traffic.
  Guest Wi-Fi commonly blocks devices from reaching each other.

## One-time setup

From the repository root, run:

```bash
make qa-setup
```

The setup asks for a VNC password. TigerVNC's VNCAuth format uses 6–8
characters, so use a password unique to this local QA session. The obfuscated
password file is stored under `.qa/credentials` with owner-only permissions.

On Arch, the launcher first uses an installed
[`tigervnc`](https://archlinux.org/packages/extra/x86_64/tigervnc/) package. If
it is not installed, it resolves TigerVNC and its missing dependencies through
`pacman`, downloads the signed packages from the configured Arch/Omarchy
mirror, verifies each package with `pacman-key`, and extracts them under
`.qa/deps`. This avoids requiring root for project-local QA. To install
TigerVNC system-wide instead:

```bash
sudo pacman -S --needed tigervnc
```

The launcher downloads pinned
[noVNC 1.7.0](https://github.com/novnc/noVNC/releases/tag/v1.7.0) from the
official noVNC GitHub release and installs
[websockify 0.13.0](https://github.com/novnc/websockify/releases/tag/v0.13.0)
in `.qa/deps/websockify-venv`. It also generates a self-signed HTTPS
certificate for the current LAN address because current noVNC requires a secure
browser context. Dependencies, credentials, configuration, logs, screenshots,
and databases under `.qa` are ignored by Git.

If setup was run non-interactively during automated verification, replace its
temporary credential before using the phone:

```bash
make qa-stop
./scripts/iphone-qa password
```

## Start, inspect, and stop

```bash
make qa-start
make qa-restart
make qa-watch
make qa-status
make qa-verify
make qa-stop
```

`make qa-start` rebuilds the native application, starts a transient user
systemd service containing Xvnc, Rec Tablet, and websockify, waits for every
component, then prints the exact Safari URL. The user service keeps running
after the terminal closes. `make qa-stop` checks the unit and recorded command
lines before stopping anything; it will not kill a process that does not match
this workflow.

Use `make qa-restart` after a requested change. It builds a replacement binary
first and restarts only Rec Tablet, so the virtual display, browser connection,
URL, and QA database stay in place. If the build fails, the currently running
app is left untouched.

During active development, run `make qa-watch` in a terminal. It watches Go and
embedded UI asset changes under `cmd` and `internal`, plus `go.mod` and
`go.sum`, then performs the same safe app-only restart after each save. Stop the
watcher with Ctrl-C; the QA session continues running until `make qa-stop`.

`make qa-status` reports all process IDs, the display and resolution, the QA
database, and the current URL. `make qa-verify` confirms the X server geometry,
finds the native `Rec Tablet` window, fetches the noVNC page, and captures the
virtual display to `.qa/screenshots/latest.png`. Use this when startup fails:

```bash
./scripts/iphone-qa logs
```

The latest logs are in `.qa/logs/{supervisor,xvnc,app,websockify}.log`.

## Connect from iPhone Safari

1. Put the iPhone and PC on the same trusted Wi-Fi network. Disable iPhone
   orientation lock and rotate the phone to landscape.
2. Run `make qa-start` and open the exact HTTPS URL it prints. The IP address
   can change after reconnecting the PC to Wi-Fi, so use the newly printed URL.
3. Safari will warn that the local certificate is not trusted because it is
   self-signed. Confirm that the address matches the one printed by the
   launcher before continuing to this local site.
4. At the noVNC Credentials prompt, enter the VNC password created with
   `./scripts/iphone-qa password`, then choose **Send credentials**.
5. In noVNC Settings, **Scaling mode** is locked to **Local scaling**. This
   scales the 800 × 480 framebuffer to the Safari viewport without changing the
   app's layout or virtual display resolution.
6. Tap the app controls directly. For text entry, open noVNC's left control bar,
   tap its keyboard control, then tap the Fyne text field and use the iPhone
   keyboard. Safari does not provide normal alphanumeric input in browser
   fullscreen, so keep the page in its regular Safari view while testing text.

The VNC server listens only on `127.0.0.1`/`::1`; the phone cannot reach the VNC
port directly. Only the HTTPS noVNC port is bound to the detected LAN address.
VNC password authentication still applies through the WebSocket bridge. Keep
this workflow on a trusted local network and stop it when testing is finished.

This machine uses UFW with incoming traffic denied. Add a rule limited to the
detected Wi-Fi subnet and configured noVNC address/port before connecting the
phone:

```bash
./scripts/iphone-qa firewall
```

The command prints the exact rule, then `sudo` asks for the PC password. On the
network used during verification it is equivalent to:

```bash
sudo ufw allow from 10.0.0.0/24 to 10.0.0.89 port 6080 proto tcp comment 'Rec Tablet iPhone QA'
```

Run the helper again after changing networks if the LAN address or subnet
changes. Do not open port 5990; it intentionally remains loopback-only. Remove
the firewall rule later with `sudo ufw status numbered` followed by
`sudo ufw delete NUMBER` if this QA access is no longer needed.

## Configuration

Copy the example and edit the copy:

```bash
cp config/iphone-qa.env.example config/iphone-qa.env
```

The copy is ignored by Git. Available settings include:

- `QA_WIDTH` and `QA_HEIGHT` (defaults: `800` and `480`)
- `QA_DISPLAY` (default: `90`)
- `QA_VNC_PORT` and `QA_WEB_PORT` (defaults: `5990` and `6080`)
- `QA_LISTEN_HOST` and `QA_PUBLIC_HOST` (normally auto-detected)
- `QA_FYNE_SCALE` (default: `1`)
- `QA_FULLSCREEN=1` to opt into app fullscreen inside the virtual display
- `QA_SOFTWARE_RENDERING=0` to try hardware rendering instead of Mesa llvmpipe
- `QA_TLS=0` only for diagnosing older browsers; current noVNC expects HTTPS
- `QA_BUILD=0` to reuse the existing `dist/native/rec-tablet` binary

Stop the session before changing ports, display, resolution, listener address,
or database settings, then start it again. Startup rejects occupied ports and X
display numbers.

## Data isolation and rendering

The QA database is persistent at:

```text
.qa/data/rec-tablet-qa/lessons.sqlite3
```

The launcher passes that path through the application's normal `--database`
option, so the usual schema initialization and migrations create it. It also
sets isolated XDG data, cache, and config directories under `.qa`. The normal
database at `$XDG_DATA_HOME/rec-tablet/lessons.sqlite3` is never opened by this
workflow.

`FYNE_SCALE=1` is explicit. Mesa software rendering (`LIBGL_ALWAYS_SOFTWARE=1`
and `GALLIUM_DRIVER=llvmpipe`) is enabled only inside the virtual session. This
avoids depending on Hyprland/Wayland, XWayland, or GPU access in Xvnc and does
not affect a normal desktop launch.

## What this proves

The virtual session is suitable for checking layout, paging, navigation,
tapping, text entry, attendance, assessments, timed heats, feedback, imports,
exports, and SQLite persistence. It confirms the x86-64 Linux app can render
through a fixed-size virtual X display and accept remote pointer and keyboard
events.

It does not measure Raspberry Pi performance, battery life, enclosure usability,
physical control size, display readability at poolside, or resistive-touch
accuracy. Those checks still require the target Pi, five-inch panel, and touch
hardware.
