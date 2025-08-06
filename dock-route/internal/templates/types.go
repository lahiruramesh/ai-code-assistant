package templates

type Template struct {
    Name         string            `yaml:"name"`
    Description  string            `yaml:"description"`
    Dockerfile   string            `yaml:"dockerfile"`
    Port         string            `yaml:"port"`
    MountPath    string            `yaml:"mount_path"`
    Environment  map[string]string `yaml:"environment"`
    BuildArgs    map[string]string `yaml:"build_args"`
    DevCommand   []string          `yaml:"dev_command"`
    ProdCommand  []string          `yaml:"prod_command"`
}
