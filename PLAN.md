# 2D Pinball Game Implementation Plan

## 1. Define the playable scope

Build a single-player, portrait-oriented pinball table with the working title
"Neon Relay":

- Two controllable flippers
- Flipper sprites preserve their authored aspect ratio while their pivot and tip
  remain aligned with the physics capsule
- Spring-loaded launch plunger and shooter lane
- Three pop bumpers
- Two slingshots
- Consistent three-pixel outlines for table walls, lanes, indicators, and both
  triangular slingshots
- Inlanes, outlanes, rollover lanes, targets, walls, posts, and drain
- Three balls per game
- Score, bonus multiplier, high score, pause, ball-lost, and game-over states
- Keyboard controls:
  - `A`/Left Arrow: left flipper
  - `D`/Right Arrow: right flipper
  - Space/Down Arrow: charge and release plunger
  - Enter: start/restart
  - `P`: pause
  - `F`: fullscreen

Initial scope excludes multiplayer, network services, mobile touch controls, and
complex 3D-style ramps.

## 2. Establish the project structure

Create a Go module using Go 1.27 and pin `github.com/gonutz/prototype` to
`v1.9.2`.

Planned layout:

```text
README.md
main.go
index.html
justfile
.golangci.toml
treefmt.toml
internal/
  game/          game states, scoring, balls, event queue
  physics/       vectors, colliders, solver, flippers
  table/         table geometry and feature definitions
  platform/      prototype/draw input, rendering, and audio adapter
assets/
  images/        generated original PNG assets
  audio/         generated original WAV effects
  README.md      asset provenance and generation notes
cmd/
  genassets/     deterministic custom asset generator
scripts/
  build-web.sh
.github/
  workflows/
    ci.yml
    pages.yml
dist/            ignored build output
```

Keep physics and game rules independent of `prototype/draw`, allowing native
unit tests without desktop graphics dependencies.

Add repository-wide Go quality tooling, following the conventions used by the
nearby `libs/algo-fft` project where applicable:

- A `justfile` is the canonical entry point for formatting, linting, testing,
  asset generation/freshness checks, the WASM build, and the full CI gate.
- Formatting rules cover all tracked Go source with `gofmt` and reject a dirty
  tree after formatting in CI.
- A versioned `.golangci.toml` enables an explicit, maintainable linter set and
  keeps generated assets and browser-only adapters scoped appropriately.
- CI invokes the same `just` recipes developers run locally.

## 3. Implement the simulation core

Use a fixed logical playfield of `720x1080` units and a refresh-rate-independent
physics loop:

- Sample elapsed wall time inside the framework update callback.
- Accumulate time and simulate at `1/240 s` per step.
- Cap accumulated time and substeps after tab suspension to prevent an unstable
  catch-up burst.
- Keep rendering separate from simulation.

Implement:

- Vector and geometry utilities
- Dynamic circular balls
- Static line-segment and circular colliders
- Sensors for lanes, targets, drain, and scoring zones
- Swept-circle collision detection to prevent fast balls tunnelling through thin
  walls
- Restitution, friction, penetration correction, velocity limits, and iterative
  contact resolution
- Rotating capsule-shaped flippers with angular contact velocity
- Plunger charge and release impulse
- Re-arm the plunger when an underpowered launch falls back into the shooter-lane
  cradle, without consuming the ball
- Per-contact cooldowns so persistent overlaps do not score every simulation
  step
- Guards against NaN values and runaway energy

## 4. Build the table and game rules

Represent table geometry as data rather than embedding coordinates throughout
rendering and physics.

Implement the game state machine:

```text
Loading -> Attract -> Ball Ready -> Playing -> Ball Lost
                                      |
                                      v
                                    Paused
                                      |
                                      v
                              Game Over -> Attract
```

Add scoring rules:

- Pop bumper: 100 points
- Slingshot: 50 points
- Rollover lane: 250 points
- Drop target: 500 points
- Completing the target bank raises the bonus multiplier
- Lit lane and target combinations award a jackpot
- End-of-ball bonus is applied before the next ball
- High score persists through browser `localStorage`, with an in-memory fallback
  outside WASM

Game logic should emit discrete events such as `BumperHit`, `TargetDown`,
`BallDrained`, and `GameStarted`. Rendering, sound, particles, and screen shake
consume those events without affecting simulation.

## 5. Create entirely original assets

Create every game-specific asset for this repository; do not download
third-party art, sound effects, or music.

Use `cmd/genassets` to produce deterministic assets from authored shapes and
synthesized waveforms:

