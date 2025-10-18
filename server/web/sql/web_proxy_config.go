package sql

type WebProxyConfig struct {
	Id         string `db:"id" json:"id"`
	RefProxyId string `db:"ref_proxy_id" json:"refProxyId"`
	CertFile   string `db:"cert_file" json:"certFile"`
	KeyFile    string `db:"key_file" json:"keyFile"`
	Proxy      string `db:"proxy" json:"proxy"`
}

func AddWebProxyConfig(p WebProxyConfig) error {
	err := Exec(`
				INSERT INTO web_proxy_config("ref_proxy_id","cert_file","key_file","proxy")
				VALUES (?, ?, ?, ?);
			`, p.RefProxyId, p.CertFile, p.KeyFile, p.Proxy)
	return err
}

func UpdateWebProxyConfig(p WebProxyConfig) error {
	return Exec(`
				UPDATE web_proxy_config SET "cert_file"=?, "key_file"=?, "proxy"=? WHERE "ref_proxy_id"=?;
			`, p.CertFile, p.KeyFile, p.Proxy, p.RefProxyId)
}

func GetWebProxyConfig(refProxyId string) *WebProxyConfig {
	res, err := Query("select id,ref_proxy_id,cert_file,key_file,proxy from web_proxy_config where ref_proxy_id=?", refProxyId)
	if err != nil {
		return nil
	}
	defer res.Close()
	for res.rows.Next() {
		var p WebProxyConfig
		if err := res.rows.Scan(&p.Id, &p.RefProxyId, &p.CertFile, &p.KeyFile, &p.Proxy); err != nil {
			return nil
		}
		return &p
	}
	return nil
}
