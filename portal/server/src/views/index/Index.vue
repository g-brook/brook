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
import {onMounted, ref, shallowRef} from 'vue';
import Left from './Left.vue';
import Right from './Right.vue';
import {Menu, menus} from '@/components/menu/menus';
import baseInfo from "@/service/baseInfo";

const selectedComponent = shallowRef<any>(null)
const title = ref<string>("")
const describe = ref<string>("")
const icon = ref<string>("")
const isLoading = ref<boolean>(false)
const version = ref('');
// 点击菜单动态加载组件
const handleSelect = async (item: Menu) => {
    try {
        isLoading.value = true
        // 动态 import
        const module = await item.comp()
        selectedComponent.value = module.default
        title.value = item.title || ""
        describe.value = item.describe || ""
        icon.value = item.icon
    } catch (error) {
        console.error('Failed to load component:', error)
    } finally {
        isLoading.value = false
    }
}

const loadBaseInfo = async () => {
  try {
    const res = await baseInfo.getBaseInfo();
    version.value = res.data.version;
  } catch (err) {
  }
};

// 默认加载第一个菜单项
onMounted(async () => {
    if (menus.length > 0) {
        menus[0].active = true
        await handleSelect(menus[0])
    }
  loadBaseInfo()
})
</script>

<template>
    <div class="flex h-screen bg-base-300/50">
        <!-- 左侧导航栏 -->
        <div class="flex-shrink-0 p-2">
            <Left @update:select="handleSelect" :version="version" />
        </div>
        
        <!-- 右侧内容区域 -->
        <div class="flex-1 flex flex-col min-w-0">
            <Right 
                :selected="selectedComponent" 
                :describe="describe" 
                :icon="icon" 
                :title="title"
                :is-loading="isLoading"
            />
        </div>
    </div>
</template>

<style scoped>
/* 添加平滑过渡效果 */
.flex {
    transition: all 0.3s ease;
}
</style>