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

Create `scripts/build-web.sh`, committed with executable mode (`0755`) so
`just web` works in a fresh CI checkout, to:

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

## 11. Post-review remediation

A full codebase review (physics, game logic, platform/rendering, build tooling,
and Go/test quality) produced the findings below. They are grouped into four
phases; phase 11.1 comes first because the current test harness cannot detect
most of the other defects.

Legend: `[ ]` open, `[x]` done, `[~]` in progress, `[-]` deliberately skipped.

### 11.1 Phase A - make the safety net real

The flagship integration test never plays the table.
`TestScriptedSimulationStaysFinite` holds the plunger for 12 frames at 1/60 s,
but `Plunger.Hold` runs once per _fixed step_, so the charge reaches only 0.15.
The ball rises 274 px, never leaves the shooter lane, and the run ends with 11
plunger re-arms and a final score of 0 after 20 simulated seconds. Bumpers,
slingshots, targets, rollovers, flippers, the drain sensor, and the guide walls
are therefore untested end to end. Several defects in phase 11.2 survived
precisely because of this.

- [x] Charge the plunger to full (~1.4 s) in the scripted tests, and assert the
      ball actually exits the lane (`Ball.Position.Y < 300`) and that the final
      score is non-zero. Same fix for
      `TestWeakLaunchCanBeRechargedAndRelaunched` and
      `TestScriptedSimulationIsConsistentAcrossRefreshRates`.
- [x] Add `./internal/platform` to the `test` recipe in `justfile:7`; its only
      test file has never executed in CI or locally.
- [x] Add `build` to the `check` recipe so a native compile break cannot land
      green; install `libgl1-mesa-dev` and `xorg-dev` in `ci.yml`.
- [x] `git rm -r internal/table/.go-cache` (two build-cache files are tracked)
      and add an unanchored `.go-cache/` pattern to `.gitignore`, which today
      only anchors `/.gocache/`.
- [x] Remove `go = "1.27"` from `.golangci.toml:6`. The pinned
      golangci-lint v2.13.1 release binary is built with go1.26 and refuses to
      load the config. CI installs the pinned linter from source with Go 1.27;
      release binaries built with Go 1.26 still cannot analyze this Go 1.27
      module even after the explicit setting is removed.

### 11.2 Phase B - correctness defects

- [x] **Open right playfield boundary.** `OuterWalls`
      (`internal/table/table.go:151-158`) contains `wall_outer_left_lower` and
      `wall_outer_left_drain` with no mirror image. The region x in (575, 625),
      y > 915 is unbounded, and the `Drain` box stops at x = 510, so a ball
      there falls through and is caught only by the emergency
      `Y > table.Height+80` fallback - roughly 0.9 s late and 82 px below the
      visible playfield, with drain effects drawn outside the table. Add
      `wall_outer_right_lower` (625,815)->(510,1015) and
      `wall_outer_right_drain` (510,1015)->(510,1050).
- [x] Add a table test that drops a ball from a grid of playfield positions and
      asserts the drain **sensor** - not the fallback - catches every one.
- [x] **Zero-impulse contacts score points.** `RayCircle`
      (`internal/physics/geometry.go:36-44`) returns `TOI=0` whenever the start
      is inside the circle, regardless of travel direction, and the capsule
      overlap branch (`collider.go:66-76`) does the same. The solver correctly
      applies zero impulse but `StepBall` still emits a `Contact`, and
      `game.go:179-188` scores on the collider-ID prefix alone. A ball grazing
      or separating from a bumper is awarded `BumperScore` every 75 ms. Reject
      separating candidates in `earliestCollision` and add an
      `if contact.Impulse <= 0 { continue }` guard in `stepPlaying`.
- [x] **Unchecked flipper indexing.** `game.go:114-124` indexes
      `World.Flippers[0]` and `[1]` in the core step. `New` accepts any
      exported `*table.Definition`, so a definition with fewer than two
      flippers panics. Validate in `New`, and prefer named `left`/`right`
      fields over positional indexing.
- [x] **Flipper impulse saturates `MaxSpeed`.** `RiseSpeed: 18` rad/s over a
      105-unit blade gives a 1890 u/s tip speed, and `combineMaterial`
      (`collider.go:113`) takes `math.Max` of the two restitutions (0.82), for
      an exit speed up to 1.82x the surface speed - measured at 2515 u/s
      against the 2600 cap. Shots near the pivot and near the tip clip to the
      same speed, so aiming is impossible. Lower `RiseSpeed` to about 9-11 and
      combine restitution geometrically. Add a test asserting a resting ball
      leaves the blade below `MaxSpeed`.
- [x] **High score written on every point.** `addScore` (`game.go:320-326`)
      calls `store.SaveHighScore` whenever the record is beaten, which is every
      award once it is - from inside the 240 Hz fixed-step loop. On wasm that
      is a synchronous `localStorage.setItem` per bumper hit. Buffer in memory
      and flush once on the `GameEnded` transition.
