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
engine captures at 360×540, 720×1080, and 1440×2160. Native graphics checks remain dependent
on unavailable OpenGL/X11 development headers in this environment. Completed
infrastructure does not mean the visual target has been reached.

## What still separates the rendering from the reference

| Area | Current rendering | Target in the prototype |
| --- | --- | --- |
| Composition | New upper arch, framed lanes, and title above the lower bumper | Built-up upper arch; title between upper pair and lower bumper |
| Construction | Fitted guides and slings, paired return/drain lanes, cabinet lip, and apron | Curved metal guides, raised assemblies, substantial cabinet and apron |
| Materials | Shared chrome/rubber/glass treatment across mechanisms; highlights are baked | Beveled chrome, dark rubber, smoked plastic, textured graphite |
| Light | Contact shadows, local reflected light, and timed hardware responses | Selective LED bloom, reflected color, contact shadows and dark recesses |
| Detail | Fitted target sockets, routing pads, fasteners, panel seams, and grain | Routed circuitry around fitted housings, brackets, fasteners and inserts |
| Presentation | Plain score panels and overlay text | Integrated instrument displays and a consistent technical type system |

The priority is composition and physical construction, followed by materials
and lighting. Adding more circuit lines alone will not close this gap.

## Phase 1 — establish the new layout and render structure

- [x] Block out a stronger upper arch with three defined rollover channels;
  place the title between the upper bumper pair and a lower central bumper.
  Enlarge target housings and give the lower slingshots, return lanes, and drain
  apron the proportions of assembled pinball hardware.
- [x] Keep the existing feature set and logical resolution initially. Change
  `internal/table` geometry where needed, checking launch clearance, shot paths,
  flipper reach, and drain access rather than copying impossible concept geometry.
- [x] Define shared component bounds, sprite pivots, and draw order: playfield,
  shadows, mechanisms, moving parts, foreground covers, emission, and HUD.
  Visible contact edges must follow colliders; decorative overhangs need explicit
  clearance and occlusion rules.
- [x] Save a baseline and matching engine captures for comparison with the concept.

**Exit:** the untextured table already has the reference's visual hierarchy;
collision overlays and scripted launches confirm that the new layout is playable.

**Completed:** curved upper guides and three physical channels; larger bumpers
and target housings; title above the lower bumper; taller slingshots and revised
return guides. Shared sprite anchors and separate shadow/hardware/foreground
layers are in place. Swept-flipper clearance also caught and removed a pre-existing
right guide overlap.

**Evidence:** engine captures in `output/preview-phase1-before/` (baseline),
`output/preview-phase1/`, `output/preview-phase1-colliders/`, and
`output/preview-phase1-blockout/`; half-size captures in `output/preview-phase1-small/`.
These are local, ignored outputs. Reproduce with `just render-preview` or the
`-colliders` / `-blockout` options documented in README. Rollover passage, full and
weak launches, flipper clearance/anchors, drain routing, and reviewed trajectory
checks pass. The scripted rally scores and drops targets. Material depth remains
Phase 2 work.

## Phase 2 — prove the material treatment on one assembly

- [x] Finish one bumper and its surrounding playfield patch first: layered base,
  cylindrical sidewall, rubber ring, inset relay cap, screws, segmented LEDs,
  directional highlights, contact shadow, and reflected cyan/pink light.
- [x] Establish one lighting direction and a small material palette. Use baked
  shading, transparent sprites, shadow masks, and glow overlays freely; a full
  3D renderer is not required.
- [x] Author at sufficient resolution for a crisp 2× capture. Check alpha edges,
  downscaling, physical footprint, and pivot metadata in the production renderer.
- [x] Keep procedural generation where useful; allow original painted or
  AI-assisted raster sources when they improve quality. Separate source art from
  generated outputs so `cmd/genassets` cannot overwrite curated work. Record
  provenance and update asset inventory/freshness checks for the chosen pipeline.

**Exit:** a bumper crop at playing size reads as raised metal, rubber, and glass
with convincing light interaction. Settle this treatment before expanding it.

**Completed:** upper-left bumper with a layered chrome base, molded rubber skirt,
reflected cylinder sidewall, recessed smoked-glass relay cap, captive screws,
segmented LEDs, and a textured/routed playfield patch. Four transparent 2× layers
share the unchanged physical anchor and contact radius. Lighting uses a cool
upper-left key, down-right shadow, and selective cyan/pink emission. Other bumpers
retain the previous finish as controls until Phase 3.

**Evidence:** `output/preview-phase2/`, `output/preview-phase2-small/`,
`output/preview-phase2-2x/`, and `output/preview-phase2-colliders/`. Each contains
`bumper-comparison.png`: the engine's left study and right original at native
capture scale. Footprint, transparent padding, alpha filtering, asset freshness,
WASM tests/lint/build, and distribution checks pass. Generation remains entirely
reproducible from `cmd/genassets/bumper_material.go`; inventory and source/output
rules are in `assets/README.md`. Impact animation remains Phase 4.

## Phase 3 — rebuild the table as fitted hardware

- [x] Apply the proven treatment to all bumpers, the four-target bank, lane
  inserts, posts, flipper hubs/blades, and a recessed spring/plunger assembly.
