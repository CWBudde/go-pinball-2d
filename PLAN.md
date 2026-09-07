# Neon Relay — visual development plan

Target: bring the playable table substantially closer to [prototype.png](prototype.png).
Use its composition, material depth, and electronics identity as the reference;
retain an original, readable 2D pinball game. A pixel-perfect clone is unnecessary.

## Completed foundation

- [x] Go / `prototype/draw` game with a 720×1080 logical playfield, fixed-step
  physics, collision handling, flippers, rechargeable plunger, and table features.
- [x] Three-ball game, scoring, lane/target combinations, bonus, persistent high
  score, keyboard controls, loading, pause, drain, and restart states.
- [x] Earlier review remediation: physics/scoring defects, table validation,
  input/storage resilience, simulation coverage, build tooling, and performance.
- [x] Original generated images and audio, asset freshness checks, formatting,
  native/WASM quality recipes, distribution verification, and Pages workflows.
- [x] First Neon Relay art pass: graphite circuits, baked rails and slingshot
  plastics, revised hardware sprites, lit/unlit inserts, typography, and effects.
- [x] `just render-preview`: real engine screenshots without a browser, including
  launch, play, raised flippers, and pause; reference concept saved in the repo.

Recent validation passed for WASM tests/lint/build, core and asset checks, and
engine captures at 720×1080 and 360×540. Native graphics checks remain dependent
on unavailable OpenGL/X11 development headers in this environment. Completed
infrastructure does not mean the visual target has been reached.

## What still separates the rendering from the reference

| Area | Current rendering | Target in the prototype |
| --- | --- | --- |
| Composition | Sparse upper lanes; third bumper above the title | Built-up upper arch; title between upper pair and lower bumper |
| Construction | Thin, angular rails and flat triangular plastics | Curved metal guides, raised assemblies, substantial cabinet and apron |
| Materials | Clean outlines and simple gradients | Beveled chrome, dark rubber, smoked plastic, textured graphite |
| Light | Bright borders with limited surface interaction | Selective LED bloom, reflected color, contact shadows and dark recesses |
| Detail | Isolated traces and small floating targets | Routed circuitry around fitted housings, brackets, fasteners and inserts |
| Presentation | Plain score panels and overlay text | Integrated instrument displays and a consistent technical type system |

The priority is composition and physical construction, followed by materials
and lighting. Adding more circuit lines alone will not close this gap.

## Phase 1 — establish the new layout and render structure

- [ ] Block out a stronger upper arch with three defined rollover channels;
  place the title between the upper bumper pair and a lower central bumper.
  Enlarge target housings and give the lower slingshots, return lanes, and drain
  apron the proportions of assembled pinball hardware.
- [ ] Keep the existing feature set and logical resolution initially. Change
  `internal/table` geometry where needed, checking launch clearance, shot paths,
  flipper reach, and drain access rather than copying impossible concept geometry.
- [ ] Define shared component bounds, sprite pivots, and draw order: playfield,
  shadows, mechanisms, moving parts, foreground covers, emission, and HUD.
  Visible contact edges must follow colliders; decorative overhangs need explicit
  clearance and occlusion rules.
- [ ] Save a baseline and matching engine captures for comparison with the concept.

**Exit:** the untextured table already has the reference's visual hierarchy;
collision overlays and scripted launches confirm that the new layout is playable.

## Phase 2 — prove the material treatment on one assembly

- [ ] Finish one bumper and its surrounding playfield patch first: layered base,
  cylindrical sidewall, rubber ring, inset relay cap, screws, segmented LEDs,
  directional highlights, contact shadow, and reflected cyan/pink light.
- [ ] Establish one lighting direction and a small material palette. Use baked
  shading, transparent sprites, shadow masks, and glow overlays freely; a full
  3D renderer is not required.
- [ ] Author at sufficient resolution for a crisp 2× capture. Check alpha edges,
  downscaling, physical footprint, and pivot metadata in the production renderer.
- [ ] Keep procedural generation where useful; allow original painted or
  AI-assisted raster sources when they improve quality. Separate source art from
  generated outputs so `cmd/genassets` cannot overwrite curated work. Record
  provenance and update asset inventory/freshness checks for the chosen pipeline.

