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
import { Response } from "@/types/response";

const getProxyConfigs = <Q>(): Promise<Response<Q>> => {
  return Http.post("/api/getProxyConfigs");
};

const genClientConfig = <Q>(): Promise<Response<Q>> => {
  return Http.post("/api/genClientConfig");
};

const addProxyConfig = <Q>(params: any): Promise<Response<Q>> => {
  return Http.post("/api/addProxyConfigs", params);
};

const delProxyConfig = <Q>(id: number): Promise<Response<Q>> => {
  return Http.post("/api/delProxyConfigs", {
    id: id,
  });
};

const updateProxyConfig = <Q>(params: any): Promise<Response<Q>> => {
  return Http.post("/api/updateProxyConfig",params);
};


const updateProxyState = <Q>(params: any): Promise<Response<Q>> => {
  return Http.post("/api/updateProxyState",params);
};


const addWebConfigs = <Q>(params: any): Promise<Response<Q>> => {
  return Http.post("/api/addWebConfigs", params);
};

const getWebConfigs = <Q>(parmas: any): Promise<Response<Q>> => {
  return Http.post("/api/getWebConfigs", parmas);
};

const functions = {
  genClientConfig,
  getProxyConfigs,
  addProxyConfig,
  delProxyConfig,
  updateProxyConfig,
  updateProxyState,
  addWebConfigs,
  getWebConfigs,
};

export default functions;