- [x] **`bankReset` leaks across balls and games.** It is decremented only in
      `stepPlaying` (`game.go:167-173`) and not reset by `startGame`, so a
      completed target bank stays frozen through the drain and pops back up
      mid-play on the next ball. Reset it in `startGame`/`drainBall`, or
      advance it in `step` for all live states.
- [x] **Flippers stick down on focus loss.** The framework binds keydown/keyup
      on `window` with no blur handler, so alt-tabbing while holding A or D
      holds the flipper engaged for the rest of the session. Dispatch synthetic
      `keyup` events from `index.html` on `blur` and `visibilitychange`.
- [x] **`localStorage` access can kill the game.** `store_wasm.go:15,31` let a
      JS exception (sandboxed iframe, Safari private mode) panic through to
      `main.go`. Wrap both in a recovering helper that returns a zero value.
- [x] Prevent the default action for Space and the arrow keys in `index.html`;
      the framework's own handler covers neither, so charging the plunger
      scrolls the page.
- [x] Fix `truncate` (`internal/platform/render.go:310-315`): it slices bytes
      and so can split a UTF-8 rune, and it panics for `limit < 3`.
- [x] Restore the responsive layout when leaving fullscreen; the framework
      writes inline `720px`/`1080px` styles that outrank the stylesheet, so the
      playfield is cropped rather than scaled on narrow viewports.
- [x] Clear `statusError` (`internal/platform/app.go:22`) - it is sticky, so
      one transient audio failure pins the error banner for the session - and
      stop retrying `SetIcon` every frame after the first failure.

### 11.3 Phase C - design and data model

The table is data-driven in name only. `Bumper.Score`, `Slingshot.Score`,
`Lane.Score`, and `DropTarget.Score` are populated in `table.New` and never
read; the game hardcodes `table.BumperScore` and friends and dispatches on
`strings.HasPrefix(id, "bumper_")`. An ID typo produces a silently inert,
scoreless feature, and no validation exists anywhere.

- [ ] Give colliders and sensors a `Kind` field (or expose a
      `map[string]Feature`), look up the score by ID, and delete the prefix
      matching. This makes the existing `Score` fields load-bearing.
- [ ] Add a table test asserting every collider and sensor ID resolves to a
      known feature kind.
- [ ] Implement or remove the four inlane/outlane sensors
      (`table.go:204-211`): they match no prefix in `stepPlaying`, so they can
      never fire, yet they run an `Overlaps` call at 240 Hz.
- [ ] Resolve `Ball.Mass` (`physics/body.go:14`): it scales only the _reported_
      `Contact.Impulse` and has no effect on the dynamics, while `Guard`
      carefully repairs it. Either divide by it in `ResolveStaticContact` or
      delete it and document the ball as unit-mass.
- [ ] Replace the `g.World.Lines = g.World.Lines[:0]` reach-in
      (`game.go:302-310`) with a `World.SetLines` method or an `Enabled` flag
      on `LineCollider`.
- [ ] Build slingshot colliders from all three triangle edges; the renderer
      draws three but only the upper face is solid, and the lower edge
      overlaps `guide_left_inlane`.
- [ ] Separate the flipper posts from the flipper capsules - `post_left_flipper`
      sits 22.4 units from a pivot with a combined radius of 27, i.e. embedded
      in the flipper body.
- [ ] Delete the unused exported API: `Vec.Cross`, `Segment.Length`,
      `Clock.Reset`, `Clock.Alpha`, `CircleSensor`, `OverlappingSensors`,
      `DefaultSolverIterations`, and `Contact.SurfaceVelocity`/`Flipper`.
- [ ] Emit `BallDrained` before `BonusAwarded` (`game.go:271,276`); consumers
      play the effects in order, so the bonus fanfare currently precedes the
      drain sound.
- [ ] Give `Loading` a timeout fallback so a missing asset cannot strand the
      game in a state with no recovery.

### 11.4 Phase D - tests, tooling, and performance

Coverage is 86.9% across `internal/`, but the assertions are shallower than the
number suggests: `hasEvent` compares only `Event.Kind`, so `ID`, `Points`, and
`At` are never checked anywhere; the refresh-rate test compares the code
against itself and passes if the physics is uniformly wrong; and
`TestTargetBankMultiplierAndLitJackpot` calls `hitTarget` and writes `litLanes`
directly, bypassing sensor dispatch entirely.

Tests:

- [ ] Assert `Event.ID`, `Points`, and `At`, not just `Kind`.
- [ ] Add a golden-trajectory regression test pinning actual ball positions;
      the simulation is deliberately deterministic, so this is nearly free.
- [ ] Cover the scoring paths in `stepPlaying` (46.9% covered): bumper and
      slingshot contacts, rollover lighting, drain by sensor, the
      out-of-bounds recovery, and `bankReset` expiry - through the sensor
      dispatch rather than around it.
