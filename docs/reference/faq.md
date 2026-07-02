# FAQ & Troubleshooting

## Installation & Build

### "cgo: C compiler not found" or similar CGO errors

Ebitengine (which PIGO8 is built on) requires CGO and a C compiler. Install one for your
platform:

- **macOS**: `xcode-select --install`
- **Linux**: `sudo apt install gcc` (Debian/Ubuntu) or `sudo dnf install gcc` (Fedora)
- **Windows**: install [TDM-GCC](https://jmeubank.github.io/tdm-gcc/) or MinGW-w64 and ensure it's on your `PATH`.

See [Installation](../01-installation.md) for full platform notes.

### Linux build fails with missing audio/graphics libraries

Install the required development libraries before building:

```bash
# Debian/Ubuntu
sudo apt install libasound2-dev libgl1-mesa-dev xorg-dev

# Fedora
sudo dnf install alsa-lib-devel mesa-libGL-devel xorg-x11-server-devel
```

### `go build` for WebAssembly fails with "GOOS=js GOARCH=wasm" errors

Make sure you're not manually setting `GOOS`/`GOARCH` in your shell environment before running
`cmd/webexport` - it sets these itself. If building manually, use:

```bash
GOOS=js GOARCH=wasm go build -o game.wasm .
```

and copy `wasm_exec.js` from your Go installation's `$(go env GOROOT)/lib/wasm/wasm_exec.js`
(Go 1.24+) or `misc/wasm/wasm_exec.js` (earlier versions). See
[Web Export](../90-advanced/02-web-export.md) for the full workflow, which automates this.

## Runtime

### The window opens but shows a blank/black screen

- Make sure you call `p8.Cls(...)` at the start of your `Draw()` method - without clearing, the
  screen may show stale or undefined content on some platforms.
- Confirm `Draw()` is actually being called: add a temporary `p8.Print("draw", 0, 0, 7)` to
  verify.
- If you're loading sprites/maps from `spritesheet.json`/`map.json`, confirm those files are in
  the same directory as your executable when running `go run .` (or embedded - see
  [Resource Embedding](../90-advanced/01-resource-embedding.md)).

### My sprite doesn't show up / looks wrong

- Sprite indices are 0-based. Sprite 0 is the top-left sprite of your spritesheet.
- Check `Palt()` - if the sprite's background color happens to match your current transparent
  color, parts of it may render as "invisible" unexpectedly.
- If you recently edited `spritesheet.json` with the [PIGO8 Editor](../editor.md), confirm you
  saved (the editor autosaves on switching modes, but verify the file's modification time).

### Audio doesn't play on Linux

Confirm `libasound2-dev` (or your distro's ALSA development package) is installed - see the
Linux build section above. If audio still doesn't play, check that your audio files follow the
`music0.wav`, `music1.wav`, ... naming convention expected by `Music()`/`MusicLoop()`.

### Browser build: audio doesn't play until I click something

This is expected - browsers block autoplaying audio until the user interacts with the page.
The web-exported page's virtual controls handle this automatically; audio starts working after
the first button press. See [Web Export](../90-advanced/02-web-export.md#audio-notes).

### Movement feels too fast/slow after changing `TargetFPS`

Fixed per-frame movement values (like `g.x++`) move faster at higher FPS. Use `p8.Time()` to
compute frame-rate-independent movement instead of relying on a fixed FPS. See
[The Game Loop](../03-game-loop.md#time-function).

## Where to Get Help

If your issue isn't covered here, check the
[GitHub issues](https://github.com/drpaneas/pigo8/issues) or open a new one with a minimal
reproduction.
