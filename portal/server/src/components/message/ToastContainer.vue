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
        <div v-for="position in positions" :key="position" :class="getPositionClass(position)"
            class="fixed p-4 space-y-3 z-11000 pointer-events-none">
            <transition-group name="toast" tag="div" class="space-y-3">
                <div v-for="toast in getToastsByPosition(position)" :key="toast.id"
                    :class="getToastClass(toast.type)" 
                    class="pointer-events-auto cursor-pointer"
                    role="alert"
                    :aria-live="toast.type === 'error' ? 'assertive' : 'polite'"
                    @click="removeToast(toast.id)">
                    <div v-html="TOAST_ICONS[toast.type]" class="flex-shrink-0"></div>
                    <span class="flex-1 text-sm font-medium">{{ toast.message }}</span>
                    <button 
                        class="toast-close-btn flex-shrink-0 ml-3 p-1 rounded-full hover:bg-black/10 transition-colors duration-200" 
                        @click.stop="removeToast(toast.id)">
                        <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
                            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                        </svg>
                    </button>
                </div>
            </transition-group>
        </div>
    </div>
</template>

<script lang="ts" setup>
import {computed} from 'vue';
import {type ToastPosition, type ToastType, useToast} from "./useToast";

const { toasts, removeToast } = useToast();

const positions: ToastPosition[] = [
    "top-left",
    "top-right",
    "bottom-left",
    "bottom-right",
    "top-center",
    "bottom-center",
];

// 优化：使用常量存储图标，避免重复创建
const TOAST_ICONS: Record<ToastType, string> = {
    success: `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current flex-shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
    error: `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current flex-shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>`,
    warning: `<svg xmlns="http://www.w3.org/2000/svg" class="stroke-current flex-shrink-0 h-5 w-5" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" /></svg>`,
    info: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" class="h-5 w-5 shrink-0 stroke-current"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path></svg>`
};

// 优化：使用计算属性缓存过滤结果
const getToastsByPosition = (position: ToastPosition) => {
    return computed(() => toasts.filter(t => t.position === position && !t.removing)).value;
};

const getPositionClass = (position: ToastPosition): string => {
    switch (position) {
        case "top-left":
            return "top-4 left-4";
        case "top-right":
            return "top-4 right-4";
        case "bottom-left":
            return "bottom-4 left-4";
        case "bottom-right":
            return "bottom-4 right-4";
        case "top-center":
            return "top-4 left-1/2 transform -translate-x-1/2";
        case "bottom-center":
            return "bottom-4 left-1/2 transform -translate-x-1/2";
        default:
            return "top-4 right-4";
    }
};

const getToastClass = (type: ToastType): string => {
    const baseClasses = "alert flex items-center gap-3 min-w-80 max-w-md";
    switch (type) {
        case "success":
            return `${baseClasses} alert-success`;
        case "error":
            return `${baseClasses} alert-error`;
        case "warning":
            return `${baseClasses} alert-warning`;
        default:
            return `${baseClasses} alert-info`;
    }
};

</script>

<style scoped>
/* 优化的过渡动画 */
.toast-enter-from {
    opacity: 0;
    transform: translateX(100%) scale(0.95);
}

.toast-enter-to {
    opacity: 1;
    transform: translateX(0) scale(1);
}

.toast-leave-from {
    opacity: 1;
    transform: translateX(0) scale(1);
}

.toast-leave-to {
    opacity: 0;
    transform: translateX(100%) scale(0.95);
}

.toast-enter-active {
    transition: all 0.4s cubic-bezier(0.175, 0.885, 0.32, 1.275);
}

.toast-leave-active {
    transition: all 0.3s cubic-bezier(0.55, 0.085, 0.68, 0.53);
}

/* 针对左侧位置的特殊动画 */
.top-4.left-4 .toast-enter-from,
.bottom-4.left-4 .toast-enter-from {
    transform: translateX(-100%) scale(0.95);
}

.top-4.left-4 .toast-leave-to,
.bottom-4.left-4 .toast-leave-to {
    transform: translateX(-100%) scale(0.95);
}

/* 针对中心位置的特殊动画 */
.top-4.left-1\/2 .toast-enter-from,
.bottom-4.left-1\/2 .toast-enter-from {
    transform: translateX(-50%) translateY(-20px) scale(0.95);
}

.top-4.left-1\/2 .toast-leave-to,
.bottom-4.left-1\/2 .toast-leave-to {
    transform: translateX(-50%) translateY(-20px) scale(0.95);
}

/* 关闭按钮样式优化 */
.toast-close-btn {
    opacity: 0.7;
    transition: all 0.2s ease;
}

.toast-close-btn:hover {
    opacity: 1;
    background-color: rgba(0, 0, 0, 0.1);
    transform: scale(1.1);
}

.toast-close-btn:active {
    transform: scale(0.95);
}


/* 响应式优化 */
@media (max-width: 640px) {
    .fixed.p-4 {
        padding: 1rem 0.5rem;
        left: 0.5rem !important;
        right: 0.5rem !important;
        transform: none !important;
    }
    
    .alert {
        min-width: auto;
        max-width: none;
    }
}
</style>