- [ ] Cover the defensive guards, which are the least-tested code in the
      repository despite existing solely for bad input: `Ball.Guard` (64.7%),
      `Vec.Normalized` and `ClampLength` (75%), `Flipper.Step` (72.2%),
      `RayCircle` (81.8%), `Plunger.Hold` (66.7%). Table tests with `NaN`,
      `+/-Inf`, zero, and negative values.
- [ ] Test `cmd/verifydist.verify` (0% covered, and it gates deployment) with
      `t.TempDir()` fixtures: missing `index.html`, missing canvas, missing
      `.nojekyll`, zero-byte asset, success.
- [ ] Test `State.String()` (0% covered); it feeds the wasm accessibility
      label.
- [ ] Add the invariant test for the flipper sweep:
      `(MaxSpeed + RiseSpeed*Length)*FixedStep < BallRadius + Flipper.Radius`.
      No swept volume is computed for the rotating flipper, so this margin is
      the only thing preventing tunnelling and nothing currently guards it.
- [ ] Add a `BenchmarkStepBall`; `earliestCollision` is a full linear scan over
      ~40 colliders, up to 10 iterations, 240 times a second, with no
      broadphase.

Tooling and repository:

- [ ] Add a `LICENSE` file. `assets/README.md:86` claims the assets "carry the
      same project license as the rest of this repository", but no license file
      exists, so the reference resolves to nothing.
- [ ] Run golangci-lint for both targets. `justfile:11` runs it only under
      `GOOS=js GOARCH=wasm`, so `status_native.go` and `store_native.go` are
      never linted at all.
- [ ] Decide on Markdown formatting: either install markdownlint and prettier
      in `ci.yml` and drop `--allow-missing-formatter`, or delete the
      `[formatter.markdownlint]` block. Neither tool has ever run in CI.
- [ ] Narrow `treefmt.toml:8` from `assets/**` to `assets/images/**` and
      `assets/audio/**`; the current pattern also excludes `assets/README.md`.
- [ ] Add a `concurrency` group and a branch filter to `ci.yml` (every
      same-repo PR currently runs the gate twice), SHA-pin the third-party
      `extractions/setup-just` action, and cache the installed tool binaries.
- [ ] Gate `pages.yml` on the quality job. It currently runs only `just test`
      and `just web`, so a deploy can publish what CI would have rejected.
- [ ] Delete the dead `required[]` array in `scripts/build-web.sh:27-63`;
      `cmd/verifydist` already performs every one of those checks and derives
      its list from the source instead of a fourth hand-maintained copy.
- [ ] Fix the systematic darkening in `cmd/genassets/main.go:172`: `finish()`
      truncates with `uint8(r/ss)` while `blendPixel` rounds, biasing every
      pixel down. Use `uint8((r + ss/2) / ss)`.
- [ ] Harden asset determinism: `blendPixel` and `mix` have the `x*y + z` shape
      that permits FMA contraction on arm64, and byte-comparing PNGs makes the
      committed assets hostage to `compress/flate` output across Go releases.
      Compare decoded pixels rather than encoded bytes.
- [ ] Make `checkAssets` bidirectional; it never notices orphaned files in
      `assets/` that the generator no longer produces.

Performance:

- [ ] Replace the triple `append` in `render.go:139`, which builds and discards
      a ~19-element slice every frame purely to iterate three others.
- [ ] Cache the canvas handle and the last published values in
      `status_wasm.go`; it crosses the JS boundary about nine times per frame
      to write values that change only on scoring events.
- [ ] Allocate the contact slice lazily in `solver.go:80` rather than
      `make([]Contact, 0, 2)` 240 times a second.
- [ ] Precompute the cooldown key strings; `"contact:"+id` and `"sensor:"+id`
      (`game.go:180,202,281`) allocate thousands of times a second.
- [ ] Handle `devicePixelRatio`: the backing store is pinned at 720x1080 while
      CSS scales the element, so the game is soft on every HiDPI display.
      `newViewport` already letterboxes correctly, so this is a one-line JS
      change.
- [ ] Scale stroke widths and the plunger-bar insets with the viewport
      (`render.go:30,234`); they are fixed pixel constants today, giving
      hairlines in fullscreen and misregistered fills at any scale but 1.
- [ ] Add touch controls, or drop the mobile affordances (`viewport-fit=cover`,
      `env(safe-area-inset-*)`) that promise them. The game is currently
      unplayable on phones and tablets.

## 12. Documentation corrections

Statements in this plan and in the repository that the code does not support:

- [ ] Section 8 claims `just lint` passes; it fails locally for anyone using a
      golangci-lint release binary (see phase 11.1).
- [ ] Section 6 implies formatting covers all tracked sources; the Markdown
      half of `treefmt.toml` has never executed.
- [ ] Section 8's test list implies full verification, but `internal/platform`
      is excluded from the test recipe entirely.
- [ ] `assets/README.md:86` refers to a project license that does not exist.
- [ ] Section 2's layout omits `cmd/verifydist/`, which is implemented.
