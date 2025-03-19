package file

import (
	"fmt"
	"os"

	"github.com/eniac111/plumbops/internal/types"
)

// FileModule is our main struct for the module.
type FileModule struct{}

// Run implements the module interface by reading parameters
// and performing the requested file operation.
func (fm FileModule) Run(task types.TaskDefinition) types.ModuleResult {
	res := types.ModuleResult{
		TaskName: task.Name,
		Module:   task.Module,
		Changed:  false,
		Failed:   false,
		Msg:      "",
	}

	// 1. Gather parameters
	path, _ := task.Params["path"].(string)
	state, _ := task.Params["state"].(string)
	src, _ := task.Params["src"].(string)
	dest, _ := task.Params["dest"].(string)
	owner, _ := task.Params["owner"].(string)
	group, _ := task.Params["group"].(string)
	modeStr, _ := task.Params["mode"].(string)
	recurse, _ := task.Params["recurse"].(bool)
	modTimeParam, _ := task.Params["modification_time"].(string)
	accTimeParam, _ := task.Params["access_time"].(string)

	if state == "" {
		state = "file"
	}

	if path == "" && (state == "file" || state == "touch" || state == "directory" || state == "absent") {
		return failResult(res, "Missing 'path' parameter")
	}

	if (state == "link" || state == "hard") && (dest == "" || src == "") {
		return failResult(res, "For link/hard link state, both 'src' and 'dest' are required")
	}
	if (state == "link" || state == "hard") && path != "" && dest == "" {
		dest = path
	}

	// 2. Dispatch by state
	switch state {
	case "file":
		changed, err := ensureFile(path, false)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = changed
		res.Msg = fmt.Sprintf("File '%s' created", path)

	case "touch":
		changed, err := ensureFile(path, true)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = changed
		res.Msg = fmt.Sprintf("File '%s' touched", path)

	case "directory":
		changed, err := ensureDirectory(path)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = changed
		res.Msg = fmt.Sprintf("Directory '%s' created", path)

	case "absent":
		existed, err := removePath(path)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = existed
		res.Msg = fmt.Sprintf("Removed '%s'", path)

	case "link":
		changed, err := ensureSymlink(src, dest)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = changed
		res.Msg = fmt.Sprintf("Symlink created: %s -> %s", dest, src)

	case "hard":
		changed, err := ensureHardLink(src, dest)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = changed
		res.Msg = fmt.Sprintf("Hard link created: %s -> %s", dest, src)

	default:
		return failResult(res, fmt.Sprintf("Unknown state '%s'", state))
	}

	// 3. If not absent, set ownership, permissions, times, recursion if needed
	if state != "absent" && !res.Failed {
		fileChanged, err := setFileAttributes(path, owner, group, modeStr, recurse, modTimeParam, accTimeParam)
		if err != nil {
			return failResult(res, err.Error())
		}
		res.Changed = res.Changed || fileChanged
	}

	return res
}

// ---------------------------------------------------------
//  Helper Functions
// ---------------------------------------------------------

func ensureFile(path string, forceTouch bool) (bool, error) {
	info, err := os.Lstat(path)
	if os.IsNotExist(err) {
		f, createErr := os.Create(path)
		if createErr != nil {
			return false, createErr
		}
		_ = f.Close()
		return true, nil
	} else if err != nil {
		return false, err
	}

	if info.IsDir() {
		return false, fmt.Errorf("'%s' exists but is a directory", path)
	}
	return false, nil
}

func ensureDirectory(path string) (bool, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return false, err
		}
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}

func removePath(path string) (bool, error) {
	_, err := os.Lstat(path)
	if os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	}

	err = os.RemoveAll(path)
	if err != nil {
		return true, err
	}
	return true, nil
}

func ensureSymlink(src, dest string) (bool, error) {
	_, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return os.Symlink(src, dest) == nil, nil
	}
	return false, err
}

func ensureHardLink(src, dest string) (bool, error) {
	_, err := os.Lstat(dest)
	if os.IsNotExist(err) {
		return os.Link(src, dest) == nil, nil
	}
	return false, err
}

func setFileAttributes(path, owner, group, modeStr string, recurse bool, modTimeParam, accTimeParam string) (bool, error) {
	// This function would set ownership, permissions, and timestamps.
	// Placeholder for future implementation.
	return false, nil
}

// failResult is a helper function to set Failed = true with a given message.
func failResult(res types.ModuleResult, msg string) types.ModuleResult {
	res.Failed = true
	res.Msg = msg
	return res
}
