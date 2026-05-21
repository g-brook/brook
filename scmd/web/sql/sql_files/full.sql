create table certificate
(
    id          integer
        constraint certificate_pk
            primary key autoincrement,
    name        TEXT not null,
    content     TEXT not null,
    private_key TEXT not null,
    desc        TEXT,
    expire_time TEXT
);

create table info
(
    id    integer
        constraint id_pk
            primary key autoincrement,
    key   TEXT,
    value any
);

create table ip_rules
(
    id          INTEGER
        primary key autoincrement,
    strategy_id INTEGER not null,
    ip          TEXT    not null,
    remark      TEXT,
    created_at  TIMESTAMP default CURRENT_TIMESTAMP
);

create table ip_strategies
(
    id            INTEGER
        primary key autoincrement,
    name          TEXT                not null,
    type          TEXT      default 1 not null,
    allow_private INTEGER   default 1 not null,
    status        INTEGER   default 1 not null,
    created_at    TIMESTAMP default CURRENT_TIMESTAMP,
    updated_at    TIMESTAMP default CURRENT_TIMESTAMP
);

create table proxy_config
(
    idx           INTEGER
        primary key autoincrement,
    name          TEXT    not null,
    tag           TEXT,
    remote_port   INTEGER not null,
    proxy_id      TEXT    not null
        unique,
    protocol      TEXT    not null,
    state         INTEGER not null,
    run_state     integer,
    destination   TEXT,
    ip_strategies INTEGER
);

create table web_logger
(
    id       integer not null
        constraint web_logger_pk
            primary key autoincrement,
    protocol text,
    path     text,
    host     text,
    method   text,
    status   integer,
    proxy_id text,
    http_id  text,
    time     ANY
);

create index web_logger_proxy_id_index
    on web_logger (proxy_id);

create table web_proxy_config
(
    id           integer not null
        constraint web_proxy_config_pk
            primary key autoincrement,
    ref_proxy_id integer
        constraint web_proxy_config_pk_2
            unique,
    cert_file    TEXT,
    key_file     TEXT,
    proxy        TEXT,
    cert_id      integer
);

create table users
(
    id       integer
        constraint users_pk
            primary key autoincrement,
    user_id  text
        constraint users_pk_2
            unique,
    password text,
    icon     text,
    is_admin bool
);


