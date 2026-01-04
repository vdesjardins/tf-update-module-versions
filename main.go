package main

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/vdesjardins/terraform-module-versions/mod"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: terraform-mod-versions <path-to-terraform-project>")
		return
	}
	root := os.Args[1]

	mods, err := mod.FindModules(root)
	if err != nil {
		fmt.Println("error finding modules:", err)
		return
	}

	modulesFound := make(map[string]map[string]bool)
	for _, m := range mods {
		if _, ok := modulesFound[m.Source]; !ok {
			modulesFound[m.Source] = make(map[string]bool)
		}
		modulesFound[m.Source][m.Version] = true
	}

	for src, vers := range modulesFound {
		fmt.Printf("Module: %s\n", src)
		if len(vers) > 0 {
			var versionList []string
			for v := range vers {
				versionList = append(versionList, v)
			}
			sort.Strings(versionList)
			fmt.Printf("  Current versions: %s\n", strings.Join(versionList, ", "))
		}
		module, err := mod.UpstreamModule(src)
		if err != nil {
			fmt.Printf("  Error fetching upstream: %v\n", err)
			continue
		}
		if len(module.Versions) == 0 {
			fmt.Println("  No versions found.")
			continue
		}
		fmt.Println("  Available upstream versions:")
		for _, v := range module.Versions {
			providers := []string{}
			for _, pv := range v.Root.Providers {
				providers = append(providers, fmt.Sprintf("%s (%s)", pv.Name, pv.Version))
			}
			sort.Strings(providers)
			publicationDate := "unknown date"
			if v.RegistryModuleInfo != nil {
				publicationDate = v.RegistryModuleInfo.PublishedAt
			}
			fmt.Printf("    %s - %s [%s]\n", v.Version, strings.Join(providers, ", "), publicationDate)
		}
	}
}
