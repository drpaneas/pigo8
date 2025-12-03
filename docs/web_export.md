# Web Export

PIGO8 includes a powerful web export tool that compiles your game to WebAssembly, allowing players to enjoy your game directly in their web browser.

## Quick Start

```bash
# Export your game
go run github.com/drpaneas/pigo8/cmd/webexport -game ./your-game -o ./dist

# Test locally
cd dist && python3 -m http.server 8080
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## Command Line Options

| Flag | Description | Default |
|------|-------------|---------|
| `-game` | Path to your game directory (required) | - |
| `-o` | Output directory for web files | `./web-build` |
| `-title` | Game title shown in browser tab | `PIGO-8 Game` |
| `-serve` | Start local HTTP server after build | `false` |
| `-port` | Port for local server | `8080` |

### Examples

```bash
# Basic export
go run github.com/drpaneas/pigo8/cmd/webexport -game ./examples/pong -o ./dist

# With custom title
go run github.com/drpaneas/pigo8/cmd/webexport -game ./examples/pong -o ./dist -title "PIGO-8 Pong"

# Export and start local server
go run github.com/drpaneas/pigo8/cmd/webexport -game ./examples/pong -o ./dist -serve -port 3000
```

## Output Files

The web export generates three files:

```
dist/
├── index.html      # Game page with Game Boy-style UI
├── game.wasm       # Compiled game (WebAssembly)
└── wasm_exec.js    # Go WASM runtime
```

## Game Boy-Style UI

The generated HTML page features a retro Game Boy-inspired interface:

- **D-Pad**: Left, Right, Up, Down navigation
- **Action Buttons**: Z (O button) and X (X button)
- **Start Button**: Pause/menu functionality
- **Responsive Design**: Works on desktop and mobile devices
- **Touch Support**: Virtual controls for smartphones and tablets
- **Haptic Feedback**: Vibration on button press (mobile devices)

### Button Mapping

| Virtual Button | Keyboard Key | PIGO8 Constant |
|---------------|--------------|----------------|
| D-Pad Left | Arrow Left | `ButtonLeft` (0) |
| D-Pad Right | Arrow Right | `ButtonRight` (1) |
| D-Pad Up | Arrow Up | `ButtonUp` (2) |
| D-Pad Down | Arrow Down | `ButtonDown` (3) |
| Z Button | Z | `ButtonO` (4) |
| X Button | X | `ButtonX` (5) |
| Start | Enter | `ButtonStart` (6) |

## Mobile Support

The web export is fully optimized for mobile browsers:

- **Touch Controls**: All virtual buttons respond to touch input
- **Viewport Optimization**: Prevents unwanted zooming and scrolling
- **Haptic Feedback**: Short vibration pulse on button press
- **Responsive Layout**: Adapts to different screen sizes

### Testing on Mobile

1. Deploy to a web server (or use a tool like [ngrok](https://ngrok.com/))
2. Open the URL on your mobile device
3. For gamepad support, ensure HTTPS is enabled

## Deployment

### GitHub Pages

1. Export your game:
   ```bash
   go run github.com/drpaneas/pigo8/cmd/webexport -game ./your-game -o ./docs
   ```

2. Enable GitHub Pages in your repository settings, pointing to the `/docs` folder.

3. Your game will be available at `https://yourusername.github.io/your-repo/`

### itch.io

1. Export your game:
   ```bash
   go run github.com/drpaneas/pigo8/cmd/webexport -game ./your-game -o ./dist
   ```

2. Zip the output:
   ```bash
   cd dist && zip -r ../game-web.zip .
   ```

3. Upload to itch.io:
   - Create a new project
   - Set "Kind of project" to **HTML**
   - Upload `game-web.zip`
   - Enable "This file will be played in the browser"
   - Set viewport dimensions (recommended: 640x480)

### Netlify / Vercel / Static Hosts

Simply upload the contents of your output directory. No special configuration is needed—PIGO8 generates static files that work on any web server.

## Technical Notes

### Asset Embedding

For web builds, all assets must be embedded in the binary using Go's `embed` package. Make sure you run:

```bash
go generate
```

Before exporting, if your game uses `spritesheet.json`, `map.json`, or audio files.

### Audio Considerations

Due to browser autoplay policies, audio is initialized only after the first user interaction. The Game Boy UI handles this automatically—audio will start working once the player presses any virtual button or key.

### Gamepad Support

Physical gamepads require HTTPS to function in browsers. For local testing, `localhost` is treated as a secure origin.

### WASM Size Optimization

The export tool automatically applies optimizations:

- `-ldflags="-s -w"` strips debug information
- Final WASM size is typically 10-15MB (3-4MB gzipped)

For production, configure your web server to serve `.wasm` files with gzip compression.

## Troubleshooting

### "Failed to load game.wasm"

Make sure you're serving the files from a web server, not opening `index.html` directly as a file. Use:

```bash
cd dist && python3 -m http.server 8080
```

### Game doesn't respond to input

Ensure you're clicking/touching inside the game canvas or using the virtual buttons. The first interaction also initializes audio.

### CORS errors

If loading assets from external URLs, ensure proper CORS headers are set on the asset server.

### Performance issues on mobile

WASM performance on mobile can vary. Consider:
- Reducing sprite count
- Simplifying game logic in `Update()`
- Using smaller screen resolutions

## Live Examples

Try PIGO8 games in your browser:

- [Pong](https://drpaneas.github.io/pigo8/pong/) - Classic Pong implementation
- [Animation](https://drpaneas.github.io/pigo8/animation/) - Sprite animation demo
- [Game Boy](https://drpaneas.github.io/pigo8/gameboy/) - Game Boy-style game

See the [examples directory](https://github.com/drpaneas/pigo8/tree/main/examples) for more.

