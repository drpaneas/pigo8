package pigo8

// Full-frame profiling harness. Exercises a realistic per-frame workload --
// clear, camera move, sprites, shapes, text -- so pprof output reflects real
// usage rather than isolated micro-benchmarks. See the measurement-validity
// note below before trusting absolute numbers from these benchmarks.

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func setupFrameBenchWorld(numSprites int) {
	initializeCaches()
	currentScreen = ebiten.NewImage(128, 128)

	sprites := make([]spriteInfo, numSprites)
	for i := range sprites {
		img := ebiten.NewImage(8, 8)
		img.Fill(color.RGBA{R: uint8(i * 7), G: uint8(i * 3), B: 200, A: 255})
		sprites[i] = spriteInfo{
			ID:    i,
			Image: img,
			Flags: FlagsData{Bitfield: 0, Individual: make([]bool, 8)},
		}
	}
	currentSprites = sprites
	InvalidateSpriteIDIndex()

	initShadowBuffer(128, 128)
	initScreenPixelCache(128, 128)
	pbs := getPixelBatchSystem()
	pbs.ensureBufferSize(128, 128)
}

// NOTE ON MEASUREMENT VALIDITY: outside a running ebiten.RunGame loop, Ebiten
// never flushes its internal deferred draw-command queue (there is no frame
// boundary to trigger it, and forcing a sync via Image.At()/ReadPixels()
// panics with "ui: ReadPixels cannot be called before the game starts").
// That means the *cumulative* CPU/alloc cost attributed to Ebiten's internal
// atlas/DrawTriangles machinery grows across iterations and does not reflect
// real per-frame steady-state cost. Always run these with a small, fixed
// iteration count (`-benchtime=Nx`, not time-based) so results are bounded
// and reproducible, and treat PIGO8-side "flat" costs (not Ebiten-internal
// "cum" costs) as the trustworthy signal from pprof on this benchmark.

// BenchmarkFullFrame_Typical simulates a frame with a moderate sprite count,
// a handful of shapes, camera movement and a few text draws -- representative
// of a small PICO-8-style game.
func BenchmarkFullFrame_Typical(b *testing.B) {
	setupFrameBenchWorld(64)
	EnableFrameStats(true)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Camera(float64(i%128), float64((i*2)%128))

		Cls(1)

		for s := 0; s < 40; s++ {
			x := (s*7 + i) % 120
			y := (s*5 + i/2) % 120
			Spr(s%64, x, y)
		}

		Rectfill(2, 2, 40, 12, 5)
		Rect(0, 0, 127, 127, 6)
		Circfill(64, 64, 10, 8)
		Line(0, 0, 127, 127, 7)

		Print("SCORE: 000100", 2, 2, 7)
		Print("LIVES: 3", 2, 10, 7)

		FlushSpriteModificationsForBench()
		FlushPixelBatch()
		FlushFrameMetrics()
	}
}

// BenchmarkFullFrame_ManySprites stresses the sprite path specifically with a
// much higher sprite count, similar to a busy shooter/particle-heavy scene.
func BenchmarkFullFrame_ManySprites(b *testing.B) {
	setupFrameBenchWorld(64)
	EnableFrameStats(true)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Cls(0)
		for s := 0; s < 300; s++ {
			x := (s*3 + i) % 120
			y := (s*11 + i/3) % 120
			Spr(s%64, x, y)
		}
		FlushPixelBatch()
		FlushFrameMetrics()
	}
}

// FlushSpriteModificationsForBench exposes the package-private flush for the
// benchmark harness above.
func FlushSpriteModificationsForBench() {
	flushSpriteModifications()
}

// BenchmarkPrint isolates Print()'s own per-call cost (string formatting,
// face/options setup). Because Print() draws onto currentScreen, it still
// queues Ebiten commands each call (see the measurement-validity note
// above), so always compare with an identical fixed -benchtime=Nx between
// runs: the Ebiten-side queue growth is then the same in both cases and
// cancels out, leaving the Print()-side allocation/CPU delta meaningful.
func BenchmarkPrint(b *testing.B) {
	setupFrameBenchWorld(1)
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		Print("SCORE: 000100", 2, 2, 7)
	}
}
