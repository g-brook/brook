package sql

type ProxyConfig struct {
	Idx        int64  `db:"idx" json:"id"`
	Name       string `db:"name" json:"name"`
	Tag        string `db:"tag" json:"tag"`
	RemotePort int    `db:"remote_port" json:"remotePort"`
	ProxyID    string `db:"proxy_id" json:"proxyId"`
	Protocol   string `db:"protocol" json:"protocol"`
	State      int    `db:"state" json:"state"`
	RunState   int    `db:"run_state"`
	IsRunning  bool   `json:"isRunning"`
	Runtime    string `json:"defin"`
}

func AddProxyConfig(p ProxyConfig) error {
	err := Exec(`
            INSERT INTO proxy_config(name, tag, remote_port, proxy_id, protocol,state,run_state)
            VALUES (?, ?, ?, ?, ?,?);
        `, p.Name, p.Tag, p.RemotePort, p.ProxyID, p.Protocol, p.State)
	return err
}

func DelProxyConfig(id int64) error {
	err := Exec("DELETE FROM proxy_config WHERE idx = ?", id)
	return err
}

func GetAllProxyConfig() []*ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config where state = 1")
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err != nil {
			return nil
		}
		list = append(list, &p)
	}
	return list
}

func GetAllProxyConfigByProxyId(proxyId string) *ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config where state = 1 and proxy_id = ?", proxyId)
	if err != nil {
		return nil
	}
	defer res.Close()

	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err == nil {
			return &p
		}
	}
	return nil
}

func QueryProxyConfig() []*ProxyConfig {
	res, err := Query("select idx,name, tag, remote_port, proxy_id, protocol,state,run_state from proxy_config")
	if err != nil {
		return nil
	}
	defer res.Close()

	var list []*ProxyConfig
	for res.rows.Next() {
		var p ProxyConfig
		if err := res.rows.Scan(&p.Idx, &p.Name, &p.Tag, &p.RemotePort, &p.ProxyID, &p.Protocol, &p.State, &p.RunState); err != nil {
			return nil
		}
		list = append(list, &p)
	}
	return list
}
