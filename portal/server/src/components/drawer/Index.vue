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

<script setup lang="ts">
import {ref} from 'vue';
import Icon from '@/components/icon/Index.vue';

const emits = defineEmits<{
  "close": () => void;
}>()
// 抽屉状态
const isOpen = ref(false);

// 打开抽屉
const openDrawer = () => {
  isOpen.value = true;
};

// 关闭抽屉
const closeDrawer = () => {
  isOpen.value = false;
  emits('close');
};

// 切换抽屉状态
const toggleDrawer = () => {
  isOpen.value = !isOpen.value;
};

// 暴露方法给父组件
defineExpose({
  open: openDrawer,
  close: closeDrawer,
  toggle: toggleDrawer
});

// 接收props
const props = defineProps({
  position: {
    type: String,
    default: 'right', // 可选值: left, right, top, bottom
    validator: (value: string) => ['left', 'right', 'top', 'bottom'].includes(value)
  },
  overlay: {
    type: Boolean,
    default: true
  },
  width: {
    type: String,
    default: '80%'
  },
  height: {
    type: String,
    default: '80%'
  },
  isCollapsed: {
    type: Boolean,
    default: false
  },
  title: {
    type: String,
    default: ''
  },
  icon: {
    type: String,
    default: null
  }
});
</script>

<template>
  <!-- 抽屉容器 -->
  <div class="drawer-container overflow-hidden h-full">
    <!-- 抽屉触发器 - 可选使用 -->
    <slot name="trigger" :toggle="toggleDrawer"></slot>

    <!-- 遮罩层 -->
    <div v-if="overlay && isOpen" class="drawer-overlay" @click="isCollapsed ? closeDrawer() : false"></div>

    <!-- 抽屉内容 -->
    <div class="drawer bg-base-100 h-full flex flex-col rounded-l-2xl" :class="[
      isOpen ? 'drawer-open' : '',
      `drawer-${position}`
    ]" :style="position === 'left' || position === 'right' ?
        { width: width } :
        { height: height }">
      <!-- 关闭按钮 -->
      <div class="sticky top-0 flex flex-row p-3 w-full bg-base-300/80 items-center justify-between">
        <div class="flex flex-row items-center justify-center">
          <Icon :icon="icon" v-if="icon"/>
          <h3 v-if="title" class="font-bold text-lg ml-1">
            {{ title }}
          </h3>
        </div>
        <button class="btn btn-circle btn-sm btn-soft btn-outline" @click="closeDrawer">
          <Icon icon="brook-delete"/>
        </button>
      </div>
      <!-- 抽屉内容 -->
      <div class="drawer-content overflow-auto h-full" v-if="isOpen">
        <slot></slot>
      </div>

    </div>
  </div>
</template>

<style scoped>
.drawer-container {
  position: relative;
}

.drawer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  background-color: rgba(0, 0, 0, 0.5);
  z-index: 100;
}

.drawer {
  position: fixed;
  z-index: 100;
  transition: all 0.3s ease-in-out;
  box-shadow: 0 0 10px rgba(0, 0, 0, 0.2);
  overflow-y: auto;
}

.drawer-left {
  top: 0;
  left: 0;
  height: 100%;
  transform: translateX(-100%);
}

.drawer-right {
  top: 0;
  right: 0;
  height: 100%;
  transform: translateX(100%);
}

.drawer-top {
  top: 0;
  left: 0;
  width: 100%;
  transform: translateY(-100%);
}

.drawer-bottom {
  bottom: 0;
  left: 0;
  width: 100%;
  transform: translateY(100%);
}

.drawer-open.drawer-left,
.drawer-open.drawer-right {
  transform: translateX(0);
}

.drawer-open.drawer-top,
.drawer-open.drawer-bottom {
  transform: translateY(0);
}
</style>