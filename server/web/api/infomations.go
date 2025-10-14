package api

type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type BaseInfo struct {
	IsRunning bool   `json:"isRunning"`
	Version   string `json:"version"`
}

type LoginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type QueryServerInfo struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

type ServerInfo struct {
	Name        string `json:"name"`
	Port        string `json:"port"`
	TunnelType  string `json:"tunnelType"`
	TAG         string `json:"tag"`
	Connections int    `json:"connections"`
	Users       int    `json:"users"`
}

type InitInfo struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}
