package main

import (
	//"fmt"
	//"go/build"
	"os"
	//"path/filepath"
	//"strings"
	//"testing"

	"code.google.com/p/go.exp/ssa"
	"code.google.com/p/go.exp/ssa/interp"
	"flag"
	"log"
)

// interop "package main; func execute(args map[string]interface{}) { println(args[\"a\"]);}
//package interp

// import (
// 	"fmt"
// 	"go/ast"
// 	"go/parser"
// 	"go/token"
// 	"os"

// 	"code.google.com/p/go.exp/ssa"
// 	"runtime"
// )

// func ParseString(fset *token.FileSet, src string) (parsed []*ast.File, err error) {
// 	var f *ast.File
// 	f, err = parser.ParseFile(fset, "", src, parser.DeclarationErrors)
// 	if nil == err {
// 		parsed = append(parsed, f)
// 		return
// 	}

// 	stmt := `package main

// func Execute(args interface{}) (ok bool, res map[string]interface{}) {
// ` + src + `
//   return
// }`

// 	f, err = parser.ParseFile(fset, "", stmt, parser.DeclarationErrors)
// 	if nil == err {
// 		parsed = append(parsed, f)
// 		return
// 	}
// 	fmt.Println(stmt)
// 	return
// }

// func CreateScript(mainpkg *ssa.Package, mode Mode, filename string, args []string) (exitCode int, ir interface{}, fn interface{}) {
// 	i := &interpreter{
// 		prog:    mainpkg.Prog,
// 		globals: make(map[ssa.Value]*value),
// 		mode:    mode,
// 	}
// 	initReflect(i)

// 	for importPath, pkg := range i.prog.Packages {
// 		// Initialize global storage.
// 		for _, m := range pkg.Members {
// 			switch v := m.(type) {
// 			case *ssa.Global:
// 				cell := zero(indirectType(v.Type()))
// 				i.globals[v] = &cell
// 			}
// 		}

// 		// Ad-hoc initialization for magic system variables.
// 		switch importPath {
// 		case "syscall":
// 			var envs []value
// 			for _, s := range os.Environ() {
// 				envs = append(envs, s)
// 			}
// 			envs = append(envs, "GOSSAINTERP=1")

// 			if runtime.GOOS != "windows" {
// 				setGlobal(i, pkg, "envs", envs)
// 			}

// 		case "runtime":
// 			// TODO(gri): expose go/types.sizeof so we can
// 			// avoid this fragile magic number;
// 			// unsafe.Sizeof(memStats) won't work since gc
// 			// and go/types have different sizeof
// 			// functions.
// 			setGlobal(i, pkg, "sizeof_C_MStats", uintptr(3696))

// 		case "os":
// 			Args := []value{filename}
// 			for _, s := range args {
// 				Args = append(Args, s)
// 			}
// 			setGlobal(i, pkg, "Args", Args)
// 		}
// 	}

// 	// Top-level error handler.
// 	exitCode = 2
// 	defer func() {
// 		if exitCode != 2 || i.mode&DisableRecover != 0 {
// 			return
// 		}
// 		switch p := recover().(type) {
// 		case exitPanic:
// 			exitCode = int(p)
// 			return
// 		case targetPanic:
// 			fmt.Fprintln(os.Stderr, "panic:", toString(p.v))
// 		case runtime.Error:
// 			fmt.Fprintln(os.Stderr, "panic:", p.Error())
// 		case string:
// 			fmt.Fprintln(os.Stderr, "panic:", p)
// 		default:
// 			fmt.Fprintf(os.Stderr, "panic: unexpected type: %T\n", p)
// 		}
// 	}()

// 	// Run!
// 	call(i, nil, token.NoPos, mainpkg.Init, nil)
// 	if mainFn := mainpkg.Func("Execute"); mainFn != nil {
// 		ir = i
// 		fn = mainFn
// 		exitCode = 0
// 	} else {
// 		fmt.Fprintln(os.Stderr, "No Execute function.")
// 		exitCode = 1
// 	}
// 	return
// }

// func Call(ir interface{}, fn interface{}, args map[string]interface{}) (exitCode int) {
// 	i := ir.(*interpreter)

// 	// Top-level error handler.
// 	exitCode = 2
// 	defer func() {
// 		if exitCode != 2 || i.mode&DisableRecover != 0 {
// 			return
// 		}
// 		switch p := recover().(type) {
// 		case exitPanic:
// 			exitCode = int(p)
// 			return
// 		case targetPanic:
// 			fmt.Fprintln(os.Stderr, "panic:", toString(p.v))
// 		case runtime.Error:
// 			fmt.Fprintln(os.Stderr, "panic:", p.Error())
// 		case string:
// 			fmt.Fprintln(os.Stderr, "panic:", p)
// 		default:
// 			fmt.Fprintf(os.Stderr, "panic: unexpected type: %T\n", p)
// 		}
// 	}()

// 	values := map[value]value{}
// 	for k, v := range args {
// 		values[k] = v
// 	}
// 	//values := args

// 	call(i, nil, token.NoPos, fn.(value), array{values})
// 	exitCode = 0
// 	return
// }

func run(input string) (ret bool) {

	b := ssa.NewBuilder(ssa.SanityCheckFunctions, ssa.GorootLoader, nil)
	files, err := interp.ParseString(b.Prog.Files, input)
	if err != nil {
		log.Printf("ParseString(%s) failed: %s\n", input, err.Error())
		return false
	}

	// // Print a helpful hint if we don't make it to the end.
	// var hint string
	// defer func() {
	// 	if hint != "" {
	// 		fmt.Println("FAIL")
	// 		fmt.Println(hint)
	// 	} else {
	// 		fmt.Println("PASS")
	// 	}
	// }()

	mainpkg, err := b.CreatePackage("main", files)
	if err != nil {
		log.Printf("ssa.Builder.CreatePackage(%s) failed: %s\n", input, err.Error())
		return false
	}

	b.BuildAllPackages()
	b = nil // discard Builder

	exitCode, ir, fn := interp.CreateScript(mainpkg, 0, "script", []string{})
	if exitCode != 0 {
		log.Printf("interp.Interpret(%s) exited with code %d, want zero\n", input, exitCode)
		return false
	}

	if exitCode = interp.Call(ir, fn, map[string]interface{}{"a": 3}); exitCode != 0 {
		log.Printf("interp.Call() exited with code %d, want zero\n", exitCode)
		return false
	}
	if exitCode = interp.Call(ir, fn, map[string]interface{}{"a": 4}); exitCode != 0 {
		log.Printf("interp.Call() exited with code %d, want zero\n", exitCode)
		return false
	}

	ret = true
	return true
}

const slash = string(os.PathSeparator)

func main() {
	flag.Parse()
	run(flag.Arg(0))
}
