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

export class AuthToken {
   public token !: string;
   public createTime!: Date;
}

const getAuthToken = <AuthToken>(): Promise<Response<AuthToken>> => {
  return Http.post("/api/getToken");
};

const generateAuthToken = (): Promise<Response<AuthToken>> => {
  return Http.post("/api/generateToken");
};

const delToken = (): Promise<Response<void>> => {
  return Http.post("/api/delToken");
};

const functions = { getAuthToken ,generateAuthToken,delToken};

export default functions;