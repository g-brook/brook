<!--
  - Copyright ©  sixh sixh@apache.org
  -
  - Licensed under the Apache License, Version 2.0 (the "License");
  - you may not use this file except in compliance with the License.
  - You may obtain a copy of the License at
  -
  -     http://www.apache.org/licenses/LICENSE-2.0
  -
  - Unless required by applicable law or agreed to in writing, software
  - distributed under the License is distributed on an "AS IS" BASIS,
  - WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  - See the License for the specific language governing permissions and
  - limitations under the License.
  -->

<template>
  <div>
    <transition-group name="modal" tag="div">
      <dialog v-for="modal in visibleModals" :key="modal.id" class="modal modal-open overflow-hidden p-0 m-0"
        @click="handleMaskClick(modal)">
        <div :class="getModalClass(modal.size)" class="modal-box  bg-base-100 relative rounded-4xl p-0" @click.stop>
          <!-- 标题栏 -->
          <div class="flex items-center justify-between m-4">
            <!-- 标题 -->
            <h3 v-if="modal.title" class="font-bold text-lg ">
              {{ modal.title }}
            </h3>
            <h3 v-else class="font-bold text-lg">
            </h3>

            <!-- 关闭按钮 -->
            <button v-if="modal.closable"
              class="btn btn-sm btn-circle btn-ghost hover:bg-error/10 hover:text-error transition-colors"
              @click="closeModal(modal.id)" title="关闭">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          <!-- 内容区域 -->
          <div class="modal-content mt-4 mx-4">
            <component :is="modal.component" v-bind="modal.props" :modal-id="modal.id" @close="closeModal(modal.id)"
              @confirm="handleConfirm(modal)" @cancel="handleCancel(modal)"/>
          </div><!-- 底部操作栏 -->
          <div v-if="modal.showFooter" class="modal-action border-t border-base-300 p-4 mt-10">
            <button class="btn btn-ghost" :disabled="modal.loading" @click="handleCancel(modal)">
              {{ modal.cancelText }}
            </button>
            <button class="btn btn-primary" :class="{ loading: modal.loading }" :disabled="modal.loading"
              @click="handleConfirm(modal)">
              {{ modal.confirmText }}
            </button>
          </div>
        </div>
      </dialog>
    </transition-group>
  </div>
</template>

<script lang="ts" setup>
import {computed, onMounted, onUnmounted, watch} from 'vue';
import {type ModalSize, useModal} from './useModal';

const { modals, closeModal, setModalLoading } = useModal();

const visibleModals = computed(() => modals.filter(m => m.visible));

// Lock body and html scroll when modal is visible to prevent background scrolling
let lockCount = 0;
const originalStyle = {
  body: { overflow: '', paddingRight: '' },
  html: { overflow: '' },
};

function getScrollbarWidth() {
  const hasScrollbar = document.documentElement.scrollHeight > document.documentElement.clientHeight;
  if (!hasScrollbar) return 0;
  return window.innerWidth - document.documentElement.clientWidth;
}

function lockBodyScroll() {
  if (lockCount === 0) {
    originalStyle.body.overflow = document.body.style.overflow;
    originalStyle.body.paddingRight = document.body.style.paddingRight;
    originalStyle.html.overflow = document.documentElement.style.overflow;

    const scrollbarWidth = getScrollbarWidth();
    if (scrollbarWidth > 0) {
      document.body.style.paddingRight = `${parseFloat(getComputedStyle(document.body).paddingRight || '0') + scrollbarWidth
        }px`;
    }
    document.body.style.overflow = 'hidden';
    document.documentElement.style.overflow = 'hidden';
  }
  lockCount++;
}

function unlockBodyScroll() {
  requestAnimationFrame(() => {
    lockCount = Math.max(0, lockCount - 1);
    if (lockCount === 0) {
      document.body.style.overflow = originalStyle.body.overflow;
      document.body.style.paddingRight = originalStyle.body.paddingRight;
      document.documentElement.style.overflow = originalStyle.html.overflow;
    }
  });
}

watch(
  visibleModals,
  (newModals, oldModals) => {
    const newLength = newModals.length;
    const oldLength = oldModals?.length ?? 0;
    if (newLength > oldLength) {
      for (let i = 0; i < newLength - oldLength; i++) {
        lockBodyScroll();
      }
    } else if (newLength < oldLength) {
      for (let i = 0; i < oldLength - newLength; i++) {
        unlockBodyScroll();
      }
    }
  },
  { deep: true }
);

