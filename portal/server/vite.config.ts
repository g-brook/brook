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

import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  const FE_PORT = parseInt(env.VITE_PORT || "3000");
  const BE_PORT = parseInt(env.VITE_SERVER_PORT || "8000");
  const BASE_API = env.VITE_BASE_API;
  const isDev = mode === "development";
  console.log("isDev", isDev);
  console.log("BASE_API", BASE_API);
  console.log("FE_PORT", FE_PORT);
  console.log("BE_PORT", BE_PORT);
  return {
    plugins: [tailwindcss(), vue()],
    resolve: {
      alias: {
        "@": path.resolve(__dirname, "src"),
      },
      extensions: [".js", ".ts", ".jsx", ".tsx", ".json", ".vue"],
    },
    server: isDev
      ? {
          port: FE_PORT,
          open: true,
          proxy: {
            [BASE_API]: {
              target: `http://127.0.0.1:${BE_PORT}`,
              changeOrigin: true,
              rewrite: (path) => path.replace(/^\/remote/, ""),
            },
          },
        }
      : undefined,
  };
});