- Neon circuit-board playfield background
- Logo and favicon
- Ball, flipper, bumper, post, target, lane-light, and plunger PNGs
- Glow and particle sprites
- Original WAV effects for flippers, bumpers, targets, launch, jackpot, drain,
  and game over

Commit both the generator and its generated outputs. Document the palette,
dimensions, generation command, and provenance in `assets/README.md`. This makes
the custom-asset requirement reproducible and auditable.

## 6. Integrate with `prototype/draw`

Create a thin platform adapter that:

- Starts the game with `draw.RunWindow`
- Converts `draw.Window` keyboard state into game input
- Draws the background, table components, ball, effects, and HUD
- Uses `DrawImageFileTo` for scaled and rotated sprites
- Uses framework primitives as fallbacks while assets load
- Treats `draw.ErrImageLoading` as an expected loading state and exposes real
  load failures visibly
- Plays only WAV effects and avoids queuing attract-mode sounds before the
  browser's first user gesture
- Scales logical playfield coordinates into the current canvas while preserving
  aspect ratio

This matches the framework's documented drawing/input/audio interface and its
asynchronous WASM image behavior. See the
[prototype README](https://github.com/gonutz/prototype),
[Window API](https://github.com/gonutz/prototype/blob/v1.9.2/draw/window.go), and
[WASM backend](https://github.com/gonutz/prototype/blob/v1.9.2/draw/window_wasm.go).

## 7. Create the browser shell and WASM build

Provide a custom `index.html` containing:

- Required `gameCanvas`
- Responsive portrait layout and dark page background
- Loading and actionable error messages
- Control instructions
- WebAssembly streaming-instantiation fallback
- Relative URLs so deployment works under `/pinball-2d/`
- Accessible page title and canvas description

Create `scripts/build-web.sh` to:

1. Generate or verify assets.
2. Run the pinned `drawsm build`.
3. Stage `index.html`, `main.wasm`, `wasm_exec.js`, assets, favicon, and
   `.nojekyll` into `dist/`.
4. Verify that required files are present and non-empty.

Using `drawsm` ensures `wasm_exec.js` comes from the same Go installation that
compiled `main.wasm`; this is how the framework's
[WASM build tool](https://github.com/gonutz/prototype/blob/v1.9.2/cmd/drawsm/main.go)
is designed.

## 8. Add verification

Automated tests:

- Vector and geometry operations
- Swept ball/segment and ball/circle collisions
- Reflection, restitution, and penetration correction
- Flipper contact impulse
- Plunger charge limits
- Sensor cooldown and scoring
- Ball/life/state transitions
- Deterministic scripted simulation with no NaNs or escaped balls
- Asset-generation freshness
- Successful `js/wasm` production build
- Validation that every HTML and game asset reference exists in `dist`
- `just fmt-check` and `just lint` pass with the repository's pinned
  formatting and golangci-lint rules

Manual browser checks:

- Chrome/Chromium, Firefox, and Safari
- 60 Hz, 120 Hz, and 144 Hz displays
- Fresh-cache load from the final Pages URL
- Audio activation after the first key press
- Complete start -> launch -> score -> drain -> game-over -> restart loop
- Underpowered launch -> shooter-lane return -> recharge -> successful relaunch
- Fullscreen, pause/resume, and tab-background recovery

## 9. Deploy through GitHub Pages

Create a Pages workflow triggered by pushes to the default branch and manual
dispatch:

1. Check out the repository.
2. Install Go from `go.mod` with `actions/setup-go`.
3. Run tests and the web build.
4. Configure Pages.
5. Upload only `dist/`.
6. Deploy to the `github-pages` environment.

Use:

- `contents: read`
- `pages: write`
- `id-token: write`
- A Pages concurrency group
- Current stable, pinned action versions
- GitHub Actions as the repository's Pages source

This follows GitHub's
[custom Pages workflow guidance](https://docs.github.com/en/pages/getting-started-with-github-pages/using-custom-workflows-with-github-pages)
and
[official static Pages workflow](https://github.com/actions/starter-workflows/blob/main/pages/static.yml).

## 10. Completion criteria

The implementation is complete when:

- The public GitHub Pages URL loads without console or asset errors.
- A player can finish a full three-ball game using only the documented controls.
- Physics remain materially consistent across common refresh rates.
- Scoring and high-score persistence work.
- All visual and audio assets are original and documented.
- Unit tests and the WASM build pass in CI.
- A push to the default branch reproducibly publishes the game.
