package shell

import (
	"bytes"
	"os/exec"

	"github.com/eniac111/plumbops/internal/types"
)

type ShellModule struct{}

func (sm ShellModule) Run(task types.TaskDefinition) types.ModuleResult {
	res := types.ModuleResult{
		TaskName: task.Name,
		Module:   task.Module,
		Changed:  false, // We'll set true if we consider it "changed"
		Failed:   false,
		Msg:      "",
	}

	cmdString, ok := task.Params["cmd"].(string)
	if !ok || cmdString == "" {
		res.Failed = true
		res.Msg = "Missing 'cmd' parameter for shell module"
		return res
	}

	// Execute the shell command
	cmd := exec.Command("sh", "-c", cmdString) // On Windows you'd do "cmd /C"
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	if err != nil {
		res.Failed = true
		res.Msg = "Command failed: " + errBuf.String()
		return res
	}

	// If it ran successfully, let's consider this a "change" for demonstration
	res.Changed = true
	res.Msg = "Command output: " + outBuf.String()
	return res
}
