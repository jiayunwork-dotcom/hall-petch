# hall-petch

A self-contained calculator for the **Hall–Petch grain-refinement strengthening
law**. Given a friction stress `σ0`, a strengthening coefficient `ky`, and a
grain diameter `d`, it computes the yield strength

```
σy = σ0 + ky · d^(-1/2)
```

and supports the inverse problem (given `σy` recover `d`), scanning the
strength curve over a diameter range, and converting between length units and
ASTM grain-size numbers.

This is a materials-science *kernel*, not a microstructure simulator: it models
the single, well-established relationship between grain size and yield strength
under the Hall–Petch assumption.

## What it computes

- **Forward:** `σy = σ0 + ky·d^(-1/2)`. Yield strength rises as the grain size
  shrinks, because the grain-boundary strengthening term `ky·d^(-1/2)` grows.
- **Inverse:** recover `d = (ky / (σy − σ0))²` given a measured strength.
- **Scan:** sample the `σy(d)` curve between two diameters for plotting.
- **Units:** `d` may be supplied in `um`, `mm`, `m`, or `nm`; the kernel works
  internally in metres so the square-root relationship is dimensionally clean.
- **ASTM:** convert an ASTM grain-size number `G` to an equivalent diameter and
  back via `N = ... ` the standard `d = 1 / (2^(G-1)·√2)` area relation.

## Inputs, outputs, and boundaries

- `σ0` (friction stress) must be ≥ 0; `ky` and `d` must be strictly positive.
- `σy` must exceed `σ0` for the inverse problem; otherwise no physical grain
  size exists.
- All conditions are checked up front; invalid input returns a clear error
  rather than a wrong number.

## Usage

Start the web server and JSON API:

```bash
go run . -http :8080
```

Then open <http://localhost:8080/>. The page can **load the example**
(`example/mild-steel.json`, a mild steel with `σ0 = 50`, `ky = 0.7`,
`d = 20 µm`), compute a single strength, and draw the `σy(d)` curve from the
backend point list.

You can also query the API directly. A reproducible command using the bundled
example:

```bash
curl -s -X POST http://localhost:8080/api/sy \
  -H 'Content-Type: application/json' \
  -d '{"sigma0":50,"ky":0.7,"d":20,"d_unit":"um"}'
```

CLI examples:

```bash
go run . sy       --example example/mild-steel.json
go run . sy       --sigma0 50 --ky 0.7 --d 20 --d-unit um
go run . inverse  --sigma0 50 --ky 0.7 --sigma-y 206.5 --d-unit um
go run . line     --sigma0 50 --ky 0.7 --d-min 5 --d-max 200 --steps 40 --d-unit um
go run . convert  --value 20 --from um --to mm
```

## API

- `POST /api/sy` — `{ "sigma0", "ky", "d", "d_unit" }` → `{ "sigma_y", "d_sqrt_inv" }`
- `POST /api/line` — `{ "sigma0", "ky", "d_min", "d_max", "steps", "d_unit" }` → `{ "points": [{ "d", "d_sqrt_inv", "sigma_y" }] }`

## Build & test

```bash
go build ./...
go test ./...
```
