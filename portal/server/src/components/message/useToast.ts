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

import {reactive, readonly} from "vue";

export type ToastType = "success" | "error" | "warning" | "info";
export type ToastPosition =
  | "top-left"
  | "top-right"
  | "bottom-left"
  | "bottom-right"
  | "top-center"
  | "bottom-center";

export interface Toast {
  id: number;
  message: string;
  type: ToastType;
  duration: number;
  position: ToastPosition;
  removing?: boolean;
}

const positions: ToastPosition[] = [
  "top-left",
  "top-right",
  "bottom-left",
  "bottom-right",
  "top-center",
  "bottom-center",
];

const state = reactive<{ toasts: Toast[] }>({
  toasts: [],
});

let idCounter = 0;

const addToast = ({
  message,
  type = "info",
  duration = 5000,
  position = "top-right",
}: Omit<Partial<Toast>, "id"> & { message: string }) => {
  if (!positions.includes(position)) position = "top-right";

  const toast: Toast = {
    id: idCounter++,
    message,
    type,
    duration,
    position,
  };

  state.toasts.push(toast);

  setTimeout(() => {
    removeToast(toast.id);
  }, duration);
};

const removeToast = (id: number) => {
  const toast = state.toasts.find((t) => t.id === id);
  if (toast && !toast.removing) {
    // 标记为正在移除，触发离开动画
    toast.removing = true;
    
    // 延迟真正移除，让动画完成
    setTimeout(() => {
      const index = state.toasts.findIndex((t) => t.id === id);
      if (index !== -1) state.toasts.splice(index, 1);
    }, 300); // 与 CSS 动画时长匹配
  }
};

export function useToast() {
  return {
    toasts: readonly(state.toasts),
    addToast,
    removeToast,
  };
}
