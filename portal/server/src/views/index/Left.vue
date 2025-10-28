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
import {menus} from '@/components/menu/menus';
import {ref} from 'vue';
import {getGlobalTheme} from '@/components/theme/useTheme';

defineProps<{
    version?: string,
}>();

const emit = defineEmits<{
    (e: 'update:select', value: any): void
}>()

const menuList = ref(menus);
const isCollapsed = ref(false);

// 使用全局主题管理
const { isDark } = getGlobalTheme();

const onSelect = (item: any) => {
    menuList.value.forEach(m => (m.active = false));
    item.active = true
    emit('update:select', item)
}

const toggleSidebar = () => {
    isCollapsed.value = !isCollapsed.value;
}
</script>
<template>
    <!-- 左侧导航栏 -->
    <aside :class="[
        'transition-all duration-300 ease-in-out h-full ',
        isCollapsed ? 'w-16' : 'w-70'
    ]">
        <div class="flex flex-col h-full rounded-2xl shadow-[0_8px_40px_rgba(0,0,0,0.10)]
            bg-gradient-to-br
            from-primary/10 from-10%
            to-base-300/10 to-40%">
            <!-- 顶部 Logo 区域 -->
            <div class="navbar sticky top-0 z-20 ">
                <div class="flex-1 px-1 py-2">
                    <div class="flex items-center justify-between w-full">
                        <!-- Logo 切换动画 -->
                        <div class="logo-container relative overflow-hidden">
                            <transition name="logo-switch" mode="out-in">
                                <div v-if="!isCollapsed" key="expanded"
                                    class="flex items-center space-x-3 logo-expanded">
                                    <div class="flex-shrink-0">
                                        <img v-if="isDark" src="@/assets/svg-dark.svg" alt="Brook Logo"
                                            style="max-width: 100px;" />
                                        <img v-else src="@/assets/svg-light.svg" alt="Brook Logo"
                                            style="max-width: 100px;" />
                                    </div>
                                </div>
                                <div v-else key="collapsed" class="flex items-center justify-center logo-collapsed">
                                    <div class="flex-shrink-0">
                                        <img src="@/assets/svg-logo.svg" alt="Brook Logo" class="logo-image-small"
                                            style="max-width: 33px;" />
                                    </div>
                                </div>
                            </transition>
                        </div>
                    </div>
                </div>
                 <!-- 折叠按钮 - 展开状态下在右下角 -->
                    <div class="relative">
                        <button @click="toggleSidebar"
                            class="btn btn-xs btn-circle border-0  bg-base-100 transition-transform duration-300 absolute -left-1.8 -top-1"
                            :class="{ 'rotate-180': isCollapsed }">
                            <Icon icon="brook-Left-" class="text-base-content" style="font-size: 11px;" />
                        </button>
                    </div>
            </div>

            <div class="flex-1 overflow-y-auto  py-4 h-full">
                <ul class="menu px-2 space-y-1 w-full">
                    <li v-for="(item, index) in menuList" :key="index">
                        <a @click="onSelect(item)" :class="[
                            'group relative flex items-center rounded-xl transition-all duration-300',
                            'hover:bg-primary/10 hover:text-primary h-9',
                            item.active ? (isCollapsed? 'bg-primary' : 'font-semibold text-primary') : 'text-base-content/60'
                        ]" :title="isCollapsed ? item.title : ''">
                            <div  class="flex-shrink-0 w-5 h-5 flex items-center justify-center">
                                <Icon v-if="item.icon" :icon="item.icon" :class="[
                                    'transition-colors duration-300',
                                    item.active ? (isCollapsed ? 'text-primary-content' : 'icon_primary') : 'text-base-content/60 group-hover:text-primary'
                                ]" style="font-size: 20px;" />
                            </div>
                            <span v-show="!isCollapsed" :class="[
                                'ml-2  font-mono  whitespace-nowrap overflow-hidden text-ellipsis transition-all duration-300',
                            ]">
                                {{ item.title }}
                            </span>

                            <div v-if="item.active && !isCollapsed" :class="[
                                'absolute w-2 h-2 bg-primary rounded-full transition-all duration-300',
                                isCollapsed ? 'right-1 opacity-100' : 'right-2 opacity-100'
                            ]"></div>
                        </a>
                    </li>
                </ul>
            </div>

            <div class="p-4">
                <div v-show="!isCollapsed" class="space-y-2 text-xs">
                    <div class="flex items-center justify-between">
                        <span class="text-base-content/60">Node Status</span>
                        <div class="flex items-center space-x-1">
                            <div class="w-2 h-2 bg-success rounded-full animate-pulse"></div>
                            <span class="font-medium text-success">Online</span>
                        </div>
                    </div>
                    <div class="flex items-center justify-between">
                        <span class="text-base-content/60">Version</span>
                        <span class="font-mono font-medium text-base-content">{{version}}</span>
                    </div>

                </div>

                <!-- 折叠状态下的简化显示 -->
                <div v-show="isCollapsed" class="flex flex-col items-center space-y-2">
                    <div class="w-2 h-2 bg-success rounded-full animate-pulse" title="Online" />
                    <div class="text-xs font-mono text-base-content/60 writing-mode-vertical" title="v1.3.0">v{{version}}</div>

                </div>
            </div>
        </div>
    </aside>
</template>
