# Web Export

Compile your game to WebAssembly for browser play.

## Quick Export

```bash
go run github.com/drpaneas/pigo8/cmd/webexport -game ./your-game -o ./dist
```

## Options

| Flag | Default | Description |
|------|---------|-------------|
| `-game` | (required) | Game directory path |
| `-o` | `./web-build` | Output directory |
| `-title` | "PIGO-8 Game" | Browser title |
| `-serve` | false | Start local server |
| `-port` | 8080 | Server port |

## Output Files

```
dist/
├── index.html    # Game page with virtual controls
├── game.wasm     # Compiled game
└── wasm_exec.js  # Go WASM runtime
```

## Testing Locally

```bash
cd dist && python3 -m http.server 8080
```

Open http://localhost:8080

Or use the built-in server:

```bash
go run github.com/drpaneas/pigo8/cmd/webexport -game ./my-game -o ./dist -serve
```

## Deployment

### GitHub Pages

```bash
go run github.com/drpaneas/pigo8/cmd/webexport -game ./my-game -o ./docs
```

Enable Pages in repo settings, point to `/docs`.

### itch.io

```bash
go run github.com/drpaneas/pigo8/cmd/webexport -game ./my-game -o ./dist
cd dist && zip -r ../game-web.zip .
```

Upload `game-web.zip` as HTML game.

### Netlify / Vercel

Just deploy the output directory as a static site.

## Mobile Support

The generated page includes touch controls:

- D-pad for movement
- A/B buttons for actions
- Haptic feedback on supported devices

The interface is styled like a Game Boy for nostalgia.

## Audio Notes

Browser autoplay policies require user interaction before audio. The page handles this automatically—audio works after the first button press.

## Customization

The generated `index.html` can be edited for custom styling:

- Change the Game Boy color scheme
- Add your own branding
- Modify button layout
- Add external links

## Complete Workflow

```bash
# 1. Create your game
mkdir my-game && cd my-game
go mod init my-game
go get github.com/drpaneas/pigo8

# 2. Write game code (main.go)

# 3. Add resources (spritesheet.json, etc.)

# 4. Export to web
go run github.com/drpaneas/pigo8/cmd/webexport -game . -o ./dist -title "My Game"

# 5. Test locally
cd dist && python3 -m http.server 8080

# 6. Deploy to your favorite hosting
```

