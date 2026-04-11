
CREATE TABLE IF NOT EXISTS certificate(
    id          integer
        constraint certificate_pk
            primary key autoincrement,
    name        TEXT not null,
    content     TEXT not null,
    private_key TEXT not null,
    desc        TEXT,
    expire_time TEXT
);

alter table web_proxy_config
    add cert_id integer;