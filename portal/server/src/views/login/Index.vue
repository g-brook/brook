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

// 响应式变量
const version = ref('');
const isRunning = ref<boolean | null>(null); // null 表示还没加载
const loadingError = ref(false);

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

onMounted(() => {
  loadBaseInfo();
});
</script>

<template>
  <div>

    <!-- 等待加载 -->
    <div v-if="isRunning === null">
      Loading...
    </div>

    <!-- 根据配置渲染组件 -->
    <div v-else>
      <Login v-if="isRunning" :version="version" />
      <Initializer v-else :isInit="!isRunning" />
    </div>
  </div>
</template>