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
import {ref} from 'vue'
import baseInfo from '@/service/baseInfo';
import {useRouter} from 'vue-router'
import {useI18n} from '@/components/lang/useI18n'

const props = defineProps({
    version:{
      type: String,
      required:true,
    }
})
const router = useRouter()
const username = ref('')
const password = ref('')
const { t } = useI18n()
const handleLogin = () => {
  baseInfo.login({ username: username.value, password: password.value })
    .then((res) => {
      if (res.code === "OK") {
        localStorage.setItem('token', res.data)
        router.replace('/index')
      }
    })
}
</script>

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
</style>

<template>
  <div class="min-h-screen flex items-center justify-center">
    <div class="transform">
      <div class="card-container w-96 p-[2px] rounded-2xl relative">
        <div class="absolute -top-12 left-1/2 -translate-x-1/2 flex items-center justify-center">
          <div class="relative">
            <div
              class="absolute inset-1 w-30 h-30 rounded-full bg-gradient-to-r from-sky-300 via-blue-400 to-indigo-400 opacity-40 blur-2xl">
            </div>
            <img src="@/assets/svg-light.svg" class="w-30 h-auto relative z-10 drop-shadow-md" />
          </div>
        </div>
        <div
          class="card-content backdrop-blur-xl bg-white/60 border border-white/40 shadow-xl rounded-2xl p-8 relative z-10">
          <div class="flex flex-1 justify-between">
            <h2 class="text-2xl  text-center text-gray-800 mb-8 tracking-wide font-semibold">
              {{ t('login.title') }}
            </h2>
                               <div class=" badge badge-xs badge-soft font-mono text-xs ">{{props.version}}</div>
          </div>
          <div class="form-control w-full mb-4">
            <label class="label">
              <span class="label-text text-gray-600 text-sm">{{ t('login.username') }}</span>
            </label>
            <input type="text" v-model="username" :placeholder="t('login.usernamePlaceholder')"
              class="input input-bordered w-full rounded-xl bg-white/80 text-gray-800 placeholder-gray-400 focus:ring-2 focus:ring-sky-400 border-gray-200 transition-all duration-300" />
          </div>

          <div class="form-control w-full mb-6">
            <label class="label">
              <span class="label-text text-gray-600 text-sm">{{ t('login.password') }}</span>
            </label>
            <input type="password" v-model="password" :placeholder="t('login.passwordPlaceholder')"
              class="input input-bordered w-full rounded-xl bg-white/80 text-gray-800 placeholder-gray-400 focus:ring-2 focus:ring-indigo-400 border-gray-200 transition-all duration-300" />
          </div>

          <button
            class="w-full py-3 cursor-pointer rounded-xl bg-gradient-to-r from-sky-400 via-blue-500 to-indigo-500 text-white font-semibold shadow-lg transform transition-all duration-300 hover:scale-105 hover:shadow-sky-300/50 active:scale-95"
            @click="handleLogin">
            {{ t('login.loginButton') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>