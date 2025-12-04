# Controller Demo

This example demonstrates input handling by drawing an NES-style controller and highlighting buttons when they are pressed.

## What It Shows

- Drawing shapes (rectangles, circles) to create a controller UI
- Using `p8.Btn()` to detect button presses
- Visual feedback when buttons are pressed
- PICO-8 button mapping (0-5 for directions and face buttons)

## Controls

| Button | Key | Index |
|--------|-----|-------|
| Left   | Arrow Left | 0 |
| Right  | Arrow Right | 1 |
| Up     | Arrow Up | 2 |
| Down   | Arrow Down | 3 |
| O      | Z | 4 |
| X      | X | 5 |

## Running

```bash
cd examples/controller
go run main.go
```

## Web Demo

Try it in your browser: [Controller Demo](https://drpaneas.github.io/pigo8/controller/)

## Code Highlights

```go
// Check if a button is currently pressed
if p8.Btn(0) { // p8.LEFT
    // Highlight the left button
    p8.Print("0", 22, 66, 4)  // Brown color when pressed
}
```

The demo draws a classic NES controller layout and shows each button's index number. When you press a button, it changes color to provide visual feedback.

