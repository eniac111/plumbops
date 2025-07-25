package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Playbook describes the list of tasks in a playbook file.
type Playbook struct {
	Tasks []Task `yaml:"tasks"`
}

// Task represents a single playbook task.
type Task struct {
	Name   string                 `yaml:"name"`
	Module string                 `yaml:"module"`
	Params map[string]interface{} `yaml:"params"`
}

// manifestEntry records metadata about a built task binary.
type manifestEntry struct {
	Sha256  string `json:"sha256"`
	Source  string `json:"source"`
	Module  string `json:"module"`
	BuiltAt string `json:"builtAt"`
}

func slugify(s string) string {
	s = strings.ToLower(s)
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b.WriteRune(r)
			prevDash = false
		} else {
			if !prevDash {
				b.WriteRune('-')
				prevDash = true
			}
		}
	}
	res := b.String()
	res = strings.Trim(res, "-")
	return res
}

func buildCmd(args []string) error {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	playbookPath := fs.String("playbook", "playbook.yaml", "path to playbook")
	outDir := fs.String("out", "./dist", "output directory")
	target := fs.String("target", runtime.GOOS+"/"+runtime.GOARCH, "GOOS/GOARCH")
	force := fs.Bool("force", false, "force rebuild")
	fs.Parse(args)

	pbData, err := os.ReadFile(*playbookPath)
	if err != nil {
		return fmt.Errorf("failed to read playbook: %w", err)
	}

	var pb Playbook
	if err := yaml.Unmarshal(pbData, &pb); err != nil {
		return fmt.Errorf("failed to parse playbook: %w", err)
	}

	parts := strings.SplitN(*target, "/", 2)
	goos := parts[0]
	goarch := runtime.GOARCH
	if len(parts) > 1 {
		goarch = parts[1]
	}

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		return fmt.Errorf("failed to create out dir: %w", err)
	}

	manifestPath := filepath.Join(*outDir, "build-manifest.json")
	manifest := map[string]manifestEntry{}
	if f, err := os.Open(manifestPath); err == nil {
		defer f.Close()
		_ = json.NewDecoder(f).Decode(&manifest)
	}

	for i, t := range pb.Tasks {
		taskID := slugify(fmt.Sprintf("%d-%s-%s", i, t.Module, t.Name))
		taskKey := fmt.Sprintf("%s-%s-%s", taskID, goos, goarch)
		binPath := filepath.Join(*outDir, taskKey)
		entry, ok := manifest[taskKey]
		if ok && !*force {
			if f, err := os.Open(binPath); err == nil {
				h := sha256.New()
				if _, err := io.Copy(h, f); err == nil {
					sha := hex.EncodeToString(h.Sum(nil))
					if sha == entry.Sha256 {
						fmt.Printf("Skipping %s (unchanged)\n", taskKey)
						f.Close()
						continue
					}
				}
				f.Close()
			}
		}

		dir := filepath.Join(*outDir, taskID)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "mkdir %s: %v\n", dir, err)
			continue
		}

		paramsJSON, _ := json.Marshal(t.Params)
		os.WriteFile(filepath.Join(dir, "params.json"), paramsJSON, 0o644)

		mainSrc := fmt.Sprintf(`package main

import (
    "encoding/json"
    "%s/internal/modules/%s"
)

//go:embed params.json
var raw []byte

func main() {
    var p map[string]string
    _ = json.Unmarshal(raw, &p)
    if err := module.Run(p); err != nil {
        panic(err)
    }
}
`, "github.com/eniac111/plumbops", t.Module)
		os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainSrc), 0o644)

		cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", binPath)
		cmd.Env = append(os.Environ(), "GOOS="+goos, "GOARCH="+goarch)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "build %s: %v\n", taskKey, err)
			continue
		}

		f, err := os.Open(binPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "open built binary: %v\n", err)
			continue
		}
		h := sha256.New()
		_, _ = io.Copy(h, f)
		sha := hex.EncodeToString(h.Sum(nil))
		f.Close()

		manifest[taskKey] = manifestEntry{
			Sha256:  sha,
			Source:  *playbookPath,
			Module:  t.Module,
			BuiltAt: time.Now().Format(time.RFC3339),
		}
		fmt.Printf("Built %s\n", taskKey)
	}

	mf, err := os.Create(manifestPath)
	if err != nil {
		return fmt.Errorf("write manifest: %w", err)
	}
	defer mf.Close()
	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	if err := enc.Encode(manifest); err != nil {
		return fmt.Errorf("encode manifest: %w", err)
	}
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: plumbops <command> [options]")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := buildCmd(os.Args[2:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	default:
		fmt.Fprintln(os.Stderr, "unknown command:", os.Args[1])
		os.Exit(1)
	}
}
