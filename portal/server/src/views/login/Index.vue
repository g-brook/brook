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
import {onMounted, ref} from 'vue';
import baseInfo from '@/service/baseInfo';
import Login from '@/views/login/Login.vue';
import Initializer from '@/views/login/Initializer.vue';
import useI18n from '@/components/lang/useI18n';
import Icon from "@/components/icon/Index.vue";
import {getGlobalTheme} from '@/components/theme/useTheme';
import LanguageSwitcher from "@/components/lang/LanguageSwitcher.vue";

// 响应式变量
const version = ref('');
const isRunning = ref<boolean | null>(null); // null 表示还没加载
const loadingError = ref(false);

const { t } = useI18n();

// 使用全局主题管理
const { isDark, toggleTheme } = getGlobalTheme();

const loadBaseInfo = async () => {
  try {
    const res = await baseInfo.getBaseInfo();
    version.value = res.data.version;
    isRunning.value = res.data.isRunning;
  } catch (err) {
    console.error(err);
    loadingError.value = true;
  }
};

const handleGithubClick = () => {
  window.open('https://github.com/g-brook/brook', '_blank');
}

onMounted(() => {
  loadBaseInfo();
});
</script>

<template>
  <div class="bg-gradient-to-br from-primary/10 to-base-100">
    <div class="absolute  p-3  w-full
            bg-base-100 shadow-sm shadow-base-300/90 rounded-b-3xl
             border-primary">
      <!-- 操作按钮区域 -->
      <div class="flex items-center gap-2 justify-end">

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
    <!-- 等待加载 -->
    <div v-if="isRunning === null" class="flex items-center justify-center h-64">
      <div class="flex flex-col items-center space-y-4">
        <div class="loading loading-spinner loading-lg text-primary"></div>
        <p class="text-base-content/60">{{ t('common.loading') }}</p>
      </div>
    </div>

    <!-- 根据配置渲染组件 -->
    <div v-else>
      <Login v-if="isRunning" :version="version" />
      <Initializer v-else :isInit="!isRunning" />
    </div>
  </div>
</template>