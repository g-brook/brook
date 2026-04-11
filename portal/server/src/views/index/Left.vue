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
import {useI18n} from '@/components/lang/useI18n';

defineProps<{
  version?: string,
}>();

const emit = defineEmits<{
  (e: 'update:select', value: any): void
}>()

const menuList = ref(menus);
const isCollapsed = ref(false);
const expandedItems = ref<string[]>([]);

// 初始化时展开包含激活子菜单的父项，或者显式标记为展开的项
const initExpanded = () => {
  menuList.value.forEach(item => {
    const hasActiveChild = item.children && item.children.some(child => child.active);
    if (hasActiveChild || item.expanded) {
      expandedItems.value.push(item.title);
    }
  });
}
initExpanded();

// 使用全局主题管理
const {isDark} = getGlobalTheme();

// 国际化
const {t} = useI18n();

const toggleExpand = (item: any) => {
  const index = expandedItems.value.indexOf(item.title);
  if (index > -1) {
    expandedItems.value.splice(index, 1);
    item.expanded = false;
  } else {
    expandedItems.value.push(item.title);
    item.expanded = true;
  }
}

const isExpanded = (item: any) => {
  return expandedItems.value.includes(item.title);
}

const isParentActive = (item: any) => {
  if (item.active) return true;
  if (item.children && item.children.length > 0) {
    return item.children.some((child: any) => child.active);
  }
  return false;
}

const onSelect = (item: any) => {
  if (item.children && item.children.length > 0) {
    if (isCollapsed.value) {
      isCollapsed.value = false;
      if (!isExpanded(item)) {
        toggleExpand(item);
      }
    } else {
      toggleExpand(item);
    }
    
    // 如果父菜单本身也有关联组件，则执行选择逻辑
    if (item.comp) {
      setActive(item);
      emit('update:select', item);
    }
    return;
  }
  
  setActive(item);
  emit('update:select', item);
}