// 调试信息
watch(modals, (newModals) => {
}, { deep: true });

watch(visibleModals, (newVisibleModals) => {
}, { deep: true });

// ESC 键关闭模态框
const handleKeydown = (event: KeyboardEvent) => {
  if (event.key === 'Escape' && visibleModals.value.length > 0) {
    const topModal = visibleModals.value[visibleModals.value.length - 1];
    if (topModal.closable) {
      closeModal(topModal.id);
    }
  }
};

onMounted(() => {
  document.addEventListener('keydown', handleKeydown);
});

onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown);
  // Fallback to ensure scroll is unlocked if component unmounts unexpectedly
  while (lockCount > 0) {
    unlockBodyScroll();
  }
});

const getModalClass = (size?: ModalSize,): string => {
  const baseClasses = "max-h-[90vh] overflow-y-auto";

  switch (size) {
    case "sm":
      return `${baseClasses} w-80 max-w-sm`;
    case "md":
      return `${baseClasses} w-96 max-w-md`;
    case "lg":
      return `${baseClasses} w-[32rem] max-w-2xl`;
    case "xl":
      return `${baseClasses} w-[64rem] max-w-4xl`;
    case "full":
      return `${baseClasses} w-[95vw] max-w-none h-[95vh] max-h-none`;
    case "auto":
      return `${baseClasses} w-fit max-w-none`;
    default:
      return `${size}`;
  }
};

const handleMaskClick = (modal: any) => {
  if (modal.maskClosable) {
    closeModal(modal.id);
  }
};

const handleConfirm = async (modal: any) => {
  if (modal.onConfirm) {
    try {
      setModalLoading(modal.id, true);
      const result = await modal.onConfirm();
      if (result) {
        closeModal(modal.id);
      }
    } catch (error) {
      console.error('Modal confirm error:', error);
    } finally {
      setModalLoading(modal.id, false);
    }
  } else {
    closeModal(modal.id);
  }
};

const handleCancel = (modal: any) => {
  if (modal.onCancel) {
    try {
      modal.onCancel();
    } catch (error) {
      console.error('Modal cancel error:', error);
    }
  }
  closeModal(modal.id);
};
</script>

<style scoped>
/* 模态框位置调整 - 向上偏移 */
.modal {
  align-items: flex-start;
  padding-top: 8vh;
  /* 距离顶部8%的视口高度 */

}

/* 模态框动画 */
.modal-enter-from {
  opacity: 0;
}

.modal-enter-from .modal-box {
  transform: scale(0.9) translateY(-40px);
}

.modal-enter-to {
  opacity: 1;
}

.modal-enter-to .modal-box {
  transform: scale(1) translateY(0);
}

.modal-leave-from {
  opacity: 1;
}

.modal-leave-from .modal-box {
  transform: scale(1) translateY(0);
}

.modal-leave-to {
  opacity: 0;
}

.modal-leave-to .modal-box {
  transform: scale(0.9) translateY(-40px);
}

.modal-enter-active {
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-enter-active .modal-box {
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.modal-leave-active {
  transition: opacity 0.2s cubic-bezier(0.4, 0, 1, 1);
}

.modal-leave-active .modal-box {
  transition: transform 0.2s cubic-bezier(0.4, 0, 1, 1);
}

/* 确保模态框在最上层 */
.modal {
  z-index: 1000;
}

/* 多层模态框支持 */
.modal:nth-child(n+2) {
  z-index: 1010;
}

.modal:nth-child(n+3) {
  z-index: 1020;
}

/* 响应式优化 */
@media (max-width: 640px) {
  .modal-box {
    width: 95vw !important;
    max-width: none !important;
    margin: 1rem;
  }
}

/* 滚动条样式 */
.modal-content::-webkit-scrollbar {
  width: 6px;
}

.modal-content::-webkit-scrollbar-track {
  background: transparent;
}

.modal-content::-webkit-scrollbar-thumb {
  background: hsl(var(--bc) / 0.2);
  border-radius: 3px;
}

.modal-content::-webkit-scrollbar-thumb:hover {
  background: hsl(var(--bc) / 0.3);
}
</style>