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

<script lang="ts" setup>
import {computed, onMounted, ref} from 'vue'
import Icon from '@/components/icon/Index.vue';
import ms, {AuthToken} from '@/service/mysetting'
import Message from '@/components/message'
import useI18n from '@/components/lang/useI18n'
import TlsSetting from "@/views/mysetting/TlsSetting.vue";

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
        Message.success(t('success.tokenGenerated'))
        getToken()
      }
    })
  } finally {
    isGenerating.value = false
  }
}

// 撤销 Token
const revokeToken = async () => {
  if (!confirm(t('mysetting.confirmRevoke'))) {
    return
  }
  isRevoking.value = true
  try {
    ms.delToken().then(res => {
      if (res.success()) {
        tokenInfo.value = null
        Message.success(t('success.tokenRevoked'))
      }
    })
  } finally {
    isRevoking.value = false
  }
}

const getToken = () => {
  ms.getAuthToken<AuthToken>().then(res => {
    if (res.success()) {
      tokenInfo.value = res.data ||{
        token:""
      }
    }
  })
}

const tolgenToken =()=>{
  showToken.value=!showToken.value
}

const copyToken = ()=> {
  navigator.clipboard.writeText(tokenInfo.value.token)
      .then(() => Message.success(t('success.copied')))
      .catch(() => Message.error(t('errors.copyFailed')))
}

onMounted(() => {
  getToken()
})

const { t, locale } = useI18n()

</script>

<template>
  <div class="overflow-hidden">

    <div class="max-w-6xl mx-auto p-1 space-y-4 fade-in">
      <!-- Token 管理 - 参考 ConfigForm 风格 -->
      <div class="bg-base-200/40 rounded-3xl p-6 border border-base-content/5 space-y-6 shadow-sm mx-1">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
              <Icon icon="brook-token" style="font-size: 20px" />
            </div>
            <div>
              <h3 class="text-sm font-black uppercase tracking-widest">{{ t('mysetting.currentToken') }}</h3>
              <p v-if="tokenInfo.token" class="text-[10px] font-black opacity-30 uppercase tracking-tighter">
                {{ t('mysetting.createdAt') }}: {{ tokenInfo.createTime }}
              </p>
            </div>
          </div>
          <div v-if="tokenInfo.token" class="flex gap-2">
            <button @click="copyToken" class="btn btn-ghost btn-xs h-8 px-3 font-black uppercase tracking-widest hover:bg-primary hover:text-primary-content transition-all">
              <Icon icon="brook-copy" class="mr-1" />
              {{ t('common.copy') }}
            </button>
          </div>
        </div>

        <!-- Token 显示区域 -->
        <div v-if="tokenInfo.token" class="space-y-4">
          <div class="form-control w-full">
            <div class="relative group">
              <input :value="showToken ? tokenInfo.token : maskedToken" type="text"
                class="input input-bordered focus:input-primary w-full h-11 font-mono text-sm font-black tracking-tight bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm border-base-content/5 pr-24" readonly />
              
              <div class="absolute right-1 top-1 join">
                <button class="btn btn-ghost btn-sm h-9 join-item px-3 hover:bg-base-content/5" 
                        :title="showToken ? t('mysetting.hideToken') : t('mysetting.showToken')" @click="tolgenToken">
                  <Icon :icon="showToken ? 'brook-eye-close' : 'brook-eye'" style="font-size: 16px;" class="opacity-40" />
                </button>
              </div>
            </div>
          </div>

          <div class="flex gap-3 pt-2">
            <button @click="generateToken" class="btn btn-primary h-11 flex-1 font-black uppercase tracking-widest shadow-lg shadow-primary/20" :class="{ 'loading': isGenerating }"
              :disabled="isGenerating">
              <Icon v-if="!isGenerating" icon="brook-refresh" class="mr-2" />
              {{ t('mysetting.regenerate') }}
            </button>

            <button @click="revokeToken" class="btn btn-error btn-outline h-11 px-6 font-black uppercase tracking-widest border-2 hover:border-error" :disabled="isRevoking">
              <Icon icon="brook-delete" class="mr-2" />
              {{ t('mysetting.revoke') }}
            </button>
          </div>
        </div>

        <!-- 无 Token 状态 -->
        <div v-else class="text-center py-12 bg-base-100/30 rounded-2xl border border-dashed border-base-content/10">
          <div class="w-20 h-20 bg-base-200 rounded-3xl flex items-center justify-center mx-auto mb-6 rotate-12">
            <Icon icon="brook-token" class="text-primary/20" style="font-size: 48px;" />
          </div>
          <h4 class="text-lg font-black tracking-tight mb-2 opacity-80">{{ t('mysetting.noTokenTitle') }}</h4>
          <p class="text-xs font-medium opacity-40 leading-relaxed mb-8 max-w-xs mx-auto">{{ t('mysetting.noTokenDesc') }}</p>
          
          <button @click="generateToken" class="btn btn-primary btn-md gap-3 px-10 shadow-xl shadow-primary/20 font-black uppercase tracking-widest text-xs" :class="{ 'loading': isGenerating }" :disabled="isGenerating">
            <Icon v-if="!isGenerating" icon="brook-add" style="font-size: 18px;"/>
            {{ t('mysetting.generate') }}
          </button>
        </div>
      </div>

      <!-- TLS 设置部分 -->
      <div class="mx-1">
        <TlsSetting :key="`tls-setting-${locale}`"/>
      </div>
    </div>
  </div>
</template>
