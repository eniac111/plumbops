package types

// Inventory holds a list of hosts to manage.
type Inventory struct {
	Hosts []Host `yaml:"hosts"`
}

// Host represents one machine in the inventory.
type Host struct {
	Name     string `yaml:"name"`
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"`
	Port     int    `yaml:"port,omitempty"`
	KeyPath  string `yaml:"key_path,omitempty"` // Optional SSH key path
}

// Playbook holds a list of tasks.
type Playbook struct {
	Tasks []TaskDefinition `yaml:"tasks"`
}

// TaskDefinition describes a single task to run (similar to an Ansible task).
type TaskDefinition struct {
	Name       string                 `json:"name"   yaml:"name"`
	Module     string                 `json:"module" yaml:"module"`
	Params     map[string]interface{} `json:"params" yaml:"params"`
	Become     bool                   `json:"become" yaml:"become"`
	BecomeUser string                 `json:"become_user,omitempty" yaml:"become_user,omitempty"`
}

// ModuleResult is what each module returns.
type ModuleResult struct {
	TaskName string `json:"task_name"`
	Module   string `json:"module"`
	Changed  bool   `json:"changed"`
	Failed   bool   `json:"failed"`
	Msg      string `json:"msg"`
}
