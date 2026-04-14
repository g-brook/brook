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

alter table proxy_config
    add ip_strategies integer;

CREATE TABLE IF NOT EXISTS ip_strategies (
                                             id INTEGER PRIMARY KEY AUTOINCREMENT,
                                             name TEXT NOT NULL,                    -- 策略名称
                                             type TEXT NOT NULL DEFAULT 1,        -- WL=白名单 BL=黑名单 IL=仅内网
                                             bind_handler TEXT NOT NULL UNIQUE,      -- 绑定你的插件/Handler 名
                                             allow_private INTEGER NOT NULL DEFAULT 1, -- 允许内网IP 1=允许 0=禁止
                                             status INTEGER NOT NULL DEFAULT 1,      -- 1=启用 0=禁用
                                             created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                                             updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ip_rules (
                                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                                        strategy_id INTEGER NOT NULL,          -- 关联策略ID
                                        ip TEXT NOT NULL,                      -- IP 或 CIDR
                                        remark TEXT,
                                        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);