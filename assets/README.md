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
The playfield layers and mechanism sprites ship at 2× logical resolution
(6× authoring samples). Their filter averages premultiplied colors to preserve
clean transparent edges. The ball, logo, and legacy effects retain their earlier sizes.
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
| `images/background.png` | 1440×2160 | flat graphite circuit playfield, bank routing, and seams |
| `images/playfield-markings.png` | 1440×2160 | printed title, relay symbol, shot chevrons, and lower lane inserts |
| `images/table-shadows.png` | 1440×2160 | transparent static contact shadows |
| `images/table-hardware.png` | 1440×2160 | rails, routing plate, slingshot covers, and apron panels |
| `images/table-foreground.png` | 1440×2160 | cabinet perimeter, display housings, and drain lip |
| `images/logo.png` | 640×200 | Neon Relay circuit wordmark |
| `images/favicon.png` | 64×64 | compact relay-mark icon |
| `images/ball.png` | 64×64 | spherical chrome reflections with a dark horizon and cool key |
| `images/ball-shadow.png` | 96×64 | separate soft contact shadow |
| `images/flipper.png` | 360×128 | ivory blade, shaded rubber skirt, and machined pivot |
| `images/bumper.png` | 128×128 | legacy bumper asset, no longer drawn by the game |
| `images/bumper-material.png` | 384×384 | shared raised metal/rubber/glass body for all three bumpers |
| `images/bumper-hit.png` | 384×384 | compressed cap within the same opaque contact footprint |
| `images/bumper-patch.png` | 384×384 | flat textured graphite, routed pads, and reflected LED pools |
| `images/bumper-shadow.png` | 384×384 | directional contact shadow and soft penumbra |
| `images/bumper-emission.png` | 384×384 | segmented diffusers, narrow LED cores, and local bloom |
| `images/post.png` | 96×96 | machined post, rubber skirt, inset cyan cap and fastener |
| `images/target.png` | 128×192 | raised chrome/rubber pulse target face |
| `images/target-down.png` | 128×192 | flush, dark target socket with recessed contacts |
| `images/lane-light.png` | 96×192 | illuminated rollover chevrons |
| `images/lane-light-off.png` | 96×192 | unlit rollover insert |
| `images/plunger.png` | 112×360 | legacy combined plunger, no longer drawn |
| `images/plunger-head.png` | 72×24 | translating chrome head and fixed foot |
| `images/plunger-coil.png` | 64×16 | single spring turn, repeated with variable spacing |
| `images/plunger-rod.png` | 12×64 | polished shaft with variable exposed length |
| `images/glow.png` | 192×192 | soft cyan-violet additive glow |
| `images/particle.png` | 32×32 | legacy spark, replaced by short runtime strokes |

## Rendering approach

`cmd/genassets/art.go`, `cmd/genassets/layout.go`, and
`cmd/genassets/bumper_material.go`, and `cmd/genassets/hardware_material.go` draw the electronics-themed artwork.
Runtime composition follows this order:

1. Flat playfield and circuitry, then all bumper patches, then printed markings.
2. Bumper shadows and static contact shadows.
3. Static hardware, followed by lane inserts, bumpers (body then emission), posts, and targets.
4. Moving flippers, plunger, and ball.
5. Foreground cabinet/display housings and the lip below the drain sensor.
6. Emission/effects and the instrument HUD.

Rails and slingshots use the same geometry as `internal/table.New()`. Upper guide
curves are tessellated once in the table definition for both drawing and physics.
Regenerate assets after changing the table definition. Shared source bounds and
anchors in `internal/table/artwork.go` keep sprites aligned with physical centers
and flipper pivots/tips during rotation.

The lower lanes, slings, and flippers are mirrored about the usable playfield
center (x=332.5), excluding the shooter lane. Continuous return dividers separate
each inlane from its outlane; their visible edges and shadows use the same
physical paths. Cyan return and amber outlane markings are flush and take their
positions from the lane definitions.

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

## Shared material treatment

All three bumpers use the approved Phase 2 treatment. A cool softbox above/left drives
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
1×, and 2×. The emission supplies the idle appearance; Phase 4 adds timed cores,
reflections, and the compressed body variant.

The source art is authored Go in `cmd/genassets/bumper_material.go`; the four PNGs
are generated outputs in `assets/images/`. There are no curated raster inputs to
preserve in this phase. Do not hand-edit generated PNGs. If later work introduces
painted sources, store them outside `images/` and `audio/` (for example
`assets/sources/`) and explicitly add the conversion step to the generator.
Freshness verification covers 27 PNGs and seven WAVs.

Phase 3 extends the palette to rails, flippers, posts, target faces, lane inserts,
and the plunger. `hardware_material.go` authors the reusable surfaces. Capsule
footprint tests cover the 2× flipper, target, and post sprites; all physical geometry
is unchanged. Static rails use the exact collider paths and radii, with inset
chrome bands, rubber faces, captive brackets, and fasteners on the upper curves.
Slingshot covers have smoked translucent faces, routed traces, chrome mounting
posts, rubber skirts, and down-right shadows. Only the cabinet and drain lip render
after the ball. Flush target sockets remain below a ball after a target drops.

The spring well is recessed. Separate head, rod, and coil sprites keep the head
size and wire thickness fixed as coil spacing compresses; the foot stays inside
the well. The ball uses a separate shadow and a chrome body with directional
reflections. Flipper key highlights remain baked and rotate with the blade.

Material shading is offline; `internal/platform/response.go` adds local light,
short sparks, and mechanical response at runtime. A bumper cap lowers 3 logical
units for 95 ms; impact light fades within 450 ms. Target faces lower across
100 ms and rise across 160 ms while colliders follow game rules immediately.
Bank/jackpot light sequences end within 900 ms, and spark storage is capped at 36.
Drawing never advances effect time; pause freezes it, and a new ball clears it.

The extra printed-markings layer
prevents bumper patches from covering the title and makes those planes independent.
The five full-table textures are 1440×2160 (about 59 MiB decoded RGBA together).
This intentionally trades texture memory for crisp 2× captures while keeping
static detail in five image draws. The complete PNG inventory is about 1.1 MiB.
The shared bumper textures are reused three times, adding six draws over Phase 2;
separating the markings adds one more. No per-pixel shading runs during play. Phase 4 adds five small textures (about
61 KiB combined) and draws target sockets beneath moving faces, a ball shadow,
and nine compact plunger parts. Idle hardware still uses image draws only;
additional light/spark primitives are restricted to active responses.

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
`cmd/genassets/bumper_material.go`, `cmd/genassets/hardware_material.go`, and the
shared table definition. The committed
binaries are direct outputs of that source and carry the same project license as
the rest of this repository. Because generation consumes no external
inputs, the source plus its fixed constants are the complete provenance trail.
