# Neon Relay

Neon Relay is an original portrait-oriented 2D pinball game written in Go. It
uses [`prototype/draw`](https://github.com/gonutz/prototype) for browser input,
rendering, and audio, with a deterministic fixed-step physics simulation behind
it.

[Try online](https://cwbudde.github.io/go-pinball-2d/) — the browser version is
deployed automatically by CI to GitHub Pages.

## Play locally

Install Go 1.27, [`just`](https://just.systems/), and Python 3, then run:

```sh
just run-web
```

Open <http://localhost:8080>. The build is written to the ignored `dist/`
directory and is the same artifact deployed by GitHub Pages.

To build and launch a native window, run:

```sh
just run-native
```

The Linux version uses GLFW/OpenGL and requires a C compiler and the X11/OpenGL
development libraries. On Debian/Ubuntu these are available as `build-essential`, `pkg-config`,
`libx11-dev`, `libxrandr-dev`, `libgl1-mesa-dev`, `libxcursor-dev`, `libxinerama-dev`,
`libxxf86vm-dev`, and `libxi-dev`. Python is only needed for serving the browser version.
Use `just build` to compile without launching the game.

## Controls

| Control             | Action                               |
| ------------------- | ------------------------------------ |
| Enter               | Start or restart a game              |
| A or Left Arrow     | Left flipper                         |
| D or Right Arrow    | Right flipper                        |
| Space or Down Arrow | Hold to charge and release to launch |
| P                   | Pause or resume                      |
| F                   | Toggle fullscreen                    |

Each game has three balls. Complete the drop-target bank to increase the bonus
multiplier, and combine lit lanes and targets to score a jackpot. The high score
is retained in browser storage.

Beside each flipper, the cyan-marked inlane returns the ball to the paddle; the
amber-marked outlane leads to the drain. The lower assembly is centered within
the main playfield, excluding the launch lane on the right.

## Development

The common repository tasks are exposed through `just`:

```sh
just fmt          # apply Go formatting
just lint         # run golangci-lint for native and browser targets
just test         # run native simulation and tooling tests
just assets-check # verify generated assets are reproducible
just web          # build and verify dist/
just check        # run the complete CI quality gate
```

All artwork and sound effects are generated specifically for this project. See
[`assets/README.md`](assets/README.md) for provenance and regeneration details.

## Render engine screenshots

Run `just render-preview` to write PNGs for attract, launch-ready, playing,
raised flippers, pause, a scoring rally, ball lost, and game over into `output/preview/`. This advances the real
simulation and calls the production renderer through a software drawing
surface. It uses Node.js and Go's WebAssembly runtime; it does not open a
browser or require OpenGL, SDL, or a display server.

For another output size:

```sh
./scripts/render-preview.sh -width 360 -height 540 -out output/preview-small
```

Each run also saves `bumper-comparison.png`, cropped directly from the rally
frame: the upper-left and upper-right bumpers in their actual surroundings.
The crops retain the capture's pixel scale. Inspect the material treatment at 2× with:

```sh
./scripts/render-preview.sh -width 1440 -height 2160 -out output/preview-phase3-2x
```

Use `-frame rally` (or `attract`, `uncharged`, `charge`, `ready`, `playing`,
`flippers`, `paused`, `lost`, `gameover`)
to save one state while still running the scripted simulation checks. This is
useful for expensive double-size captures.

Capture timed mechanical and lighting responses with:

```sh
./scripts/render-preview.sh -sequence all -out output/preview-phase4
./scripts/render-preview.sh -sequence bumper -width 1440 -height 2160 -out output/preview-phase4-2x
```

Choose `bumper`, `sling`, `lane`, `bank`, `plunger`, or `all`. Each sequence saves
full engine frames and a labeled crop sheet. The fixtures position a real ball
before contact, require the expected game event, and then park it away from the
assembly so recovery can be inspected without additional collisions. No effects
are injected. Bumper/sling/lane samples show impact, 50 ms, and 600 ms; the bank
also shows its light sequence and automatic reset. Plunger samples show rest,
half charge, full charge, and release. `-colliders` also works with sequences.

`PreviewRenderer.Advance` accepts the events returned by `Game.Update` once per
update, even between saved frames; `Draw` does not advance animation time.
The software surface caches up to 512 unrotated image/glyph resamples for faster sequences
and antialiases thick response strokes.

Gameplay typography uses the same original glyph atlas on every backend.
Software captures use CatmullRom image filtering; native GLFW uses GPU mipmaps.
Diagnostic text and blockout captions retain backend fonts (Go Mono in captures).

To inspect table geometry during layout work, use engine capture modes:

```sh
./scripts/render-preview.sh -colliders -out output/preview-colliders
./scripts/render-preview.sh -blockout -out output/preview-blockout
```

The overlay shows solid collision surfaces in cyan, sensors in amber, flippers
in magenta, and the ball in white. The blockout uses flat geometry to check
composition without textures. Sprite bounds and anchors live in
`internal/table/artwork.go`; static guide paths come from `internal/table/table.go`.

The score, ball, bonus, and apron status display use an original technical font
atlas shared by native, WASM, and software rendering. The apron shows launch
instructions, charge level, pause/resume, and end-of-ball messages; during play
it shows the current relay objective. These flush displays render beneath the
ball. The capture sequence exercises the drain sensor and all three balls to
reach game over; `charge` and `ready` show partial and full plunger charge.
