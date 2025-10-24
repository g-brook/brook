/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {computed, ref, watch} from 'vue'

// 主题类型
export type Theme = 'light' | 'dark'

// 全局主题状态
const isDarkTheme = ref<boolean>(false) // 默认为 light 主题

// 主题管理 composable
export function useTheme() {
  // 切换主题
  const toggleTheme = () => {
    isDarkTheme.value = !isDarkTheme.value
  }

  // 设置主题
  const setTheme = (theme: Theme) => {
    isDarkTheme.value = theme === 'dark'
  }

  // 设置暗色主题状态
  const setDark = (dark: boolean) => {
    isDarkTheme.value = dark
  }

  // 应用主题到 DOM
  const applyTheme = () => {
    const theme = isDarkTheme.value ? 'dark' : 'light'
    document.documentElement.setAttribute('data-theme', theme)
    localStorage.setItem('isDark', JSON.stringify(isDarkTheme.value))
  }

  // 初始化主题
  const initTheme = () => {
    const saved = localStorage.getItem('isDark')
    if (saved !== null) {
      isDarkTheme.value = JSON.parse(saved)
    }
    applyTheme()
  }

  // 计算属性
  const isDark = computed(() => isDarkTheme.value)
  const isLight = computed(() => !isDarkTheme.value)
  const currentTheme = computed(() => isDarkTheme.value ? 'dark' : 'light')

  // 监听主题变化并自动应用
  watch(isDarkTheme, () => {
    applyTheme()
  })

  return {
    isDark,
    isLight,
    currentTheme,
    toggleTheme,
    setTheme,
    setDark,
    initTheme
  }
}

// 创建全局主题实例
let globalThemeInstance: ReturnType<typeof useTheme> | null = null

// 获取全局主题实例（单例模式）
export function getGlobalTheme() {
  if (!globalThemeInstance) {
    globalThemeInstance = useTheme()
    // 自动初始化
    globalThemeInstance.initTheme()
  }
  return globalThemeInstance
}