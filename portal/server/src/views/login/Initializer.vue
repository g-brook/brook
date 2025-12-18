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
import baseInfo from '@/service/baseInfo';
import {computed, ref} from 'vue';
import {useI18n} from '@/components/lang/useI18n';
import LanguageSwitcher from '@/components/lang/LanguageSwitcher.vue';
import Message from '@/components/message';

const { t } = useI18n();

// 表单数据
const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const isLoading = ref(false);

// 表单验证
const isUsernameValid = computed(() => username.value.length >= 3);
const isPasswordValid = computed(() => password.value.length >= 6);
const isConfirmPasswordValid = computed(() => 
  confirmPassword.value === password.value && password.value.length > 0
);
const isFormValid = computed(() => 
  isUsernameValid.value && isPasswordValid.value && isConfirmPasswordValid.value
);

// 验证错误信息
const usernameError = computed(() => {
  if (username.value.length === 0) return '';
  if (!isUsernameValid.value) return t('validation.minLength', { min: 3 });
  return '';
});

const passwordError = computed(() => {
  if (password.value.length === 0) return '';
  if (!isPasswordValid.value) return t('validation.minLength', { min: 6 });
  return '';
});

const confirmPasswordError = computed(() => {
  if (confirmPassword.value.length === 0) return '';
  if (!isConfirmPasswordValid.value) return t('validation.passwordMismatch');
  return '';
});