**Exit:** a bumper crop at playing size reads as raised metal, rubber, and glass
with convincing light interaction. Settle this treatment before expanding it.

## Phase 3 — rebuild the table as fitted hardware

- [ ] Apply the proven treatment to all bumpers, the four-target bank, lane
  inserts, posts, flipper hubs/blades, and a recessed spring/plunger assembly.
- [ ] Replace uniform rail strokes with curved guide assemblies, chrome edges,
  supports, and rubber contact faces. Derive matching collision segments from
  shared paths wherever the shape changes ball travel.
- [ ] Give slingshots layered translucent covers, mounting posts, rubber skirts,
  and shadows; add a substantial cabinet lip, upper display housing, and drain apron.
- [ ] Add purposeful PCB routing, pads, panel seams, restrained wear, and fine
  surface grain around the mechanisms. Preserve clear ball travel areas.
- [ ] Split the currently monolithic background where foreground covers or
  changing components require separate layers; keep static detail baked for speed.

**Exit:** the full unlit table looks mechanically assembled, with consistent
scale, depth, joins, and contact edges. No ball or moving flipper clips through
visible hardware unexpectedly.

## Phase 4 — integrate light and mechanical response

- [ ] Use narrow bright LED cores, soft local bloom, and tinted reflections on
  nearby metal. Keep graphite dark and reserve strong emission for useful signals.
- [ ] Add separate idle, lit, and hit appearances: bumper compression/flash,
  slingshot kick, target drop/reset, lane activation, and plunger travel.
- [ ] Add a grounded ball shadow and chrome reflections; maintain ball visibility
  over both dark panels and illuminated mechanisms.
- [ ] Drive short light sequences and restrained electrical sparks from gameplay
  events. Avoid stars, cosmic rings, and effects that obscure shots.
- [ ] Extend the capture harness to retain renderer state, consume events, and
  advance effect time. `RenderFrame` currently creates a fresh renderer per capture,
  so existing stills do not validate transient effects.

**Exit:** engine frame sequences show readable idle → impact → recovery states,
with consistent shadows/reflections and no lingering flashes or excessive bloom.

## Phase 5 — finish displays and visual hierarchy

- [ ] Build inset score/ball displays with technical or segmented numerals;
  integrate bonus, lane, and bank labels into dedicated table areas.
- [ ] Refine the title and relay symbol to match the hardware's finish. Use a
  consistent font atlas or equivalent shared typography for capture and runtime.
- [ ] Design attract, charge, pause, ball-lost, and game-over treatments as part
  of the table presentation. Keep instructions legible without covering active shots.
- [ ] Balance detail and contrast at 720×1080 and 360×540; inspect 1440×2160 for
  blurry assets, seams, halos, and inconsistent shading.

**Exit:** scores, ball, flippers, and objectives remain immediately readable;
labels do not float over mechanisms or collide with the cabinet.

## Phase 6 — compare, tune, and verify

- [ ] Capture matching full-table views and crops of the upper arch, bumper,
  target bank, and lower assembly beside `prototype.png`. Resolve the largest
  remaining composition/material gaps before adding small decorative details.
- [ ] Capture launch-ready, normal play, impacts, completed lanes, dropped/reset
  targets, both flipper extremes, pause, and game over through the actual engine.
- [ ] Recheck full/weak launches, scoring, ball containment, drain, and a complete
  three-ball loop after geometry edits. Update trajectory expectations only after
  reviewing the intended physical changes.
- [ ] Run relevant simulation/platform tests, asset checks, formatting, lint,
  WASM build, and distribution verification. Measure rendering cost and asset
  size against the baseline; cache textures and bake static layers as needed.
- [ ] Update asset provenance and development documentation to describe the
  final pipeline and any remaining backend differences accurately.

**Done when:** the table convincingly shares the prototype's composition,
material richness, mechanical depth, and controlled neon lighting at normal
playing size, while remaining responsive, readable, and fully playable.

All testing and screenshots must use the engine directly, never a browser.
Use the existing software capture path; install easy missing dependencies
without sudo, and stop to ask if sudo is required. Each phase ends with an
engine-rendered comparison before proceeding to finer polish.
