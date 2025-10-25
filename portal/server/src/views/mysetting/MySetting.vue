
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
import {computed, onMounted, ref} from 'vue'
import ms, {AuthToken} from '@/service/mysetting'
import Message from '@/components/message'

// Token 相关状态
const showToken = ref(false)
const isGenerating = ref(false)
const isRevoking = ref(false)

const tokenInfo =ref<AuthToken>({
  token:'',
  createTime:null
})

// 计算属性
const maskedToken = computed(() => {
  if (!tokenInfo.value?.token) return ''
  const token = tokenInfo.value.token
  return token.substring(0, 8) + '*'.repeat(Math.max(0, token.length - 16)) + token.substring(token.length - 8)
})


// 生成 Token
const generateToken = async () => {
  isGenerating.value = true
  try {
    ms.generateAuthToken().then(res => {
      if (res.success()) {
        Message.success('Token生成成功')
        getToken()
      }
    })
  } finally {
    isGenerating.value = false
  }
}

// 撤销 Token
const revokeToken = async () => {
  if (!confirm('确定要撤销当前令牌吗？撤销后所有使用此令牌的客户端将无法连接。')) {
    return
  }
  isRevoking.value = true
  try {
    ms.delToken().then(res => {
      if (res.success()) {
        tokenInfo.value = null
        Message.success('Token撤销成功')
      }
    })
  } finally {
    isRevoking.value = false
  }
}

const getToken = () => {
  ms.getAuthToken<AuthToken>().then(res => {
    if (res.success()) {
      tokenInfo.value = res.data
    }
  })
}

const tolgenToken =()=>{
  showToken.value=!showToken.value
}

const copyToken = () => {

}
onMounted(() => {
  getToken()
})

</script>

<template>
  <div class="min-h-screen">
    <div class="max-w-4xl mx-auto p-6 space-y-8 fade-in">

      <!-- Token 管理 - 简洁展示 -->
      <div class="space-y-4">
        <div class="flex items-center gap-3">
          <div class="w-8 h-8 bg-base-200 rounded-md flex items-center justify-center">
            <i class="iconfont brook-token"></i>
          </div>
          <div>
            <h2 class="text-base font-medium text-base-content">访问令牌</h2>
            <p class="text-xs text-base-content/60">用于客户端连接的安全令牌</p>
          </div>
        </div>

        <div class="border border-base-300 rounded-lg p-6 w-full">
          <!-- Token 显示 -->
          <div v-if="tokenInfo.token" class="space-y-4">
            <div class="form-control">
              <label class="label">
                <span class="label-text font-medium">当前令牌</span>
                <span class="badge badge-success badge-sm">{{ tokenInfo.createTime }}</span>
              </label>
              <div class="flex items-center gap-2">
                <div class="join w-full">
                <input :value="showToken ? tokenInfo.token : maskedToken" type="text"
                  class="input input-ghost flex-1 font-mono text-sm bg-base-200 join-item input-bordered" readonly />
                  <button  class="btn rounded-r-full border-0 join-item"
                           :title="showToken ? '隐藏令牌' : '显示令牌'" @click="tolgenToken">
                    <svg v-if="!showToken" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none"
                         viewBox="0 0 24 24" stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z" />
                    </svg>
                    <svg v-else xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                         stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                            d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.878 9.878L3 3m6.878 6.878L21 21" />
                    </svg>
                  </button>
                </div>
                  <button class="btn btn-secondary " title="复制令牌" @click="copyToken">
                    <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                      stroke="currentColor">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
                    </svg>
                  </button>
              </div>
            </div>
          </div>

          <!-- 无 Token 状态 -->
          <div v-else class="text-center py-8">
            <div class="w-16 h-16 bg-base-200 rounded-full flex items-center justify-center mx-auto mb-4">
              <i class="iconfont brook-token" style="font-size: 48px;"></i>
            </div>
            <h4 class="text-lg font-semibold text-base-content mb-2">尚未生成令牌</h4>
            <p class="text-sm text-base-content/60 mb-6">您需要生成一个访问令牌客户端才能连接</p>
          </div>

          <!-- Token 操作按钮 -->
          <div class="flex gap-3 mt-6">
            <button @click="generateToken" class="btn btn-primary flex-1" :class="{ 'loading': isGenerating }"
              :disabled="isGenerating">
              <svg v-if="!isGenerating" xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none"
                viewBox="0 0 24 24" stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
              </svg>
              {{ tokenInfo.token ? '重新生成令牌' : '生成令牌' }}
            </button>

            <button v-if="tokenInfo.token" @click="revokeToken" class="btn btn-error btn-outline" :disabled="isRevoking">
              <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24"
                stroke="currentColor">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2"
                  d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
              </svg>
              撤销令牌
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


