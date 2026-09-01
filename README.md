# Neon Relay

Neon Relay is an original portrait-oriented 2D pinball game written in Go. It
uses [`prototype/draw`](https://github.com/gonutz/prototype) for browser input,
rendering, and audio, with a deterministic fixed-step physics simulation behind
it.

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
just fmt          # apply Go and Markdown formatting
just lint         # run golangci-lint for the browser target
just test         # run native simulation and tooling tests
just assets-check # verify generated assets are reproducible
just web          # build and verify dist/
just check        # run the complete CI quality gate
```

All artwork and sound effects are generated specifically for this project. See
[`assets/README.md`](assets/README.md) for provenance and regeneration details.
