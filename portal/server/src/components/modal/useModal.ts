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

import {type Component, reactive, readonly} from "vue";
import useI18n from "../lang/useI18n";

var t = useI18n();
export type ModalSize = "sm" | "md" | "lg" | "xl" | "full" | "auto" | string;

　

export interface ModalOptions {
  id: string;
  component: Component;
  props?: Record<string, any>;
  size?: ModalSize;
  title?: string;
  closable?: boolean;
  maskClosable?: boolean;
  showFooter?: boolean;
  confirmText?: string;
  cancelText?: string;
  onConfirm?: (() => void) | (() => Promise<void>);
  onCancel?: () => void;
  onClose?: () => void;
}

export interface Modal extends ModalOptions {
  id: string;
  visible: boolean;
  loading?: boolean;
}

const state = reactive<{ modals: Modal[] }>({
  modals: [],
});

let idCounter = 0;

const openModal = (options: Omit<ModalOptions, "id"> & { id?: string }) => {
  const modal: Modal = {
    id: options.id || `modal_${idCounter++}`,
    component: options.component,
    props: options.props || {},
    size: options.size || "md",
    title: options.title,
    closable: options.closable !== false,
    maskClosable: options.maskClosable !== false,
    showFooter: options.showFooter || false,
    confirmText: options.confirmText || t.common.confirm.value,
    cancelText: options.cancelText || t.common.cancel.value,
    onConfirm: options.onConfirm,
    onCancel: options.onCancel,
    onClose: options.onClose,
    visible: true,
    loading: false,
  };

  // 如果已存在相同 id 的模态框，先关闭它
  const existingIndex = state.modals.findIndex(m => m.id === modal.id);
  if (existingIndex !== -1) {
    state.modals.splice(existingIndex, 1);
  }

  state.modals.push(modal);
  return modal.id;
};

const closeModal = (id: string) => {
  const modal = state.modals.find(m => m.id === id);
  if (modal) {
    modal.visible = false;
    // 延迟移除，让动画完成
    setTimeout(() => {
      const index = state.modals.findIndex(m => m.id === id);
      if (index !== -1) {
        const modal = state.modals[index];
        modal.onClose?.();
        state.modals.splice(index, 1);
      }
    }, 300);
  }
};

const closeAllModals = () => {
  state.modals.forEach(modal => {
    modal.visible = false;
  });
  
  setTimeout(() => {
    state.modals.forEach(modal => modal.onClose?.());
    state.modals.splice(0);
  }, 300);
};

const setModalLoading = (id: string, loading: boolean) => {
  const modal = state.modals.find(m => m.id === id);
  if (modal) {
    modal.loading = loading;
  }
};

const updateModalProps = (id: string, props: Record<string, any>) => {
  const modal = state.modals.find(m => m.id === id);
  if (modal) {
    modal.props = { ...modal.props, ...props };
  }
};

export function useModal() {
  return {
    modals: readonly(state.modals),
    openModal,
    closeModal,
    closeAllModals,
    setModalLoading,
    updateModalProps,
  };
}