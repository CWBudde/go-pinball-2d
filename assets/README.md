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

`go test ./cmd/genassets` compares decoded image pixels and exact audio bytes and
also validates all PNG dimensions and canonical WAV headers. Images are drawn
at 3× their final PNG size and box-filtered for deterministic antialiasing.
The Phase 2 bumper layers ship at 2× logical resolution (6× authoring samples);
their filter averages premultiplied colors to preserve clean transparent edges.
Noise in the synthesized effects comes from fixed xorshift seeds.

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
| `images/background.png` | 720×1080 | flat graphite circuit playfield and printed title |
| `images/table-shadows.png` | 720×1080 | transparent static contact shadows |
| `images/table-hardware.png` | 720×1080 | rails, routing plate, slingshot covers, and apron panels |
| `images/table-foreground.png` | 720×1080 | cabinet perimeter, display housings, and drain lip |
| `images/logo.png` | 640×200 | Neon Relay circuit wordmark |
| `images/favicon.png` | 64×64 | compact relay-mark icon |
| `images/ball.png` | 64×64 | shaded steel pinball |
| `images/flipper.png` | 180×64 | ivory flipper with pink rubber and steel pivot |
| `images/bumper.png` | 128×128 | original bumper treatment, retained on right and center |
| `images/bumper-material.png` | 384×384 | Phase 2 raised metal/rubber/glass body for the left bumper |
| `images/bumper-patch.png` | 384×384 | flat textured graphite, routed pads, and reflected LED pools |
| `images/bumper-shadow.png` | 384×384 | directional contact shadow and soft penumbra |
| `images/bumper-emission.png` | 384×384 | segmented diffusers, narrow LED cores, and local bloom |
| `images/post.png` | 48×48 | steel post with cyan rubber ring |
| `images/target.png` | 64×96 | raised pulse target face |
| `images/target-down.png` | 64×96 | recessed, unlit target housing |
| `images/lane-light.png` | 48×96 | illuminated rollover chevrons |
| `images/lane-light-off.png` | 48×96 | unlit rollover insert |
| `images/plunger.png` | 56×180 | spring plunger assembly |
| `images/glow.png` | 192×192 | soft cyan-violet additive glow |
| `images/particle.png` | 32×32 | eight-point impact spark |

## Rendering approach

`cmd/genassets/art.go`, `cmd/genassets/layout.go`, and
`cmd/genassets/bumper_material.go` draw the electronics-themed artwork.
Runtime composition follows this order:

1. Flat playfield artwork and printed circuitry/title, then the bumper study patch.
2. Bumper study shadow and static contact shadows.
3. Static hardware, followed by lane inserts, bumpers (body then emission), posts, and targets.
4. Moving flippers, plunger, and ball.
5. Foreground cabinet/display housings and the lip below the drain sensor.
6. Emission/effects and the instrument HUD.

Rails and slingshots use the same geometry as `internal/table.New()`. Upper guide
curves are tessellated once in the table definition for both drawing and physics.
Regenerate assets after changing the table definition. Shared source bounds and
anchors in `internal/table/artwork.go` keep sprites aligned with physical centers
and flipper pivots/tips during rotation.

Printed routing plates have no height. Apron panels occupy sealed regions outside
the lower walls, and foreground covers stay outside live ball travel. Sprite
padding and shadows are decorative; opaque contact edges fit the physical shapes.
Future raised covers must explicitly define ball clearance and draw order.
Lane and target states select illuminated or recessed variants at runtime.
The browser uses a system monospace font for live displays; software captures
currently use Go Mono, so text rasterization can differ.

The visual direction follows the Neon Relay concept developed during design;
no generated concept bitmap is embedded in the game. All shipped pixels remain
reproducible from the authored Go geometry.

## Phase 2 material treatment

The upper-left bumper (`bumper_left`) is the material study; the right and center
bumpers retain the earlier finish until Phase 3. A cool softbox above/left drives
metal highlights and a down-right shadow. Brushed chrome has narrow bright
reflections separated by dark bands; graphite and molded rubber have broad, dim
shading. The inset smoked-glass cap has a soft diagonal reflection and a cyan relay
schematic. Cyan LEDs occupy its upper-left sector, with pink around the rest.
Reflections on the cylinder and small light pools on the playfield share those colors.

All four layers use `BumperMaterialFrame`: 384×384 source pixels, anchor (192,192),
contact radius 108. At the current table size they occupy 192×192 logical pixels
around the unchanged radius-54 body. The raised cap is visually offset upward
inside that footprint. The patch is flat; shadow and emission have no collision
surface. All four layers render before the ball. Tests check the opaque footprint,
transparent margins, and alpha filtering; engine crops check placement at 0.5×,
1×, and 2×. The separate emission is an idle appearance; impact states remain Phase 4.

The source art is authored Go in `cmd/genassets/bumper_material.go`; the four PNGs
are generated outputs in `assets/images/`. There are no curated raster inputs to
preserve in this phase. Do not hand-edit generated PNGs. If later work introduces
painted sources, store them outside `images/` and `audio/` (for example
`assets/sources/`) and explicitly add the conversion step to the generator.
Freshness verification currently covers 21 PNGs and seven WAVs.

The study adds four cached textures (about 148 KiB PNG total, 2.25 MiB decoded
RGBA) and three image draws per frame relative to the previous one-sprite bumper.
All material shading is offline; runtime uses the existing image compositor.

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
`cmd/genassets/art.go`, `cmd/genassets/layout.go`,
`cmd/genassets/bumper_material.go`, and the shared table definition. The committed
binaries are direct outputs of that source and carry the same project license as
the rest of this repository. Because generation consumes no external
inputs, the source plus its fixed constants are the complete provenance trail.
