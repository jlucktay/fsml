package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	fsmf "github.com/OpenFactorioServerManager/factorio-server-manager/factorio"
)

var savePath string

func init() {
	const usage = "path to Factorio savegame ZIP"
	flag.StringVar(&savePath, "s", "", usage+" (shorthand)")
	flag.StringVar(&savePath, "save-path", "", usage)
}

type modListWrapper struct {
	Mods []modEntry `json:"mods"`
}

type modEntry struct {
	Enabled bool   `json:"enabled"`
	Name    string `json:"name"`
}

func main() {
	flag.Parse()

	if len(savePath) == 0 {
		log.Fatalf("No save path given; see '%s --help'.", os.Args[0])
	}

	datFiles := []string{
		"level-init.dat",
		"level.dat",
	}

	modNames := map[string]struct{}{}

	for _, datFile := range datFiles {
		archiveFile, err := fsmf.OpenArchiveFile(savePath, datFile)
		if err != nil {
			if errors.Is(os.ErrNotExist, err) {
				continue
			}

			fmt.Fprintf(os.Stderr, "fsmf.OpenArchiveFile(%s): %v\n", savePath, err)
			os.Exit(2)
		}

		fsh := fsmf.SaveHeader{}
		if err := fsh.ReadFrom(archiveFile); err != nil {
			fmt.Fprintf(os.Stderr, "fsh.ReadFrom(%s): %v\n", datFile, err)
			archiveFile.Close()
			continue
		}

		archiveFile.Close()

		for _, mod := range fsh.Mods {
			if strings.EqualFold(mod.Name, "base") {
				continue
			}

			if utf8.ValidString(mod.Name) {
				modNames[mod.Name] = struct{}{}
			}
		}
	}

	mapKeys := make([]string, 0, len(modNames))

	for k := range modNames {
		mapKeys = append(mapKeys, k)
	}

	sort.Slice(mapKeys, func(i, j int) bool {
		return strings.ToLower(mapKeys[i]) < strings.ToLower(mapKeys[j])
	})

	modEntries := make([]modEntry, 0, len(modNames)+1)
	modEntries = append(modEntries, modEntry{Name: "base", Enabled: true})

	for _, mapKey := range mapKeys {
		modEntries = append(modEntries, modEntry{Name: mapKey, Enabled: true})
	}

	modWrapper := modListWrapper{Mods: modEntries}

	if err := json.NewEncoder(os.Stdout).Encode(modWrapper); err != nil {
		fmt.Fprintf(os.Stderr, "json.Encode(%s): %v\n", savePath, err)
		os.Exit(4)
	}
}
