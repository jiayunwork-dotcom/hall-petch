// Command hall-petch is the entry point for the Hall-Petch grain-refinement
// strengthening calculator. It exposes both a command-line interface (compute a
// yield strength, invert for the grain diameter, scan a curve, convert units)
// and a small web server that serves the same computation over HTTP and renders
// the curve in the browser.
//
// Examples:
//
//	go run . sy --example example/mild-steel.json
//	go run . sy --sigma0 50 --ky 0.7 --d 20 --d-unit um
//	go run . inverse --sigma0 50 --ky 0.7 --sigma-y 206.5 --d-unit um
//	go run . line --sigma0 50 --ky 0.7 --d-min 5 --d-max 200 --steps 40 --d-unit um
//	go run . convert --value 20 --from um --to mm
//	go run . serve --addr :8080
package main

import (
	"flag"
	"fmt"
	"os"

	"hall-petch/internal/core"
	"hall-petch/internal/line"
	"hall-petch/internal/scale"
	"hall-petch/internal/server"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "sy":
		cmdSy(os.Args[2:])
	case "inverse":
		cmdInverse(os.Args[2:])
	case "line":
		cmdLine(os.Args[2:])
	case "convert":
		cmdConvert(os.Args[2:])
	case "serve", "-http":
		cmdServe(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n", os.Args[1])
		printUsage()
		os.Exit(2)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `hall-petch: Hall-Petch grain-refinement strengthening calculator

Usage:
  hall-petch sy       --example FILE | --sigma0 F --ky F --d F [--d-unit um|mm|m|nm]
  hall-petch inverse  --sigma0 F --ky F --sigma-y F [--d-unit um|mm|m|nm]
  hall-petch line     --sigma0 F --ky F --d-min F --d-max F [--steps N] [--d-unit um|mm|m|nm]
  hall-petch convert  --value F --from UNIT --to UNIT | --astm G
  hall-petch serve    --addr :8080
`)
}

// loadExampleOrParams reads parameters either from an example file or from the
// explicit flags, returning the parsed Params. It is shared by the CLI
// subcommands that need a material definition.
func loadExampleOrParams(fs *flag.FlagSet, example string, sigma0, ky, d float64, dUnit string) (core.Params, error) {
	if example != "" {
		ex, err := core.LoadExample(example)
		if err != nil {
			return core.Params{}, err
		}
		return ex.ToParams(), nil
	}
	if dUnit == "" {
		dUnit = "um"
	}
	p := core.Params{Sigma0: sigma0, Ky: ky, D: d, DUnit: dUnit}
	if err := p.Validate(); err != nil {
		return core.Params{}, err
	}
	return p, nil
}

func cmdSy(args []string) {
	fs := flag.NewFlagSet("sy", flag.ExitOnError)
	example := fs.String("example", "", "path to an example JSON file")
	sigma0 := fs.Float64("sigma0", 0, "friction stress sigma0 in MPa")
	ky := fs.Float64("ky", 0, "strengthening coefficient ky in MPa*m^(1/2)")
	d := fs.Float64("d", 0, "grain diameter")
	dUnit := fs.String("d-unit", "um", "grain diameter unit (um|mm|m|nm)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	p, err := loadExampleOrParams(fs, *example, *sigma0, *ky, *d, *dUnit)
	if err != nil {
		fail(err)
	}
	res, err := p.Evaluate()
	if err != nil {
		fail(err)
	}
	fmt.Printf("sigma0   = %.4f MPa\n", res.Sigma0)
	fmt.Printf("ky       = %.4f MPa*m^0.5\n", res.Ky)
	fmt.Printf("d        = %s\n", core.FormatDiameter(res.D))
	fmt.Printf("d^(-1/2) = %s\n", core.FormatSqrtInv(res.DSqrtInv))
	fmt.Printf("sigma_y  = %s\n", core.FormatMPa(res.SigmaY, 4))
}

func cmdInverse(args []string) {
	fs := flag.NewFlagSet("inverse", flag.ExitOnError)
	sigma0 := fs.Float64("sigma0", 0, "friction stress sigma0 in MPa")
	ky := fs.Float64("ky", 0, "strengthening coefficient ky (MPa*m^0.5); here the *measured* yield strength role is taken by --sigma-y")
	sigmaY := fs.Float64("sigma-y", 0, "measured yield strength in MPa")
	dUnit := fs.String("d-unit", "um", "unit for the recovered diameter (um|mm|m|nm)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	// Reuse Params.InverseDiameterParams by stashing sigmaY in Ky.
	p := core.Params{Sigma0: *sigma0, Ky: *sigmaY, D: *ky, DUnit: *dUnit}
	d, err := p.InverseDiameterParams()
	if err != nil {
		fail(err)
	}
	fmt.Printf("recovered d = %.6f %s\n", d, *dUnit)
}

func cmdLine(args []string) {
	fs := flag.NewFlagSet("line", flag.ExitOnError)
	sigma0 := fs.Float64("sigma0", 0, "friction stress sigma0 in MPa")
	ky := fs.Float64("ky", 0, "strengthening coefficient ky in MPa*m^(1/2)")
	dMin := fs.Float64("d-min", 0, "smallest grain diameter")
	dMax := fs.Float64("d-max", 0, "largest grain diameter")
	steps := fs.Int("steps", 40, "number of samples")
	dUnit := fs.String("d-unit", "um", "diameter unit (um|mm|m|nm)")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	pts, err := line.ScanLine(line.ScanRequest{
		Sigma0: *sigma0,
		Ky:     *ky,
		DMin:   *dMin,
		DMax:   *dMax,
		Steps:  *steps,
		DUnit:  *dUnit,
	})
	if err != nil {
		fail(err)
	}
	fmt.Printf("%-14s %-16s %s\n", "d("+*dUnit+")", "d^(-1/2)", "sigma_y(MPa)")
	for _, p := range pts {
		dDisp, cerr := scale.ConvertLength(p.D, "m", *dUnit)
		if cerr != nil {
			fail(cerr)
		}
		fmt.Printf("%-14.6g %-16.6g %.4f\n", dDisp, p.DSqrtInv, p.SigmaY)
	}
}

func cmdConvert(args []string) {
	fs := flag.NewFlagSet("convert", flag.ExitOnError)
	value := fs.Float64("value", 0, "value to convert")
	from := fs.String("from", "um", "source unit")
	to := fs.String("to", "mm", "target unit")
	astm := fs.Float64("astm", 0, "ASTM grain-size number to convert to a diameter")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	if *astm != 0 {
		d, err := scale.DiameterFromASTM(*astm)
		if err != nil {
			fail(err)
		}
		fmt.Printf("ASTM G=%.2f -> d = %s\n", *astm, core.FormatDiameter(d))
		return
	}
	out, err := scale.ConvertLength(*value, *from, *to)
	if err != nil {
		fail(err)
	}
	fmt.Printf("%.6g %s = %.6g %s\n", *value, *from, out, *to)
}

func cmdServe(args []string) {
	fs := flag.NewFlagSet("serve", flag.ExitOnError)
	addr := fs.String("addr", ":8080", "listen address")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}
	srv := server.NewServer("web", "example")
	fmt.Printf("hall-petch serving on http://localhost%s\n", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %s\n", core.Message(err))
	os.Exit(1)
}
