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
raised flippers, pause, and a scoring rally into `output/preview/`. This advances the real
simulation and calls the production renderer through a software drawing
surface. It uses Node.js and Go's WebAssembly runtime; it does not open a
browser or require OpenGL, SDL, or a display server.

For another output size:

```sh
./scripts/render-preview.sh -width 360 -height 540 -out output/preview-small
```

Each run also saves `bumper-comparison.png`, cropped directly from the rally
frame: the Phase 2 upper-left bumper on the left, the original upper-right bumper
on the right. The crops retain the capture's pixel scale. Inspect the material
study at 2× with:

```sh
./scripts/render-preview.sh -width 1440 -height 2160 -out output/preview-phase2-2x
```

The capture surface uses Go Mono for text and software image filtering, so
font rasterization and antialiasing may differ slightly from the display backend.

To inspect table geometry during layout work, use engine capture modes:

```sh
./scripts/render-preview.sh -colliders -out output/preview-colliders
./scripts/render-preview.sh -blockout -out output/preview-blockout
```

The overlay shows solid collision surfaces in cyan, sensors in amber, flippers
in magenta, and the ball in white. The blockout uses flat geometry to check
composition without textures. Sprite bounds and anchors live in
`internal/table/artwork.go`; static guide paths come from `internal/table/table.go`.
