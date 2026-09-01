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

The artwork uses this authored neon-on-ink palette. Gradients use deterministic
interpolation between these colors and a few documented highlight shades in the
generator source.

| Name | Hex | Use |
| --- | --- | --- |
| Ink | `#050816` | page and deep playfield shadow |
| Panel | `#0A122B` | recessed table components |
| Cyan | `#20E8FF` | primary rails and live circuitry |
| Cyan white | `#CFFDFF` | hot highlights |
| Magenta | `#FF2BAD` | targets, flippers, and secondary traces |
| Violet | `#7E47FF` | relay rings and ambient energy |
| Lime | `#A9FF4F` | reserved award/status accent |
| Amber | `#FFBE37` | reserved warning/bonus accent |
| Steel | `#6F8BAE` | ball and plunger metal |

## Image inventory

All dimensions are exact final PNG pixel dimensions. Sprite canvases other
than the opaque background are transparent.

| File | Dimensions | Purpose |
| --- | ---: | --- |
| `images/background.png` | 720×1080 | neon circuit-board playfield |
| `images/logo.png` | 640×200 | Neon Relay circuit wordmark |
| `images/favicon.png` | 64×64 | compact relay-mark icon |
| `images/ball.png` | 64×64 | shaded steel pinball |
| `images/flipper.png` | 180×64 | pivoted neon flipper |
| `images/bumper.png` | 128×128 | pop-bumper cap and coil light |
| `images/post.png` | 48×48 | illuminated table post |
| `images/target.png` | 64×96 | drop target face |
| `images/lane-light.png` | 48×96 | four-step lane indicator |
| `images/plunger.png` | 56×180 | spring plunger assembly |
| `images/glow.png` | 192×192 | soft cyan-violet additive glow |
| `images/particle.png` | 32×32 | eight-point impact spark |

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

The design, drawing instructions, bitmap glyphs, procedural layout, and audio
synthesis recipes were authored for Neon Relay in `cmd/genassets/main.go`. The
committed binaries are direct outputs of that source and carry the same project
license as the rest of this repository. Because generation consumes no external
inputs, the source plus its fixed constants are the complete provenance trail.
