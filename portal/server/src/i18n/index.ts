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

import {computed, reactive} from 'vue'
import zhCN from './zh-CN'
import enUS from './en-US'

// 支持的语言类型
export type Locale = 'zh-CN' | 'en-US'

// 语言包类型
export type Messages = typeof zhCN

// 语言包映射
const messages: Record<Locale, Messages> = {
  'zh-CN': zhCN,
  'en-US': enUS
}

// 全局状态
const state = reactive({
  locale: (localStorage.getItem('locale') as Locale) || 'zh-CN'
})

// 设置语言
export const setLocale = (locale: Locale) => {
  state.locale = locale
  localStorage.setItem('locale', locale)
}

// 获取当前语言
export const getLocale = () => state.locale

// 翻译函数
export const t = (key: string, params?: Record<string, any>): string => {
  const keys = key.split('.')
  let value: any = messages[state.locale]
  
  for (const k of keys) {
    if (value && typeof value === 'object' && k in value) {
      value = value[k]
    } else {
      console.warn(`Translation key "${key}" not found for locale "${state.locale}"`)
      return key
    }
  }
  
  if (typeof value !== 'string') {
    console.warn(`Translation value for "${key}" is not a string`)
    return key
  }
  
  // 简单的参数替换
  if (params) {
    return value.replace(/\{(\w+)\}/g, (match, paramKey) => {
      return params[paramKey] !== undefined ? String(params[paramKey]) : match
    })
  }
  
  return value
}

// 响应式的翻译函数
export const rt = (key: string, params?: Record<string, any>) => {
  return computed(() => t(key, params))
}

// 获取所有支持的语言
export const getAvailableLocales = (): Locale[] => {
  return Object.keys(messages) as Locale[]
}

// 获取当前语言的显示名称
export const getCurrentLocaleName = () => {
  const localeNames: Record<Locale, string> = {
    'zh-CN': '简体中文',
    'en-US': 'English'
  }
  return localeNames[state.locale]
}

// 导出状态供响应式使用
export const i18nState = state

// 默认导出
export default {
  state,
  setLocale,
  getLocale,
  t,
  rt,
  getAvailableLocales,
  getCurrentLocaleName
}