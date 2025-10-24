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

import {type ModalOptions, type ModalSize, useModal} from "./useModal";
import {Component} from "vue";
import useI18n from "../lang/useI18n";
import Confirm from "@/components/modal/Confirm.vue";

const { t } = useI18n();

type ModalFunction = (
  component: Component,
  options?: Partial<Omit<ModalOptions, "component">>
) => string;

class ModalService {
  
  private modal = useModal();

  // 打开模态框
  open: ModalFunction = (component, options = {}) => {
    return this.modal.openModal({
      component,
      ...options,
    });
  };

  // 打开确认对话框
  confirm = (
    options: Partial<Omit<ModalOptions, "component">> & {
      onConfirm?: () => void | Promise<void>;
      onCancel?: () => void;
    } = {}
  ) => {
    return this.modal.openModal({
      component:Confirm,
      props:{
        message: t('confirmations.confirmText'),
      },
      showFooter: true,
      size: "sm",
      title:t("confirmations.confirmTips"),
      ...options,
    });
  };

  // 打开信息对话框
  info = (
    component: Component,
    options: Partial<Omit<ModalOptions, "component">> = {}
  ) => {
    return this.modal.openModal({
      component,
      size: "md",
      closable: true,
      maskClosable: true,
      ...options,
    });
  };

  // 打开全屏模态框
  fullscreen = (
    component: Component,
    options: Partial<Omit<ModalOptions, "component">> = {}
  ) => {
    return this.modal.openModal({
      component,
      size: "full",
      ...options,
    });
  };

  // 关闭指定模态框
  close = (id: string) => {
    this.modal.closeModal(id);
  };

  // 关闭所有模态框
  closeAll = () => {
    this.modal.closeAllModals();
  };

  // 设置加载状态
  setLoading = (id: string, loading: boolean) => {
    this.modal.setModalLoading(id, loading);
  };

  // 更新模态框属性
  updateProps = (id: string, props: Record<string, any>) => {
    this.modal.updateModalProps(id, props);
  };
}

const Modal = new ModalService();
export default Modal;
export { useModal, type ModalOptions, type ModalSize };