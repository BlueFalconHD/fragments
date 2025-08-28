package main

import (
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/yosssi/gohtml"
)

func RecursivelyFindPages(op string, cache *FragmentCache) map[string]*Fragment {
	pageMap := make(map[string]*Fragment)

	err := filepath.WalkDir(op, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".frag" {
			return nil
		}

		rel, rerr := filepath.Rel(op, path)
		if rerr != nil {
			log.Error("Error determining relative path", "error", rerr)
			return nil
		}
		rel = filepath.ToSlash(rel)
		fragmentName := strings.TrimSuffix(rel, ".frag")

		f := GetFragmentFromName(fragmentName, PAGE, cache)
		pageMap[fragmentName] = f
		return nil
	})
	if err != nil {
		log.Error("Error walking the path", "path", op, "error", err)
	}

	return pageMap
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, os.ModePerm)
		}
		return copyFile(path, target)
	})
}

func build(siteConfigPath string) {
	cfg, err := GetConfiguration(siteConfigPath)
	if err != nil {
		log.Error("Failed to read configuration", "path", siteConfigPath, "error", err)
		return
	}

	fcache := NewFragmentCache(cfg)

	pageDir := filepath.Join(cfg.SiteRoot, cfg.PagePath)
	pageMap := RecursivelyFindPages(pageDir, fcache)

	for _, v := range pageMap {
		_ = v.Evaluate()
	}

	buildDir := filepath.Join(cfg.SiteRoot, cfg.BuildPath)
	if err := os.MkdirAll(buildDir, os.ModePerm); err != nil {
		log.Error("Failed to create build dir", "dir", buildDir, "error", err)
	}

	includeDir := filepath.Join(cfg.SiteRoot, cfg.IncludePath)
	if info, err := os.Stat(includeDir); err == nil && info.IsDir() {
		if err := copyDir(includeDir, buildDir); err != nil {
			log.Error("Failed to copy include directory", "error", err)
		}
	} else {
		log.Debug("Include directory not found or not a directory", "path", includeDir)
	}

	for k, v := range pageMap {
		log.Info("Building page", "name", k)
		res := v.Evaluate()
		res = gohtml.Format(res)

		dest := filepath.Join(buildDir, k+".html")
		if err := os.MkdirAll(filepath.Dir(dest), os.ModePerm); err != nil {
			log.Error("Error creating directories", "dir", filepath.Dir(dest), "error", err)
		}
		file, err := os.Create(dest)
		if err != nil {
			log.Error("Error creating file", "file", dest, "error", err)
			continue
		}

		if _, err := file.Write([]byte(res)); err != nil {
			log.Error("Error writing to file", "file", dest, "error", err)
		}

		file.Close()
		log.Info("Page built", "name", k, "out", dest)
	}
}

func printUsage() {
	fmt.Println(`fragments - Composable static site generator

Usage:
  fragments init [dir]
  fragments build [-c|--config path/to/config.yml]
  fragments help

Commands:
  init    Create a new project skeleton.
  build   Build the site into the configured build directory.

Examples:
  fragments init mysite
  fragments build -c mysite/config.yml`)
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "help", "-h", "--help":
		printUsage()
		return

	case "init":
		fs := flag.NewFlagSet("init", flag.ExitOnError)
		_ = fs.Parse(os.Args[2:])
		dest := "."
		if args := fs.Args(); len(args) > 0 {
			dest = args[0]
		}

		if err := setupFiles(dest); err != nil {
			log.Error("Failed to initialize project", "dir", dest, "error", err)
			os.Exit(1)
		}
		abs, _ := filepath.Abs(dest)
		log.Info("Project initialized", "dir", abs)
		log.Info("Next steps: run build", "example", fmt.Sprintf("fragments build -c %s", filepath.Join(abs, "config.yml")))
		return

	case "build":
		fs := flag.NewFlagSet("build", flag.ExitOnError)
		cfgPathLong := fs.String("config", "config.yml", "Path to site config (YAML)")
		cfgPathShort := fs.String("c", "", "Path to site config (YAML) [shorthand]")
		_ = fs.Parse(os.Args[2:])

		cfgPath := *cfgPathLong
		if *cfgPathShort != "" {
			cfgPath = *cfgPathShort
		}

		build(cfgPath)
		return

	default:
		fmt.Println("Unknown command:", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}
