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

import Message from "@/components/message";
import axios, {AxiosError, type AxiosInstance, type AxiosRequestConfig,} from "axios";

import {DefaultResponse, type ResponseData} from "@/types/response";
import NProgress from 'nprogress';
import '@/assets/nprogress.css'

NProgress.configure({
    showSpinner: false,
})
/**
 * 请求对象.
 */
export class Request {
  instance: AxiosInstance;
  baseConfig: AxiosRequestConfig = {
    baseURL: import.meta.env.VITE_BASE_API,
    timeout: 6000,
    method: "POST",
    headers: { "Content-Type": "application/json;charset=UTF-8" },
  };
  constructor(config: AxiosRequestConfig) {
    this.instance = axios.create(Object.assign(this.baseConfig, config));
    this.instance.interceptors.request.use(
      (config) => {
          NProgress.start();
        const token = localStorage.getItem("token");
        if (token) {
          config.headers["Authorization"] = `${token}`;
        }
        // 添加时间戳防止缓存
        const timestamp = Date.now();
        if (config.method?.toLowerCase() === 'get') {
          // GET 请求添加到 URL 参数
          config.params = {
            ...config.params,
            _t: timestamp
          };
        } else {
          // POST 等其他请求添加到 headers
          config.headers["X-Timestamp"] = timestamp.toString();
          // 也可以添加到请求体中（如果需要）
          if (config.data && typeof config.data === 'object') {
            config.data = {
              ...config.data,
              _timestamp: timestamp
            };
          }
        }
        
        return config;
      },
      (err) => {
          NProgress.done();
        return Promise.reject(err);
      }
    );

    this.instance.interceptors.response.use(
      (rsp) => {
          NProgress.done();
        return rsp;
      },
      (err: AxiosError) => {
          NProgress.done();
        Message.error(err.message);
        return Promise.reject(err);
      }
    );
  }

  /**
   * post请求.
   * @param url url.
   * @param data data.
   * @param showError 是否提示error.
   * @returns rsp.
   */
  public post<T>(
    url: string,
    data?: object,
    showError = true
  ): Promise<ResponseData<T>> {
    return new Promise((resolve) => {
      this.instance
        .post(url, data)
        .then((e) => {
          const data = e.data;
          const rsp = new DefaultResponse<T>(data);
          const errorCode = data.code || data.errorCode;
        
          // 处理未授权错误
          if (errorCode === "NOT_ATH") {
            // 清除本地存储的 token
            localStorage.removeItem("token");
            // 跳转到登录页面
            window.location.href = "/";
            return;
          }
          
          if (errorCode !== "OK" && showError) {
            Message.error(data.message || data.errorMsg || "Error");
          }
          resolve(rsp);
        })
        .catch((_e) => {
          resolve(
            new DefaultResponse<T>({
              code: "local_error",
              message: "",
              data: [] as T,
              success() {
                return false;
              },
            })
          );
        });
    });
  }
}
const Http = new Request({});

export default  Http ;
