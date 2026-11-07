// Package skillinstall installs bundled agent skills without replacing
// unrelated or locally modified files.
package skillinstall

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Config struct {
	Home      string
	Name      string
	Content   []byte
	Resources map[string][]byte
}

type Result struct {
	Path    string
	Updated bool
}

func Install(cfg Config) (Result, error) {
	if cfg.Home == "" || cfg.Name == "" || len(cfg.Content) == 0 {
		return Result{}, errors.New("skill installation requires home, name, and content")
	}
	if filepath.Base(cfg.Name) != cfg.Name || cfg.Name == "." || cfg.Name == ".." || strings.ContainsAny(cfg.Name, `/\\`) {
		return Result{}, errors.New("invalid skill name")
	}
	dir := filepath.Join(cfg.Home, ".agents", "skills", cfg.Name)
	path := filepath.Join(dir, "SKILL.md")
	result := Result{Path: path}
	statePath := filepath.Join(dir, ".sidefx-sha256")
	resourcesPath := filepath.Join(dir, ".sidefx-resources.json")
	// Validate every new boundary before writing any package content. Reject
	// symlinks even when their target happens to contain the expected bytes.
	files := map[string][]byte{"SKILL.md": cfg.Content}
	for name, content := range cfg.Resources {
		if !resourceName(name) || len(content) == 0 {
			return result, fmt.Errorf("invalid skill resource: %q", name)
		}
		files[name] = content
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for i, name := range names {
		for _, other := range names[i+1:] {
			if strings.HasPrefix(other, name+"/") {
				return result, fmt.Errorf("resource file/directory collision: %s", name)
			}
		}
	}
	for _, name := range append(append([]string{}, names...), ".sidefx-sha256", ".sidefx-resources.json") {
		if err := checkPath(cfg.Home, filepath.Join(dir, filepath.FromSlash(name))); err != nil {
			return result, err
		}
	}
	managed := map[string]string{}
	data, err := os.ReadFile(resourcesPath)
	if err == nil {
		if json.Unmarshal(data, &managed) != nil || managed == nil {
			return result, errors.New("invalid skill resource state")
		}
		for name, hash := range managed {
			if !resourceName(name) || !validDigest(hash) {
				return result, errors.New("invalid skill resource state entry")
			}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, fmt.Errorf("read resource state: %w", err)
	}
	changed := map[string]bool{}
	for _, name := range names {
		target := filepath.Join(dir, filepath.FromSlash(name))
		current, err := os.ReadFile(target)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return result, fmt.Errorf("read existing resource: %w", err)
		}
		if err == nil && string(current) != string(files[name]) {
			owned := managed[name] == digest(current)
			if name == "SKILL.md" {
				var stateErr error
				owned, stateErr = managedCopy(statePath, current)
				if stateErr != nil {
					return result, stateErr
				}
			}
			if !owned {
				return result, fmt.Errorf("refusing to overwrite existing unmanaged or modified skill resource: %s", target)
			}
		}
		changed[name] = err != nil || string(current) != string(files[name])
	}
	// Known conflicts are now ruled out. I/O failure/crash during the following
	// writes can still leave a partial installation; no crash-atomicity is claimed.
	for _, name := range names {
		if !changed[name] {
			continue
		}
		target := filepath.Join(dir, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return result, fmt.Errorf("create resource directory: %w", err)
		}
		if err := os.WriteFile(target, files[name], 0644); err != nil {
			return result, fmt.Errorf("write skill resource: %w", err)
		}
		result.Updated = true
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return result, err
	}
	for name, content := range cfg.Resources {
		managed[name] = digest(content)
	}
	if len(managed) > 0 {
		data, err := json.Marshal(managed)
		if err != nil {
			return result, err
		}
		if err = os.WriteFile(resourcesPath, append(data, '\n'), 0600); err != nil {
			return result, fmt.Errorf("record resource state: %w", err)
		}
	}
	if err := os.WriteFile(statePath, []byte(digest(cfg.Content)+"\n"), 0600); err != nil {
		return result, fmt.Errorf("record installed skill: %w", err)
	}
	return result, nil
}

func resourceName(name string) bool {
	if name == "" || strings.Contains(name, "\\") || !filepath.IsLocal(name) || filepath.ToSlash(filepath.Clean(name)) != name {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "." || part == ".." || part == "" || strings.HasPrefix(part, ".sidefx-") {
			return false
		}
	}
	return name != "SKILL.md"
}
func validDigest(hash string) bool {
	if len(hash) != 64 {
		return false
	}
	for _, c := range hash {
		if !strings.ContainsRune("0123456789abcdef", c) {
			return false
		}
	}
	return true
}

// checkPath checks ancestors inside the package's Home boundary, including
// .agents/skills. A user-supplied Home itself may be a platform-managed alias.
func checkPath(home, target string) error {
	rel, err := filepath.Rel(home, target)
	if err != nil || !filepath.IsLocal(rel) {
		return errors.New("skill path escapes home")
	}
	current := home
	parts := strings.Split(rel, string(filepath.Separator))
	for i, part := range parts {
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("refusing symlink in skill package: %s", current)
		}
		if i < len(parts)-1 && !info.IsDir() {
			return fmt.Errorf("skill parent is not a directory: %s", current)
		}
		if i == len(parts)-1 && !info.Mode().IsRegular() {
			return fmt.Errorf("skill file is not regular: %s", current)
		}
	}
	return nil
}

func managedCopy(statePath string, current []byte) (bool, error) {
	state, err := os.ReadFile(statePath)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read skill installation state: %w", err)
	}
	return strings.TrimSpace(string(state)) == digest(current), nil
}

func digest(content []byte) string {
	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}