- [x] Replace uniform rail strokes with curved guide assemblies, chrome edges,
  supports, and rubber contact faces. Derive matching collision segments from
  shared paths wherever the shape changes ball travel.
- [x] Give slingshots layered translucent covers, mounting posts, rubber skirts,
  and shadows; add a substantial cabinet lip, upper display housing, and drain apron.
- [x] Add purposeful PCB routing, pads, panel seams, restrained wear, and fine
  surface grain around the mechanisms. Preserve clear ball travel areas.
- [x] Split the currently monolithic background where foreground covers or
  changing components require separate layers; keep static detail baked for speed.

**Exit:** the full unlit table looks mechanically assembled, with consistent
scale, depth, joins, and contact edges. No ball or moving flipper clips through
visible hardware unexpectedly.

**Completed:** shared bumper treatment throughout; shaded target faces and flush
sockets, lane inserts, posts, flipper blades/hubs, and a recessed plunger well.
Chrome guide bands and captive brackets follow the existing collider paths,
including the curved upper arch. Slings have smoked covers, routed traces,
mounting posts, rubber skirts, and shadows. Cabinet/display bezels and apron
panels now have layered edges, fasteners, and restrained machining detail.

Printed markings are separate from the playfield so bumper patches cannot cover
the title. Static layers and mechanism sprites now ship at 2×, with shared anchors
and premultiplied alpha filtering. Physics geometry is unchanged. Source art and
texture costs are documented in `assets/README.md`; event-driven lighting remains
Phase 4, and display typography remains Phase 5.

**Evidence:** `output/preview-phase3/before-after.png` compares Phase 2 (left) with
Phase 3 (right), using matching engine rally frames. Final captures are in
`output/preview-phase3/`, `output/preview-phase3-small/`,
`output/preview-phase3-2x/`, and `output/preview-phase3-colliders/`; 2× detail crops
cover the upper assembly, target bank, and lower assembly. The capture tool now
accepts `-frame` for individual states. Opaque footprint/anchor checks, complete
WASM tests, final native asset freshness checks, lint, formatting, production
build, and distribution verification pass. The scripted rally remains 6,300
points with unchanged ball positions and clear flipper sweeps.

### Lower-table physics correction

- [x] Center the flippers and mirror both lower assemblies at x=332.5, between
  the left wall and shooter divider, moving the flipper pair 27.5 units left.
- [x] Give each side a continuous return guide separating the inlane from the
  outlane. Inlanes roll onto their adjacent flipper; outlanes bypass it and reach
  the centered drain. Remove obstructing return posts and damp guide rebound.
- [x] Refit slings, apron panels, shadows, and printed arrows to the new geometry.
  Cyan inserts mark returns; amber inserts mark outlanes. These are flush markings.
- [x] Verify continuous ball routes with 2 units of radial clearance beyond the
  14-unit ball radius, plus 108 entry-position/velocity simulations. Full flipper
  sweeps, dropped-ball drain routing, and launch regression checks pass.

**Evidence:** direct engine frames in `output/preview-lanes/` and raised-flipper
collision overlays in `output/preview-lanes-colliders/`. Core/native and WASM tests,
asset freshness, lint, and distribution verification cover the correction.

## Phase 4 — integrate light and mechanical response

- [x] Use narrow bright LED cores, soft local bloom, and tinted reflections on
  nearby metal. Keep graphite dark and reserve strong emission for useful signals.
- [x] Add separate idle, lit, and hit appearances: bumper compression/flash,
  slingshot kick, target drop/reset, lane activation, and plunger travel.
- [x] Add a grounded ball shadow and chrome reflections; maintain ball visibility
  over both dark panels and illuminated mechanisms.
- [x] Drive short light sequences and restrained electrical sparks from gameplay
  events. Avoid stars, cosmic rings, and effects that obscure shots.
- [x] Extend the capture harness to retain renderer state, consume events, and
  advance effect time. `PreviewRenderer` now advances independently of drawing;
  the app and capture harness share the same response implementation.

**Exit:** engine frame sequences show readable idle → impact → recovery states,
with consistent shadows/reflections and no lingering flashes or excessive bloom.

**Completed:** a compressed bumper cap, brief diffuser flashes and reflected
light, slingshot kicker highlights, animated target drop/rise, lane activation,
and independently moving plunger head/rod/coils. The chrome ball has a separate
contact shadow and picks up nearby impact colors. Small electrical sparks replace
star sprites and expanding rings; effects pause, expire, and reset with the game.
The corrected lower-lane collision geometry is unchanged.

**Evidence:** `output/preview-phase4/` contains real-event sequences and cropped
comparison sheets for bumper, sling, lane, bank, and plunger, plus the standard attract/ready/play/flipper/pause/rally captures.
Half-size sequences are in `output/preview-phase4-small/`; the bumper is also
captured at 2× in `output/preview-phase4-2x/`. Fixtures seed actual contacts and
park the ball after impact to isolate recovery; no synthetic events are used.
Pause/cleanup, target collider timing/reset, repeatable draws, cached alpha
compositing, material footprint, asset freshness, core/WASM tests, lint, and
build/distribution checks pass. Typography remains Phase 5.

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
