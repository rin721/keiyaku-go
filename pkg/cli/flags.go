package cli

import urfavecli "github.com/urfave/cli/v2"

// StringFlagSpec 描述字符串 flag。
type StringFlagSpec struct {
	Name        FlagName
	Aliases     []string
	Usage       string
	Default     string
	EnvVars     []EnvName
	Required    bool
	Destination *string
}

// IntFlagSpec 描述整数 flag。
type IntFlagSpec struct {
	Name        FlagName
	Aliases     []string
	Usage       string
	Default     int
	EnvVars     []EnvName
	Required    bool
	Destination *int
}

// BoolFlagSpec 描述布尔 flag。
type BoolFlagSpec struct {
	Name        FlagName
	Aliases     []string
	Usage       string
	EnvVars     []EnvName
	Destination *bool
}

// StringFlag 创建字符串 flag。
func StringFlag(spec StringFlagSpec) Flag {
	return &urfavecli.StringFlag{
		Name:        spec.Name.String(),
		Aliases:     spec.Aliases,
		Usage:       spec.Usage,
		Value:       spec.Default,
		EnvVars:     envNames(spec.EnvVars),
		Required:    spec.Required,
		Destination: spec.Destination,
	}
}

// IntFlag 创建整数 flag。
func IntFlag(spec IntFlagSpec) Flag {
	return &urfavecli.IntFlag{
		Name:        spec.Name.String(),
		Aliases:     spec.Aliases,
		Usage:       spec.Usage,
		Value:       spec.Default,
		EnvVars:     envNames(spec.EnvVars),
		Required:    spec.Required,
		Destination: spec.Destination,
	}
}

// BoolFlag 创建布尔 flag。
func BoolFlag(spec BoolFlagSpec) Flag {
	return &urfavecli.BoolFlag{
		Name:        spec.Name.String(),
		Aliases:     spec.Aliases,
		Usage:       spec.Usage,
		EnvVars:     envNames(spec.EnvVars),
		Destination: spec.Destination,
	}
}

func envNames(names []EnvName) []string {
	if len(names) == 0 {
		return nil
	}
	values := make([]string, 0, len(names))
	for _, name := range names {
		values = append(values, name.String())
	}
	return values
}