const setActive = (item: any) => {
  const clearActive = (items: any[]) => {
    items.forEach(m => {
      m.active = false;
      if (m.children) clearActive(m.children);
    });
  }
  clearActive(menuList.value);
  item.active = true;
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
    <div :class="[
            'flex flex-col h-full w-full rounded-2xl shadow-[0_8px_40px_rgba(0,0,0,0.10)] transition-all duration-300 relative',
            'overflow-visible',
            'bg-gradient-to-br from-primary/10 from-10% to-base-300/10 to-40%'
        ]">
      <!-- 顶部 Logo 区域 -->
      <div :class="['relative z-20', isCollapsed ? 'h-16 flex items-center justify-center' : 'navbar sticky top-0 px-1 py-2']">
        <div :class="[isCollapsed ? 'flex justify-center items-center w-full' : 'flex-1 flex items-center justify-between w-full']">
          <!-- Logo 切换动画 -->
          <div :class="['logo-container relative overflow-hidden flex items-center', isCollapsed ? 'justify-center' : '']">
            <transition name="logo-switch" mode="out-in">
              <div v-if="!isCollapsed" key="expanded"
                   class="flex items-center logo-expanded">
                <div class="flex-shrink-0 flex">
                  <img src="@/assets/logo.svg" alt="Brook Logo" class="h-10 w-10"/>
                  <img v-if="isDark" src="@/assets/font-light.svg" alt="Brook Logo"
                       style="max-width: 100px;" class="h-14 w-14"/>
                  <div v-else class="flex-shrink-0 flex">
                    <img src="@/assets/font-dark.svg" alt="Brook Logo"
                         style="max-width: 100px;" class="h-14 w-14"/>
                  </div>
                </div>
              </div>
              <div v-else key="collapsed" class="flex items-center justify-center logo-collapsed w-12 h-12 mx-auto">
                <div class="flex-shrink-0">
                  <img src="@/assets/logo.svg" alt="Brook Logo" class="h-10 w-10"/>
                </div>
              </div>
            </transition>
          </div>
        </div>

        <!-- 折叠按钮 - 展开状态下在右边，折叠状态下绝对定位 -->
        <button @click="toggleSidebar"
                :class="[
                    'btn btn-xs btn-circle border-0 bg-base-100 shadow-md transition-all duration-300 z-30',
                    isCollapsed ? 'absolute -right-3 top-6 rotate-180' : 'absolute -right-2 top-8'
                ]">
          <Icon icon="brook-Left-" class="text-base-content" style="font-size: 11px;"/>
        </button>
      </div>

      <div :class="['flex-1 py-4 h-full scrollbar-hide', isCollapsed ? 'overflow-visible' : 'overflow-y-auto']">
        <ul :class="['menu space-y-1 w-full', isCollapsed ? 'px-0' : 'px-2']">
          <li v-for="(item, index) in menuList" :key="index" class="w-full flex justify-center">
            <!-- 情况1: 完全没有子菜单 -->
            <a v-if="!item.children || item.children.length === 0"
               @click="onSelect(item)" :class="[
                                'group relative flex items-center rounded-xl transition-all duration-300',
                                isCollapsed ? 'justify-center w-12 h-12' : 'h-9 px-3 w-full',
                                'hover:bg-primary/10 hover:text-primary',
                                isParentActive(item) ? (isCollapsed? 'bg-primary text-primary-content shadow-lg shadow-primary/20' : 'font-semibold text-primary bg-primary/5') : 'text-base-content/60'
                            ]" :title="isCollapsed ? t(item.title) : ''">
              <div class="flex-shrink-0 w-6 h-6 flex items-center justify-center relative">
                <Icon v-if="item.icon" :icon="item.icon" :class="[
                                        'transition-colors duration-300',
                                        isParentActive(item) ? (isCollapsed ? 'text-primary-content' : 'icon_primary') : 'text-base-content/60 group-hover:text-primary'
                                    ]" style="font-size: 20px;"/>
              </div>
              <span v-show="!isCollapsed" :class="[
                                    'ml-2 font-mono whitespace-nowrap overflow-hidden text-ellipsis transition-all duration-300 flex-1',
                                ]">
                                    {{ t(item.title) }}
                                </span>
              <div v-if="item.active && !isCollapsed" :class="[
                                    'absolute w-2 h-2 bg-primary rounded-full transition-all duration-300 right-2 opacity-100'
                                ]"></div>
            </a>

            <!-- 情况2: 有子菜单 且 处于收起状态 -> 显示为悬浮式展开 (Dropdown) -->
            <div v-else-if="isCollapsed" class="dropdown dropdown-right dropdown-hover group overflow-visible flex justify-center w-12 h-12 p-0">
              <div tabindex="0" role="button" @click="onSelect(item)" :class="[
                                    'flex items-center rounded-xl transition-all duration-300 justify-center w-full h-full cursor-pointer relative',
                                    isParentActive(item) ? 'bg-primary text-primary-content shadow-lg shadow-primary/20' : 'text-base-content/60 hover:bg-primary/10 hover:text-primary'
                                ]">
                <div class="flex-shrink-0 w-6 h-6 flex items-center justify-center relative">
                  <Icon v-if="item.icon" :icon="item.icon" :class="[
                                            'transition-colors duration-300',
                                            isParentActive(item) ? 'text-primary-content' : 'text-base-content/60 group-hover:text-primary'
                                        ]" style="font-size: 20px;"/>
                  <!-- 子菜单标识点 -->
                  <div :class="[
                                            'absolute -top-0.5 -right-0.5 w-2 h-2 rounded-full border-2 transition-all duration-300',
                                            isParentActive(item) ? 'bg-primary-content border-primary' : 'bg-primary border-base-100'
                                        ]"></div>
                </div>
              </div>
              <!-- 悬浮菜单内容 -->
              <!-- before:content-[''] 用于增加一个透明的桥接区域，防止鼠标移向菜单时因间隙导致菜单关闭 -->
              <ul tabindex="0" class="dropdown-content z-[100] menu p-2 shadow-2xl bg-base-100 rounded-2xl w-56 ml-1
                                    border border-primary/10 backdrop-blur-md bg-opacity-95
                                    before:absolute before:inset-y-0 before:-left-4 before:w-4 before:content-['']">
                <div class="px-4 py-3 mb-1 border-b border-base-content/5">
                  <div class="text-xs font-bold text-primary tracking-wider uppercase opacity-70">{{ t(item.title) }}</div>
                  <div class="text-[10px] text-base-content/40 mt-0.5 truncate">{{ t(item.describe) }}</div>
                </div>
                <li v-for="(child, childIndex) in item.children" :key="childIndex">
                  <a @click="onSelect(child)" :class="[
                                            'group flex items-center rounded-xl transition-all duration-300 px-3 py-2 mb-1',
                                            child.active ? 'bg-primary/10 text-primary font-bold' : 'text-base-content/70 hover:bg-primary/5 hover:text-primary'
                                        ]">
                    <div class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
                      <Icon v-if="child.icon" :icon="child.icon" :class="[
                                                      'transition-colors duration-300',
                                                      child.active ? 'icon_primary' : 'text-base-content/40 group-hover:text-primary'
                                                  ]" style="font-size: 16px;"/>
                    </div>
                    <span class="ml-2 font-mono text-sm whitespace-nowrap overflow-hidden text-ellipsis transition-all duration-300 flex-1">
                                            {{ t(child.title) }}
                                        </span>
                    <div v-if="child.active" class="w-1.5 h-1.5 bg-primary rounded-full shadow-[0_0_8px_rgba(var(--p),0.5)]"></div>
                  </a>
                </li>
              </ul>
            </div>

            <!-- 情况3: 有子菜单 且 处于展开状态 -> 显示为手风琴 (Details) -->
            <details v-else :open="isExpanded(item)" class="group/details">
              <summary @click.prevent="toggleExpand(item)" :class="[
                                'flex items-center rounded-xl transition-all duration-300 cursor-pointer h-9 px-3 list-none',
                                'hover:bg-primary/10 hover:text-primary',
                                isParentActive(item) ? 'font-semibold text-primary bg-primary/5' : 'text-base-content/60'
                            ]">
                <div class="flex-shrink-0 w-5 h-5 flex items-center justify-center">
                  <Icon v-if="item.icon" :icon="item.icon" :class="[
                                            'transition-colors duration-300',
                                            isParentActive(item) ? 'icon_primary' : 'text-base-content/60 group-hover:text-primary'
                                        ]" style="font-size: 20px;"/>
                </div>
                <span class="ml-2 font-mono whitespace-nowrap overflow-hidden text-ellipsis transition-all duration-300 flex-1">
                                    {{ t(item.title) }}
                                </span>
                <Icon icon="brook-Down-" :class="[
                                    'transition-transform duration-300 ml-1 opacity-40 group-hover/details:opacity-100',
                                    isExpanded(item) ? 'rotate-180 text-primary opacity-100' : ''
                                ]" style="font-size: 10px;"/>
              </summary>
              <ul class="mt-1 space-y-1 ml-4 border-l border-base-content/10 pl-2">
                <li v-for="(child, childIndex) in item.children" :key="childIndex">
                  <a @click="onSelect(child)" :class="[
                                            'group relative flex items-center rounded-xl transition-all duration-300',
                                            'hover:bg-primary/10 hover:text-primary h-8 px-3',
                                            child.active ? 'font-semibold text-primary bg-primary/5' : 'text-base-content/60'
                                        ]">
                    <div class="flex-shrink-0 w-4 h-4 flex items-center justify-center">
                      <Icon v-if="child.icon" :icon="child.icon" :class="[
                                                    'transition-colors duration-300',
                                                    child.active ? 'icon_primary' : 'text-base-content/60 group-hover:text-primary'
                                                ]" style="font-size: 16px;"/>
                    </div>
                    <span class="ml-2 font-mono text-sm whitespace-nowrap overflow-hidden text-ellipsis transition-all duration-300 flex-1">
                                            {{ t(child.title) }}
                                        </span>
                    <div v-if="child.active" class="absolute w-1.5 h-1.5 bg-primary rounded-full right-2"></div>
                  </a>
                </li>
              </ul>
            </details>
          </li>
        </ul>
      </div>

      <div :class="[isCollapsed ? 'px-0 py-4 w-full flex flex-col items-center' : 'p-4']">
        <div v-show="!isCollapsed" class="space-y-2 text-xs">
          <div class="flex items-center justify-between">
            <span class="text-base-content/60">{{ t('main.systemStatus') }}</span>
            <div class="flex items-center space-x-1">
              <div class="w-2 h-2 bg-success rounded-full animate-pulse"></div>
              <span class="font-medium text-success">{{ t('common.online') }}</span>
            </div>
          </div>
          <div class="flex items-center justify-between">
            <span class="text-base-content/60">{{ t('common.version') }}</span>
            <span class="font-mono font-medium text-base-content">{{ version }}</span>
          </div>

        </div>

        <!-- 折叠状态下的简化显示 -->
        <div v-show="isCollapsed" class="flex flex-col items-center space-y-2">
          <div class="w-2 h-2 bg-success rounded-full animate-pulse" :title="t('common.online')"/>
          <div class="text-xs font-mono text-base-content/60 writing-mode-vertical" :title="t('common.version')">
            v{{ version }}
          </div>

        </div>
      </div>
    </div>
  </aside>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}

.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}

/* 移除原有的 menu-expand 动画，因为现在使用 details */
details > summary::-webkit-details-marker {
  display: none;
}

details summary {
  list-style: none;
}

.logo-switch-enter-active,
.logo-switch-leave-active {
  transition: all 0.3s ease;
}

.logo-switch-enter-from,
.logo-switch-leave-to {
  opacity: 0;
  transform: translateX(-10px);
}

.writing-mode-vertical {
  writing-mode: vertical-rl;
  text-orientation: mixed;
}
</style>
