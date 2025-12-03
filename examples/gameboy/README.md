# 🎮 PIGO-8 Game Boy Template

![Game Boy DMG-01](https://upload.wikimedia.org/wikipedia/commons/7/7c/Game-Boy-FL.jpg)

## 🎮 Play Online

**[▶️ Play Game Boy Demo in your browser](https://drpaneas.github.io/pigo8/gameboy/)**

---

Get ready to dive into the nostalgic world of Game Boy development, powered by PIGO-8! This template lets you create authentic Game Boy-style games using the beloved PICO-8 workflow, but with the added power of Go and PIGO-8's extended feature set.

## 🌟 Features

- **Authentic Game Boy Resolution** (160x144)
- **PICO-8 Style Workflow** (sprites, maps, sound)
- **Extended PIGO-8 Powers** (Go's full capabilities!)
- **Modern Dev Experience** (hot-reload, debugging)

## 🚀 Quick Start

1. **Set up your Game Boy palette**

   ```bash
   # Copy the Game Boy palette.hex to your project
   cp path/to/gameboy/palette.hex .
   ```

2. **Create your assets**

   ```bash
   # Launch the editor in Game Boy mode
   go run . -w=160 -h=144
   ```

   - Design your sprites in the sprite editor
   - Create your map in the map editor
   - Save both `spritesheet.json` and `map.json`

3. **Embed your resources**

   ```bash
   # Make sure your main.go has this line:
   //go:generate go run github.com/drpaneas/pigo8/cmd/embedgen -dir .

   # Generate the embed.go file
   go generate
   ```

4. **Configure Game Boy settings**

   ```go
   settings := p8.NewSettings()
   settings.TargetFPS = 60
   settings.ScreenWidth = 160
   settings.ScreenHeight = 144
   settings.WindowTitle = "Game Boy Style Demo"
   ```

## 🎨 Development Tips

- Use the classic Game Boy 4-color palette for authenticity
- Keep your sprites within the 8x8 pixel constraint
- Remember the Game Boy's screen limitations when designing maps
- Take advantage of PIGO-8's extended features for modern touches

## 🌈 Why This is Cool

Remember those late nights playing Tetris or Pokemon under the covers with a flashlight? Now you can create those same magical experiences, but with modern tools! This template gives you:

- The constraints that made Game Boy games special
- The intuitive workflow of PICO-8
- The power and flexibility of Go
- A perfect blend of retro and modern

## 🎵 Pro Tips

1. Think in terms of 8x8 tiles, just like real Game Boy games
2. Use the map editor to create scrolling levels
3. Keep performance in mind - the real Game Boy had limitations!
4. Experiment with PIGO-8's extended features for unique twists

## 🔧 What's Next?

1. Create your sprites and maps
2. Implement your game logic in `main.go`
3. Test on different window sizes
4. Share your creation with the world!

Now go forth and create something awesome! Whether you're making a puzzle game, platformer, or RPG, you've got the perfect foundation to build your Game Boy dreams. 🌟

## 🤝 Contributing

Found a bug? Want to add a feature? PRs are welcome! Let's make retro game development even more awesome together.

---

Made with 💚 by the PIGO-8 community
