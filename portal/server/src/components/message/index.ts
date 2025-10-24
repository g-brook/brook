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

// src/services/Message.ts
import {type ToastPosition, useToast} from "./useToast";

type MessageFunction = (
  message: string,
  position?: ToastPosition,
  duration?: number
) => void;

class MessageService {
  private toast = useToast();

  success: MessageFunction = (
    message,
    position = "top-right",
    duration = 3000
  ) => {
    this.toast.addToast({ message, type: "success", position, duration });
  };

  error: MessageFunction = (
    message,
    position = "top-right",
    duration = 3000
  ) => {
    this.toast.addToast({ message, type: "error", position, duration });
  };

  warning: MessageFunction = (
    message,
    position = "top-right",
    duration = 3000
  ) => {
    this.toast.addToast({ message, type: "warning", position, duration });
  };

  info: MessageFunction = (
    message,
    position = "top-right",
    duration = 3000
  ) => {
    this.toast.addToast({ message, type: "info", position, duration });
  };
}

const Message = new MessageService();
export default Message;
