package commons

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
)

func LoadConfig(nm string, flagSet *flag.FlagSet, isOverride bool) error {
	f, e := os.Open(nm)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	var res map[string]interface{}
	e = json.NewDecoder(f).Decode(&res)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}

	actual := map[string]string{}
	fn := func(f *flag.Flag) {
		actual[f.Name] = f.Name
	}
	if nil == flagSet {
		flag.Visit(fn)
	} else {
		flagSet.Visit(fn)
	}

	e = assignFlagSet("", res, flagSet, actual, isOverride)
	if nil != e {
		return fmt.Errorf("load config '%s' failed, %v", nm, e)
	}
	return nil
}

func assignFlagSet(prefix string, res map[string]interface{}, flagSet *flag.FlagSet, actual map[string]string, isOverride bool) error {
	for k, v := range res {
		switch value := v.(type) {
		case map[string]interface{}:
			e := assignFlagSet(combineName(prefix, k), value, flagSet, actual, isOverride)
			if nil != e {
				return e
			}
			continue
		case []interface{}:
		case string:
		case float64:
		case bool:
		case nil:
			continue
		default:
			return fmt.Errorf("unsupported type for %s - %T", combineName(prefix, k), v)
		}
		nm := combineName(prefix, k)

		if !isOverride {
			if _, ok := actual[nm]; ok {
				log.Printf("load flag '%s' from config is skipped.\n", nm)
				continue
			}
		}

		var g *flag.Flag
		if nil == flagSet {
			g = flag.Lookup(nm)
		} else {
			g = flagSet.Lookup(nm)
		}
		if nil == g {
			log.Printf("flag '%s' is not defined.\n", nm)
			continue
		}

		err := g.Value.Set(fmt.Sprint(v))
		if nil != err {
			return err
		}
	}
	return nil
}

func combineName(prefix, nm string) string {
	if "" == prefix {
		return nm
	}
	return prefix + "." + nm
}

func SetFlags(cfg map[string]string, flagSet *flag.FlagSet, isOverride bool) {
	actual := map[string]string{}
	flags := make([]*flag.Flag, 0, 10)
	if nil == flagSet {
		if !isOverride {
			flag.Visit(func(g *flag.Flag) {
				actual[g.Name] = g.Name
			})
		}
		flag.VisitAll(func(g *flag.Flag) {
			if isOverride {
				flags = append(flags, g)
			} else if _, ok := actual[g.Name]; !ok {
				flags = append(flags, g)
			}
		})
		for _, g := range flags {
			if v, ok := cfg[g.Name]; ok {
				flag.Set(g.Name, v)
			}
		}
	} else {
		if !isOverride {
			flagSet.Visit(func(g *flag.Flag) {
				actual[g.Name] = g.Name
			})
		}
		flagSet.VisitAll(func(g *flag.Flag) {
			if isOverride {
				flags = append(flags, g)
			} else if _, ok := actual[g.Name]; !ok {
				flags = append(flags, g)
			}
		})
		for _, g := range flags {
			if v, ok := cfg[g.Name]; ok {
				flagSet.Set(g.Name, v)
			}
		}
	}
}
