# Neon Relay assets

Every file in `images/` and `audio/` is an original, deterministic asset made
specifically for this repository. The generator uses only authored geometry,
colors, and synthesized waveforms in the Go standard library. No downloaded
art, typeface, sample, sound effect, music, or other third-party media is used.

## Generate and verify

Run from the repository root:

```sh
go run ./cmd/genassets
```

This rewrites every generated PNG and WAV. To verify that the committed files
are current without changing them, run:

```sh
go run ./cmd/genassets -check
```

`go test ./cmd/genassets` performs the same byte-for-byte freshness check and
also validates all PNG dimensions and canonical WAV headers. Images are drawn
at 3× their final size and box-filtered for deterministic antialiasing. Noise in
the synthesized effects comes from fixed xorshift seeds.

## Palette

The artwork uses a graphite, cyan, and pink electronics palette. Gradients use
deterministic interpolation between these colors and a few documented highlight shades in the
generator source.

| Name | Hex | Use |
| --- | --- | --- |
| Graphite | `#161C22` | matte playfield and smoked plastics |
| Silver | `#C9D7DC` | machined edges and screw highlights |
| Relay pink | `#FF3081` | rubber, LEDs, and the relay identity |
| Ink | `#050816` | page and deep playfield shadow |
| Panel | `#0A122B` | recessed table components |
| Cyan | `#20E8FF` | primary rails and live circuitry |
| Cyan white | `#CFFDFF` | hot highlights |
| Magenta | `#FF2BAD` | targets, flippers, and secondary traces |
| Violet | `#7E47FF` | relay rings and ambient energy |
| Lime | `#A9FF4F` | reserved award/status accent |
| Amber | `#FFBE37` | reserved warning/bonus accent |
| Steel | `#61717B` | ball and plunger metal |

## Image inventory

All dimensions are exact final PNG pixel dimensions. Sprite canvases other
than the opaque background are transparent.

| File | Dimensions | Purpose |
| --- | ---: | --- |
| `images/background.png` | 720×1080 | graphite circuit playfield with baked hardware |
| `images/logo.png` | 640×200 | Neon Relay circuit wordmark |
| `images/favicon.png` | 64×64 | compact relay-mark icon |
| `images/ball.png` | 64×64 | shaded steel pinball |
| `images/flipper.png` | 180×64 | ivory flipper with pink rubber and steel pivot |
| `images/bumper.png` | 128×128 | smoked relay cap with segmented LEDs |
| `images/post.png` | 48×48 | steel post with cyan rubber ring |
| `images/target.png` | 64×96 | raised pulse target face |
| `images/target-down.png` | 64×96 | recessed, unlit target housing |
| `images/lane-light.png` | 48×96 | illuminated rollover chevrons |
| `images/lane-light-off.png` | 48×96 | unlit rollover insert |
| `images/plunger.png` | 56×180 | spring plunger assembly |
| `images/glow.png` | 192×192 | soft cyan-violet additive glow |
| `images/particle.png` | 32×32 | eight-point impact spark |

## Rendering approach

`cmd/genassets/art.go` draws the electronics-themed artwork. The background
bakes layered rail highlights, contact shadows, fasteners, and slingshot
plastics directly from `internal/table.New()` so the visible hardware follows
the collision geometry. Regenerate assets after changing the table definition.
Moving mechanisms and the steel ball remain transparent sprites; lane and
target state select illuminated or recessed variants at runtime. The browser
uses a system monospace font for live score displays and instrument labels.

The visual direction follows the Neon Relay concept developed during design;
no generated concept bitmap is embedded in the game. All shipped pixels remain
reproducible from the authored Go geometry.

## Audio inventory

All WAV files are mono, 44,100 Hz, signed 16-bit PCM. Oscillators, envelopes,
frequency sweeps, chords, transient clicks, and filtered seeded noise are built
sample by sample by `cmd/genassets`.

| File | Duration | Character |
| --- | ---: | --- |
| `audio/flipper.wav` | 0.16 s | solenoid click and low mechanical thump |
| `audio/bumper.wav` | 0.31 s | bright three-part spring chime |
| `audio/target.wav` | 0.20 s | short relay tick and descending knock |
| `audio/launch.wav` | 0.62 s | rising coil sweep with release snap |
| `audio/jackpot.wav` | 1.34 s | ascending five-note relay fanfare |
| `audio/drain.wav` | 0.92 s | two-voice falling energy sweep |
| `audio/game-over.wav` | 1.65 s | four descending minor-color chords |

## Provenance and license status

The design, drawing instructions, letter paths, procedural layout, and audio
synthesis recipes were authored for Neon Relay in `cmd/genassets/main.go` and
`cmd/genassets/art.go`. The committed binaries are direct outputs of that source and carry the same project
license as the rest of this repository. Because generation consumes no external
inputs, the source plus its fixed constants are the complete provenance trail.
