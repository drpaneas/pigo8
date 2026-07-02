# API Reference

This section documents every public function, type, and constant in the PIGO8 game-development
API - organized by domain, with full parameter descriptions and runnable examples.

If you want a fast one-line-per-function lookup instead, see the [Cheatsheet](../cheatsheet.md).
This reference is the deep version: full explanations, parameter tables, and examples for each
function.

## How to Read an Entry

Each function entry follows this format:

- **Signature** - the exact Go function signature, including generic type parameters where used.
- **Description** - what it does and how it behaves, including edge cases (out-of-range
  coordinates, invalid color indices, etc.).
- **Parameters** - a table of every parameter, its type, and its meaning.
- **Example** - a runnable snippet showing typical usage.
- **See also** - links to the matching conceptual guide page.

## Generic Coordinate Types

Many drawing functions (`Spr`, `Line`, `Rect`, `Circ`, `Mget`, `Sget`, and others) accept
coordinates as Go generics constrained to a `Number` type, meaning you can pass `int`,
`float64`, or any other standard numeric type interchangeably without manual conversion:

```go
p8.Spr(1, 10, 20)         // int coordinates
p8.Spr(1, 10.5, 20.0)     // float64 coordinates
p8.Spr(1, 10, 20.0)       // mixed - also valid
```

## Advanced / Internal APIs

PIGO8 also exports a number of lower-level types used internally for performance optimization:
sprite and pixel caches (`SpriteImageCache`, `SpritePixelCache`, `LRUCache`), batching systems
(`PixelBatchSystem`, `SpriteBatchSystem`), resource tracking (`ResourceManager`,
`ResourceLimits`), and render tuning (`RenderConfig`, `SetRenderConfig`). These exist to support
the engine's own rendering pipeline and advanced tuning scenarios, not everyday game code. They
are intentionally not covered in this reference; if you need them, browse the source directly
with `go doc github.com/drpaneas/pigo8 <Name>`.

## Reference Pages

- [Screen & Drawing](01-screen-drawing.md)
- [Colors & Palette](02-colors-palette.md)
- [Sprites](03-sprites.md)
- [Maps](04-maps.md)
- [Input](05-input.md)
- [Audio](06-audio.md)
- [Camera](07-camera.md)
- [Collision](08-collision.md)
- [Math](09-math.md)
- [Settings & Lifecycle](10-settings-lifecycle.md)
