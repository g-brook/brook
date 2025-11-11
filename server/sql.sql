/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

CREATE TABLE IF NOT EXISTS proxy_config
(
    idx         INTEGER PRIMARY KEY AUTOINCREMENT,
    name        TEXT        NOT NULL,
    tag         TEXT,
    remote_port INTEGER     NOT NULL,
    proxy_id    TEXT UNIQUE NOT NULL,
    protocol    TEXT        NOT NULL,
    state       INTEGER     NOT NULL
);


CREATE TABLE IF NOT EXISTS proxy_config
(
    idx         INTEGER
        primary key autoincrement,
    name        TEXT    not null,
    tag         TEXT,
    remote_port INTEGER not null
        constraint proxy_config_pk
            unique,
    proxy_id    TEXT    not null
        unique,
    protocol    TEXT    not null,
    state       integer,
    run_state   integer
);

CREATE TABLE IF NOT EXISTS web_logger
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

CREATE TABLE IF NOT EXISTS web_proxy_config
(
    id           integer not null
        constraint web_proxy_config_pk
            primary key autoincrement,
    ref_proxy_id integer
        constraint web_proxy_config_pk_2
            unique,
    cert_file    TEXT,
    key_file     TEXT,
    proxy        TEXT
);

