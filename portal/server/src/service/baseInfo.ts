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

import Http from "@/components/request";
import {Response} from "@/types/response";
import {InitInfo} from "@/types/info";

const getBaseInfo = <Q>(): Promise<Response<Q>> => {
  return Http.post("/api/getBaseInfo");
};

const initServer = <Q>(data: InitInfo): Promise<Response<Q>> => {
  return Http.post("/api/initBrookServer",data);
};

const login = <Q>(data: any): Promise<Response<Q>> => {
  return Http.post("/api/login", data);
};

const getServerInfo = <Q>(data: any): Promise<Response<Q>> => {
  return Http.post("/api/getServerInfo", data);
};

const getServerInfoByProxyId = <Q>(data: any): Promise<Response<Q>> => {
    return Http.post("/api/getServerInfoByProxyId", data);
};

const functions = { getBaseInfo, initServer, login, getServerInfo,getServerInfoByProxyId };

export default functions;