# PICO-8 to PIGO8 Comparison

## Key Differences

| Aspect | PICO-8 | PIGO8 |
|--------|--------|-------|
| Language | Lua | Go |
| Array indexing | Starts at 1 | Starts at 0 |
| Code limits | 8192 tokens | None |
| Memory limits | 32KB | None |
| Resolution | Fixed 128×128 | Configurable |
| Platform | Fantasy console | Native + WebAssembly |
| License | Commercial ($15) | MIT (free) |

## Function Mapping

### Graphics

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `cls(c)` | `p8.Cls(c)` | Same |
| `pset(x,y,c)` | `p8.Pset(x,y,c)` | Same |
| `pget(x,y)` | `p8.Pget(x,y)` | Same |
| `line(x1,y1,x2,y2,c)` | `p8.Line(x1,y1,x2,y2,c)` | Same |
| `rect(x1,y1,x2,y2,c)` | `p8.Rect(x1,y1,x2,y2,c)` | Same |
| `rectfill(...)` | `p8.Rectfill(...)` | Same |
| `circ(x,y,r,c)` | `p8.Circ(x,y,r,c)` | Same |
| `circfill(...)` | `p8.Circfill(...)` | Same |
| `print(s,x,y,c)` | `p8.Print(s,x,y,c)` | Same |
| `color(c)` | `p8.Color(c)` | Same |
| `cursor(x,y,c)` | `p8.Cursor(x,y,c)` | Same |

### Sprites

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `spr(n,x,y,w,h,fx,fy)` | `p8.Spr(n,x,y,w,h,fx,fy)` | Same |
| `sspr(...)` | `p8.Sspr(...)` | Same |
| `sget(x,y)` | `p8.Sget(x,y)` | Same |
| `sset(x,y,c)` | `p8.Sset(x,y,c)` | Same |
| `fget(n,f)` | `p8.Fget(n,f)` | Returns (bitfield, isSet) |
| `fset(n,f,v)` | `p8.Fset(n,f,v)` | Same |

### Maps

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `map(mx,my,sx,sy,w,h,l)` | `p8.Map(mx,my,sx,sy,w,h,l)` | Same |
| `mget(x,y)` | `p8.Mget(x,y)` | Same |
| `mset(x,y,s)` | `p8.Mset(x,y,s)` | Same |

### Input

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `btn(b)` | `p8.Btn(p8.LEFT)` | Use constants |
| `btnp(b)` | `p8.Btnp(p8.X)` | Use constants |

### Palette

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `pal(c0,c1)` | `p8.Pal(c0,c1)` | Same |
| `pal()` | `p8.Pal()` | Reset |
| `palt(c,t)` | `p8.Palt(c,t)` | Same |

### Math

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `flr(x)` | `p8.Flr(x)` | Same |
| `rnd(x)` | `p8.Rnd(x)` | Returns int, not float |
| `sqrt(x)` | `p8.Sqrt(x)` | Same |
| `sin(x)` | `math.Sin(x)` | Use Go's math package |
| `cos(x)` | `math.Cos(x)` | Use Go's math package |
| `abs(x)` | `math.Abs(x)` | Use Go's math package |
| `min(a,b)` | `math.Min(a,b)` | Use Go's math package |
| `max(a,b)` | `math.Max(a,b)` | Use Go's math package |

### Audio

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `sfx(n)` | `p8.Music(n)` | Plays WAV files |
| `music(n)` | `p8.Music(n)` | Same function |

### Time

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `t()` | `p8.T()` or `p8.Time()` | Same |
| `time()` | `p8.Time()` | Same |

### Camera

| PICO-8 | PIGO8 | Notes |
|--------|-------|-------|
| `camera(x,y)` | `p8.Camera(x,y)` | Same |
| `camera()` | `p8.Camera()` | Reset to (0,0) |

## PIGO8 Extras

These functions are unique to PIGO8:

| Function | Description |
|----------|-------------|
| `p8.ColorCollision(x,y,c)` | Pixel-based collision |
| `p8.MapCollision(x,y,f,w,h)` | Tile-based collision |
| `p8.GetScreenWidth()` | Get screen dimensions |
| `p8.GetScreenHeight()` | Get screen dimensions |
| `p8.SetPalette(colors)` | Custom palettes |
| `p8.SetPaletteColor(i,c)` | Change single color |
| `p8.ClsRGBA(c)` | Clear with any color |

## Code Example Comparison

### PICO-8 (Lua)

```lua
function _init()
  x = 64
  y = 64
end

function _update()
  if btn(0) then x -= 1 end
  if btn(1) then x += 1 end
  if btn(2) then y -= 1 end
  if btn(3) then y += 1 end
end

function _draw()
  cls(0)
  spr(1, x, y)
end
```

### PIGO8 (Go)

```go
type game struct {
    x, y float64
}

func (g *game) Init() {
    g.x = 64
    g.y = 64
}

func (g *game) Update() {
    if p8.Btn(p8.LEFT)  { g.x-- }
    if p8.Btn(p8.RIGHT) { g.x++ }
    if p8.Btn(p8.UP)    { g.y-- }
    if p8.Btn(p8.DOWN)  { g.y++ }
}

func (g *game) Draw() {
    p8.Cls(0)
    p8.Spr(1, g.x, g.y)
}

func main() {
    p8.InsertGame(&game{})
    p8.Play()
}
```

## Migration Tips

1. **Arrays**: Change `arr[1]` to `arr[0]`
2. **Variables**: Declare with `var` or `:=`
3. **Functions**: Use `func` keyword
4. **Methods**: Attach to structs with receivers
5. **Math**: Import `math` package for sin/cos/etc.
6. **Strings**: Use `fmt.Sprintf` for formatting
7. **Tables**: Use structs or maps
8. **Local**: All Go variables are local by default

