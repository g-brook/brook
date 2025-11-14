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

<script lang="ts" setup>
import Icon from '@/components/icon/Index.vue';
import {getGlobalTheme} from '@/components/theme/useTheme';
import {useI18n} from '@/components/lang/useI18n';
import LanguageSwitcher from '@/components/lang/LanguageSwitcher.vue'

defineProps<{
    selected: any
    title: string
    describe: string
    icon: string
    isLoading?: boolean
}>()

// 使用全局主题管理
const { isDark, toggleTheme } = getGlobalTheme();

// 国际化
const { t } = useI18n();

const handleGithubClick = () => {
    window.open('https://github.com/g-brook/brook', '_blank');
}

</script>

<template>
    <div class="flex flex-col m-2 ml-0 rounded-2xl overflow-hidden h-full bg-base-100">
        <div class="px-3 pt-1 pb-1 bg-gradient-to-bl from-primary/10 from-10% to-base-300/10 to-40% rounded-t-2xl bg-base-200/30">
            <div v-if="selected" class="flex items-center">
                <!-- 图标 -->
                <div class="flex-shrink-0">
                    <div class="w-10 h-10 bg-primary/10 rounded-xl flex items-center justify-center">
                        <Icon v-if="icon" :icon="icon" class="icon_primary" style="font-size: 24px;" />
                    </div>
                </div>

                <!-- 标题和描述 -->
                <div class="flex-1 min-w-0 ml-2">
                    <h1 class="font-thin [&:first-line]:font-black text-base-content mb-1.5">{{ t(title) }}</h1>
                    <p class="text-xs text-base-content/60">{{ t(describe) }}</p>
                </div>


                <!-- 操作按钮区域 -->
                <div class="flex-shrink-0 flex items-center gap-2">
                   
                    <button class="btn btn-ghost btn-sm btn-square" @click="handleGithubClick">
                        <Icon icon="brook-github" style="font-size: 20px; pointer-events: none;" />
                    </button>


                    <label class="swap swap-rotate btn btn-ghost btn-sm btn-square">
                        <input 
                            type="checkbox" 
                            :checked="isDark" 
                            @change="toggleTheme"
                            class="hidden" 
                        />                
                        <!-- 太阳图标 - 在浅色主题时显示 -->
                        <svg v-show="!isDark" class="h-5 w-5 fill-current transition-transform duration-300" xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24">
                            <path
                                d="M5.64,17l-.71.71a1,1,0,0,0,0,1.41,1,1,0,0,0,1.41,0l.71-.71A1,1,0,0,0,5.64,17ZM5,12a1,1,0,0,0-1-1H3a1,1,0,0,0,0,2H4A1,1,0,0,0,5,12Zm7-7a1,1,0,0,0,1-1V3a1,1,0,0,0-2,0V4A1,1,0,0,0,12,5ZM5.64,7.05a1,1,0,0,0,.7.29,1,1,0,0,0,.71-.29,1,1,0,0,0,0-1.41l-.71-.71A1,1,0,0,0,4.93,6.34Zm12,.29a1,1,0,0,0,.7-.29l.71-.71a1,1,0,1,0-1.41-1.41L17,5.64a1,1,0,0,0,0,1.41A1,1,0,0,0,17.66,7.34ZM21,11H20a1,1,0,0,0,0,2h1a1,1,0,0,0,0-2Zm-9,8a1,1,0,0,0-1,1v1a1,1,0,0,0,2,0V20A1,1,0,0,0,12,19ZM18.36,17A1,1,0,0,0,17,18.36l.71.71a1,1,0,0,0,1.41,0,1,1,0,0,0,0-1.41ZM12,6.5A5.5,5.5,0,1,0,17.5,12,5.51,5.51,0,0,0,12,6.5Zm0,9A3.5,3.5,0,1,1,15.5,12,3.5,3.5,0,0,1,12,15.5Z" />
                        </svg>

                        <!-- 月亮图标 - 在深色主题时显示 -->
                        <svg v-show="isDark" class="h-5 w-5 fill-current transition-transform duration-300" xmlns="http://www.w3.org/2000/svg"
                            viewBox="0 0 24 24">
                            <path
                                d="M21.64,13a1,1,0,0,0-1.05-.14,8.05,8.05,0,0,1-3.37.73A8.15,8.15,0,0,1,9.08,5.49a8.59,8.59,0,0,1,.25-2A1,1,0,0,0,8,2.36,10.14,10.14,0,1,0,22,14.05,1,1,0,0,0,21.64,13Zm-9.5,6.69A8.14,8.14,0,0,1,7.08,5.22v.27A10.15,10.15,0,0,0,17.22,15.63a9.79,9.79,0,0,0,2.1-.22A8.11,8.11,0,0,1,12.14,19.73Z" />
                        </svg>
                    </label>
                       <LanguageSwitcher />
                </div>
            </div>
        </div>

        <!-- 内容区域 - 只有这个区域可以滚动 -->
        <div class="flex-1 overflow-auto min-h-0">
            <transition name="fade" mode="out-in">
                <!-- 加载状态 -->
                <div v-if="isLoading" key="loading" class="flex items-center justify-center h-64">
                    <div class="flex flex-col items-center space-y-4">
                        <div class="loading loading-spinner loading-lg text-primary"></div>
                        <p class="text-base-content/60">{{ t('common.loading') }}</p>
                    </div>
                </div>

                <!-- 组件内容 -->
                <div v-else-if="selected" key="content" class="h-full w-full">
                    <component :is="selected" :key="selected" />
                </div>

                <!-- 空状态 -->
                <div v-else key="empty" class="flex items-center justify-center h-64">
                    <div class="text-center">
                        <div class="w-24 h-24 bg-base-200 rounded-full flex items-center justify-center mx-auto mb-6">
                            <Icon icon="brook-github" class="text-base-content/30" style="font-size: 48px;" />
                        </div>
                        <h3 class="text-lg font-medium text-base-content/60 mb-2">{{ t('right.selectModule') }}</h3>
                        <p class="text-sm text-base-content/40 max-w-md">
                            {{ t('right.selectModuleDesc') }}
                        </p>
                    </div>
                </div>
            </transition>
        </div>
    </div>
</template>

<style scoped>
/* 淡入淡出过渡效果 */
.fade-enter-active,
.fade-leave-active {
    transition: all 0.3s ease;
}

.fade-enter-from {
    opacity: 0;
    transform: translateY(10px);
}

.fade-leave-to {
    opacity: 0;
    transform: translateY(-10px);
}


/* 滚动条样式 */
.overflow-auto::-webkit-scrollbar {
    width: 6px;
}

.overflow-auto::-webkit-scrollbar-track {
    background: transparent;
}

.overflow-auto::-webkit-scrollbar-thumb {
    background: hsl(var(--bc) / 0.2);
    border-radius: 3px;
}

.overflow-auto::-webkit-scrollbar-thumb:hover {
    background: hsl(var(--bc) / 0.3);
}

</style>
