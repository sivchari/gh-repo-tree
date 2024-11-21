package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

func init() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
}

func main() {
	_main()
}

func _main() {
	var out *string
	out = pflag.StringP("out", "o", "gh-repo-sync.yaml", "output file name")
	var excludes *string
	excludes = pflag.StringP("excludes", "e", "", "exclude paths. the format should be comma separated")

	pflag.Usage = func() {
		fmt.Printf(`
Usage:
  gh repo-tree [flags]

Flags:
%s
Examples:
  gh repo-tree -o out.yaml
`, pflag.CommandLine.FlagUsages())
	}
	pflag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		slog.Error("Error getting working directory", slog.String("error", err.Error()))
		return
	}

	var excludePaths []string
	if *excludes != "" {
		excludePaths = strings.Split(*excludes, ",")
	}

	var filePaths []string
	err = filepath.Walk(wd, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.Name() == ".git" {
			wd, err := os.Getwd()
			if err != nil {
				slog.Error("Error getting working directory", slog.String("error", err.Error()))
				return err
			}
			base := filepath.Base(wd)
			_, after, _ := strings.Cut(path, base)
			filePath := strings.TrimPrefix(strings.TrimSuffix(after, "/.git"), "/")
			for _, excludePath := range excludePaths {
				if strings.Contains(filePath, excludePath) {
					return nil
				}
			}
			for _, fp := range filePaths {
				if strings.Contains(filePath, fp) {
					return nil
				}
			}
			filePaths = append(filePaths, filePath)
		}
		return nil
	})
	if err != nil {
		slog.Error("Error walking directory", slog.String("error", err.Error()))
		return
	}

	var arrays string
	for _, filePath := range filePaths {
		arrays += fmt.Sprintf("  - %s\n", filePath)
	}

	file, err := os.Create(*out)
	if err != nil {
		slog.Error("Error creating file", slog.String("error", err.Error()))
		return
	}
	defer file.Close()

	if _, err := file.WriteString(fmt.Sprintf(template, arrays)); err != nil {
		slog.Error("Error writing file", slog.String("error", err.Error()))
		return
	}

	slog.Info("Success", slog.String("output", *out))

	return
}

var template = `repositories:
%s
`