// 处理初始化
const handleInit = async () => {
  if (!isFormValid.value) {
    Message.error(t('validation.required'));
    return;
  }

  try {
    isLoading.value = true;
    
    // 调用初始化API，传入用户名和密码
    const res = await baseInfo.initServer({
      username: username.value,
      password: password.value,
      confirmPassword: confirmPassword.value,
    });
    
    if (res.success()) {
      Message.success(t('initializer.initSuccess'));
      setTimeout(() => {
        window.location.reload();
      }, 1500);
    }
  } catch (error) {
    Message.error(t('initializer.initFailed'));
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-sky-50 to-indigo-100">
    <div class="transform">
      <div class="card-container w-[480px] p-[2px] rounded-2xl relative">
        <!-- Logo 区域 -->
        <div class="absolute -top-12 left-1/2 -translate-x-1/2 flex items-center justify-center">
          <div class="relative">
            <div class="absolute inset-1 w-30 h-30 rounded-full bg-gradient-to-r from-sky-300 via-blue-400 to-indigo-400 opacity-40 blur-2xl"></div>
            <img src="@/assets/svg-light.svg" class="w-30 h-auto relative z-10 drop-shadow-md" />
          </div>
        </div>

        <!-- 主卡片 -->
        <div class="card-content backdrop-blur-xl bg-white/60 border border-white/40 shadow-xl rounded-2xl p-8 relative z-10">
          <!-- 标题区域 -->
          <div class="flex flex-1 justify-between items-start mb-6">
            <div>
              <h2 class="text-2xl text-gray-800 tracking-wide font-semibold mb-2">
                {{ t('initializer.title') }}
              </h2>
             
            </div>
          </div>

          <!-- 提示信息 -->
          <div class="bg-blue-50/60 border border-blue-200/50 rounded-lg p-3 mb-4">
            <div class="flex items-center space-x-2">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4 text-blue-500 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <p class="text-xs text-blue-700 leading-tight">
                {{ t('initializer.description') }}
              </p>
            </div>
          </div>

          <!-- 初始化表单 -->
          <form @submit.prevent="handleInit" class="space-y-6">
            <!-- 用户名输入 -->
            <div class="form-control w-full relative">
              <label class="label">
                <span class="label-text text-gray-600 text-sm font-medium">
                  {{ t('common.username') }}
                  <span class="text-red-500">*</span>
                </span>
                <!-- 错误提示 - 与label对齐 -->
                <div class="h-4 overflow-hidden flex items-center">
                  <transition name="slide-down">
                    <span v-if="usernameError" class="text-red-500 text-xs leading-none">{{ usernameError }}</span>
                  </transition>
                </div>
              </label>
              <input 
                type="text" 
                v-model="username"
                :placeholder="t('login.usernamePlaceholder')"
                :class="[
                  'input input-bordered w-full rounded-xl bg-white/80 text-gray-800 placeholder-gray-400 border-gray-200 transition-all duration-300',
                  usernameError ? 'border-red-300 focus:border-red-400 focus:ring-red-200' : 'focus:ring-2 focus:ring-sky-400'
                ]"
                :disabled="isLoading"
                required
              />
            </div>

            <!-- 密码组 -->
            <div class="space-y-3">
              <!-- 密码输入 -->
              <div class="form-control w-full relative">
                <label class="label">
                  <span class="label-text text-gray-600 text-sm font-medium">
                    {{ t('common.password') }}
                    <span class="text-red-500">*</span>
                  </span>
                  <!-- 错误提示 - 与label对齐 -->
                  <div class="h-4 overflow-hidden flex items-center">
                    <transition name="slide-down">
                      <span v-if="passwordError" class="text-red-500 text-xs leading-none">{{ passwordError }}</span>
                    </transition>
                  </div>
                </label>
                <input 
                  type="password" 
                  v-model="password"
                  :placeholder="t('login.passwordPlaceholder')"
                  :class="[
                    'input input-bordered w-full rounded-xl bg-white/80 text-gray-800 placeholder-gray-400 border-gray-200 transition-all duration-300',
                    passwordError ? 'border-red-300 focus:border-red-400 focus:ring-red-200' : 'focus:ring-2 focus:ring-indigo-400'
                  ]"
                  :disabled="isLoading"
                  required
                />
              </div>

              <!-- 确认密码输入 -->
              <div class="form-control w-full relative">
                <label class="label">
                  <span class="label-text text-gray-600 text-sm font-medium">
                    {{ t('initializer.confirmPassword') }}
                    <span class="text-red-500">*</span>
                  </span>
                  <!-- 错误提示 - 与label对齐 -->
                  <div class="h-4 overflow-hidden flex items-center">
                    <transition name="slide-down">
                      <span v-if="confirmPasswordError" class="text-red-500 text-xs leading-none">{{ confirmPasswordError }}</span>
                    </transition>
                  </div>
                </label>
                <input 
                  type="password" 
                  v-model="confirmPassword"
                  :placeholder="t('initializer.confirmPasswordPlaceholder')"
                  :class="[
                    'input input-bordered w-full rounded-xl bg-white/80 text-gray-800 placeholder-gray-400 border-gray-200 transition-all duration-300',
                    confirmPasswordError ? 'border-red-300 focus:border-red-400 focus:ring-red-200' : 'focus:ring-2 focus:ring-purple-400'
                  ]"
                  :disabled="isLoading"
                  required
                />
              </div>
            </div>

             <!-- 提交按钮 -->
            <button
              type="submit"
              :disabled="!isFormValid || isLoading"
              :class="[
                'w-full py-3 rounded-xl font-semibold shadow-lg transform transition-all duration-300',
                isFormValid && !isLoading
                  ? 'bg-gradient-to-r from-sky-400 via-blue-500 to-indigo-500 text-white hover:scale-105 hover:shadow-sky-300/50 active:scale-95 cursor-pointer'
                  : 'bg-gray-300 text-gray-500 cursor-not-allowed'
              ]"
            >
              <span v-if="isLoading" class="flex items-center justify-center space-x-2">
                <svg class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                  <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                  <path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                </svg>
                <span>{{ t('initializer.initInProgress') }}</span>
              </span>
              <span v-else>{{ t('initializer.submitButton') }}</span>
            </button>

            <!-- 密码要求提示 -->
            <div class="bg-gray-50/80 border border-gray-200 rounded-lg p-3">
              <h4 class="text-xs font-medium text-gray-700 mb-2">{{ t('initializer.passwordRequirements') }}:</h4>
              <ul class="text-xs text-gray-600 space-y-1">
                <li class="flex items-center space-x-2">
                  <span :class="isPasswordValid ? 'text-green-500' : 'text-gray-400'">
                    {{ isPasswordValid ? '✓' : '○' }}
                  </span>
                  <span>{{ t('validation.minLength', { min: 6 }) }}</span>
                </li>
                <li class="flex items-center space-x-2">
                  <span :class="isUsernameValid ? 'text-green-500' : 'text-gray-400'">
                    {{ isUsernameValid ? '✓' : '○' }}
                  </span>
                  <span>{{ t('initializer.usernameRequirement') }}</span>
                </li>
                <li class="flex items-center space-x-2">
                  <span :class="isConfirmPasswordValid ? 'text-green-500' : 'text-gray-400'">
                    {{ isConfirmPasswordValid ? '✓' : '○' }}
                  </span>
                  <span>{{ t('initializer.passwordMatch') }}</span>
                </li>
              </ul>
            </div>
          </form>
         
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-container {
  position: relative;
  border-radius: 1rem;
  overflow: visible;
}

.card-content {
  position: relative;
  z-index: 10;
}

/* 输入框焦点效果 */
.input:focus {
  outline: none;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
}

/* 错误提示过渡动画 */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.2s ease-out;
}

.slide-down-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

.slide-down-leave-to {
  opacity: 0;
  transform: translateY(-8px);
}

.slide-down-enter-to,
.slide-down-leave-from {
  opacity: 1;
  transform: translateY(0);
}

/* 加载动画 */
@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}

.animate-spin {
  animation: spin 1s linear infinite;
}
</style>