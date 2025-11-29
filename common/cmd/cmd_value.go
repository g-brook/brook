package cmd

type CliCmdValue struct {
	ConfigPath string

	IsContainer bool
}

type SevCmdValue struct {
	ConfigPath string

	IsContainer bool
}

func NewCliCmdValue() *CliCmdValue {
	return &CliCmdValue{}
}

func NewSevCmdValue() *SevCmdValue {
	return &SevCmdValue{}
}
