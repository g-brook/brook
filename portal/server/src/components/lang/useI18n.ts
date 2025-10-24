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

import {computed} from 'vue'
import {getAvailableLocales, getCurrentLocaleName, i18nState, type Locale, rt, setLocale, t} from '@/i18n'

/**
 * 国际化 Composable
 * 提供响应式的国际化功能
 */
export function useI18n() {
  // 当前语言（响应式）
  const locale = computed(() => i18nState.locale)
  
  // 当前语言显示名称（响应式）
  const localeName = computed(() => getCurrentLocaleName())
  
  // 翻译函数
  const translate = t
  
  // 响应式翻译函数
  const reactiveTranslate = rt
  
  // 切换语言
  const switchLocale = (newLocale: Locale) => {
    setLocale(newLocale)
  }
  
  // 获取所有可用语言
  const availableLocales = getAvailableLocales()
  
  // 检查是否为中文环境
  const isChinese = computed(() => locale.value === 'zh-CN')
  
  // 检查是否为英文环境
  const isEnglish = computed(() => locale.value === 'en-US')
  
  // 获取语言方向（中文和英文都是从左到右）
  const direction = computed(() => 'ltr')
  
  // 格式化数字（根据语言环境）
  const formatNumber = (num: number, options?: Intl.NumberFormatOptions) => {
    const localeCode = locale.value === 'zh-CN' ? 'zh-CN' : 'en-US'
    return new Intl.NumberFormat(localeCode, options).format(num)
  }
  
  // 格式化日期（根据语言环境）
  const formatDate = (date: Date | string | number, options?: Intl.DateTimeFormatOptions) => {
    const localeCode = locale.value === 'zh-CN' ? 'zh-CN' : 'en-US'
    const dateObj = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date
    return new Intl.DateTimeFormat(localeCode, options).format(dateObj)
  }
  
  // 格式化相对时间
  const formatRelativeTime = (date: Date | string | number) => {
    const now = new Date()
    const targetDate = typeof date === 'string' || typeof date === 'number' ? new Date(date) : date
    const diffInSeconds = Math.floor((now.getTime() - targetDate.getTime()) / 1000)
    
    if (diffInSeconds < 60) {
      return t('time.now')
    } else if (diffInSeconds < 3600) {
      const minutes = Math.floor(diffInSeconds / 60)
      return t('time.minutesAgo', { count: minutes })
    } else if (diffInSeconds < 86400) {
      const hours = Math.floor(diffInSeconds / 3600)
      return t('time.hoursAgo', { count: hours })
    } else if (diffInSeconds < 604800) {
      const days = Math.floor(diffInSeconds / 86400)
      return t('time.daysAgo', { count: days })
    } else if (diffInSeconds < 2592000) {
      const weeks = Math.floor(diffInSeconds / 604800)
      return t('time.weeksAgo', { count: weeks })
    } else if (diffInSeconds < 31536000) {
      const months = Math.floor(diffInSeconds / 2592000)
      return t('time.monthsAgo', { count: months })
    } else {
      const years = Math.floor(diffInSeconds / 31536000)
      return t('time.yearsAgo', { count: years })
    }
  }
  
  // 获取常用翻译的快捷方法
  const common = {
    loading: computed(() => t('common.loading')),
    submit: computed(() => t('common.submit')),
    cancel: computed(() => t('common.cancel')),
    confirm: computed(() => t('common.confirm')),
    save: computed(() => t('common.save')),
    edit: computed(() => t('common.edit')),
    delete: computed(() => t('common.delete')),
    add: computed(() => t('common.add')),
    create: computed(() => t('common.create')),
    update: computed(() => t('common.update')),
    search: computed(() => t('common.search')),
    reset: computed(() => t('common.reset')),
    back: computed(() => t('common.back')),
    next: computed(() => t('common.next')),
    close: computed(() => t('common.close')),
    view: computed(() => t('common.view')),
    manage: computed(() => t('common.manage')),
    settings: computed(() => t('common.settings')),
    username: computed(() => t('common.username')),
    password: computed(() => t('common.password')),
    status: computed(() => t('common.status')),
    online: computed(() => t('common.online')),
    offline: computed(() => t('common.offline')),
    success: computed(() => t('common.success')),
    error: computed(() => t('common.error')),
    warning: computed(() => t('common.warning')),
    info: computed(() => t('common.info'))
  }
  
  return {
    // 状态
    locale,
    localeName,
    isChinese,
    isEnglish,
    direction,
    availableLocales,
    
    // 方法
    t: translate,
    rt: reactiveTranslate,
    setLocale: switchLocale,
    formatNumber,
    formatDate,
    formatRelativeTime,
    
    // 常用翻译
    common
  }
}

// 导出类型
export type { Locale }

// 默认导出
export default useI18n