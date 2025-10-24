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
  <div class="dropdown dropdown-end">
    <label tabindex="0" class="btn btn-ghost btn-sm gap-2">
      <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 5h12M9 3v2m1.048 9.5A18.022 18.022 0 016.412 9m6.088 9h7M11 21l5-10 5 10M12.751 5C11.783 10.77 8.07 15.61 3 18.129" />
      </svg>
      {{ localeName }}
    </label>
    <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-40">
      <li v-for="localeOption in availableLocales" :key="localeOption">
        <a 
          @click="handleLocaleChange(localeOption)"
          :class="{ 'active': locale === localeOption }"
        >
          <span class="flex items-center gap-2">
            <span class="w-2 h-2 rounded-full" :class="locale === localeOption ? 'bg-primary' : 'bg-transparent'"></span>
            {{ getLocaleName(localeOption) }}
          </span>
        </a>
      </li>
    </ul>
  </div>
</template>

<script setup lang="ts">
import {type Locale, useI18n} from '@/components/lang/useI18n'

const { locale, localeName, availableLocales, setLocale } = useI18n()

// 获取语言显示名称
const getLocaleName = (localeCode: Locale): string => {
  const localeNames: Record<Locale, string> = {
    'zh-CN': '简体中文',
    'en-US': 'English'
  }
  return localeNames[localeCode]
}

// 处理语言切换
const handleLocaleChange = (newLocale: Locale) => {
  setLocale(newLocale)
}
</script>

<style scoped>
.dropdown-content {
  z-index: 1000;
}
</style>