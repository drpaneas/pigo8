# Text and Print

## Print Function

```go
p8.Print(value, x, y, color)
```

Draws text at (x, y). The `value` can be any type—it's converted to a string.

### Parameters

| Parameter | Type | Description |
|-----------|------|-------------|
| value | any | Text or value to display |
| x | int | X coordinate (optional) |
| y | int | Y coordinate (optional) |
| color | int | Color index (optional) |

### Examples

```go
p8.Print("hello", 10, 10, 7)     // White text at (10, 10)
p8.Print(score, 10, 20, 11)      // Print an integer
p8.Print(3.14159, 10, 30, 9)     // Print a float
p8.Print(true, 10, 40, 8)        // Print a boolean
```

## Flexible Arguments

Print supports multiple calling conventions:

```go
// Position and color
p8.Print("text", 10, 20, 7)

// Position only (uses current color)
p8.Print("text", 10, 20)

// Color only (uses cursor position)
p8.Print("text", 12)

// No arguments (uses cursor position and color)
p8.Print("text")
```

## Cursor Function

Set the text cursor position and color:

```go
p8.Cursor(10, 20)      // Set position to (10, 20)
p8.Cursor(10, 20, 8)   // Set position and color
p8.Cursor()            // Reset to (0, 0)
```

After `Print`, the cursor moves down by one line (6 pixels).

## Text Layout

Characters are 4 pixels wide (including spacing) and 6 pixels tall.

```go
// Measure text width
text := "hello"
width := len(text) * 4  // 20 pixels

// Center text horizontally
x := (128 - width) / 2
p8.Print(text, x, 60, 7)
```

## Return Values

`Print` returns the position after the text:

```go
endX, endY := p8.Print("score: ", 10, 10, 7)
p8.Print(g.score, endX, 10, 11)  // Continue on same line
```

## Example: Score Display

```go
func (g *game) Draw() {
    p8.Cls(0)
    
    // Title
    p8.Print("SPACE SHOOTER", 30, 5, 7)
    
    // Score (right-aligned)
    scoreText := fmt.Sprintf("%05d", g.score)
    p8.Print("SCORE:", 85, 5, 6)
    p8.Print(scoreText, 105, 5, 11)
    
    // Lives
    p8.Print("LIVES:", 5, 5, 6)
    for i := 0; i < g.lives; i++ {
        p8.Spr(1, 30+i*10, 3)  // Heart sprites
    }
}
```

## See Also

- [Screen & Drawing](../reference/api/01-screen-drawing.md)

