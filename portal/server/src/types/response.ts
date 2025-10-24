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

export interface Response<T> {
  code: string;
  message: string;
  data: T;
  success(): boolean;
}

class DefaultResponse<T> implements Response<T> {
  private _data!: T;
  private _code!: string;
  private _message!: string;

  constructor(response: Response<T>) {
    this._data = response.data;
    this._code = response.code;
    this._message = response.message;
  }

  get data(): T {
    return this._data;
  }

  get code(): string {
    return this._code;
  }
  get message(): string {
    return this._message;
  }

  success(): boolean {
    return this._code === "OK";
  }
}

export { DefaultResponse };
export type ResponseData<T> = Response<T>;
