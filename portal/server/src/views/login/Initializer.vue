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
import Message from '@/components/message';

const { t } = useI18n();

const username = ref('');
const password = ref('');
const confirmPassword = ref('');
const isLoading = ref(false);
const databaseLoading = ref(false);
const currentStep = ref(1);
const databaseInitialized = ref(false);
const initCompleted = ref(false);

const steps = computed(() => [
  {
    value: 1,
    title: t('initializer.steps.database.title'),
    description: t('initializer.steps.database.description'),
  },
  {
    value: 2,
    title: t('initializer.steps.admin.title'),
    description: t('initializer.steps.admin.description'),
  },
  {
    value: 3,
    title: t('initializer.steps.guide.title'),
    description: t('initializer.steps.guide.description'),
  },
]);

const isUsernameValid = computed(() => username.value.length >= 3);
const isPasswordValid = computed(() => password.value.length >= 6);
const isConfirmPasswordValid = computed(() =>
  confirmPassword.value === password.value && password.value.length > 0
);
const isFormValid = computed(() =>
  isUsernameValid.value && isPasswordValid.value && isConfirmPasswordValid.value
);

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

const goStep = (step: number) => {
  if (step > currentStep.value && step === 3 && !initCompleted.value) return;
  if (step > 2 && !initCompleted.value) return;
  currentStep.value = step;
};

const nextStep = () => {
  if (currentStep.value < 2 && databaseInitialized.value) {
    currentStep.value += 1;
  }
};

const handleInitDatabase = async () => {
  try {
    databaseLoading.value = true;
    const res = await baseInfo.initDatabase();
    if (res.success()) {
      databaseInitialized.value = true;
      currentStep.value = 2;
      Message.success(t('initializer.database.initSuccess'));
    } else {
      Message.error(res.message || t('initializer.database.initFailed'));
    }
  } catch (error) {
    Message.error(t('initializer.database.initFailed'));
  } finally {
    databaseLoading.value = false;
  }
};

const previousStep = () => {
  if (currentStep.value > 1 && !isLoading.value) {
    currentStep.value -= 1;
  }
};

const handleInit = async () => {
  if (!isFormValid.value) {
    Message.error(t('validation.required'));
    return;
  }

  try {
    isLoading.value = true;
    const res = await baseInfo.initServer({
      username: username.value,
      password: password.value,
      confirmPassword: confirmPassword.value,
    });

    if (res.success()) {
      initCompleted.value = true;
      currentStep.value = 3;
      Message.success(t('initializer.initSuccess'));
      setTimeout(() => {
        window.location.reload();
      }, 2500);
    }
  } catch (error) {
    Message.error(t('initializer.initFailed'));
  } finally {
    isLoading.value = false;
  }
};
</script>

<template>
  <div class="min-h-screen flex items-center justify-center bg-gradient-to-br from-primary/10 to-base-100 px-4 py-10">
    <div class="w-full max-w-5xl rounded-3xl bg-base-100/80 backdrop-blur-xl border border-base-content/5 shadow-xl overflow-hidden">
      <div class="p-6 border-b border-base-content/5">
        <div class="badge badge-primary badge-soft font-black tracking-widest mb-3">INIT</div>
        <div class="flex flex-col gap-2 md:flex-row md:items-end md:justify-between">
          <div>
            <h2 class="text-2xl font-black tracking-tight text-base-content">{{ t('initializer.title') }}</h2>
            <p class="text-xs opacity-50 leading-relaxed mt-2">{{ t('initializer.description') }}</p>
          </div>
          <div class="text-xs font-black opacity-40 uppercase tracking-widest">{{ t('initializer.stepProgress', { current: currentStep, total: 3 }) }}</div>
        </div>
      </div>

      <div class="p-6">
        <ul class="steps steps-vertical lg:steps-horizontal w-full mb-8">
          <li
            v-for="step in steps"
            :key="step.value"
            class="step cursor-pointer"
            :class="{
              'step-primary': currentStep >= step.value,
              'opacity-50': (step.value === 2 && !databaseInitialized) || (step.value === 3 && !initCompleted),
            }"
            @click="goStep(step.value)"
          >
            <span class="font-black">{{ step.title }}</span>
          </li>
        </ul>

        <transition name="slide-down" mode="out-in">
          <section v-if="currentStep === 1" key="database" class="space-y-6">
            <div class="rounded-3xl bg-base-200/40 border border-base-content/5 p-6">
              <div class="flex items-start justify-between gap-4 mb-5">
                <div>
                  <h3 class="text-xl font-black tracking-tight">{{ t('initializer.database.title') }}</h3>
                  <p class="text-xs opacity-50 mt-2">{{ t('initializer.database.description') }}</p>
                </div>
                <div class="badge badge-success badge-soft font-black">{{ t('initializer.database.ready') }}</div>
              </div>
              <div class="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div class="rounded-2xl bg-base-100 border border-base-content/5 p-4">
                  <div class="text-[11px] font-black opacity-40 uppercase tracking-widest">Database</div>
                  <div class="text-sm font-black mt-1">{{ t('initializer.database.structureTitle') }}</div>
                  <p class="text-xs opacity-40 mt-2 leading-relaxed">{{ t('initializer.database.structureDesc') }}</p>
                </div>
                <div class="rounded-2xl bg-base-100 border border-base-content/5 p-4">
                  <div class="text-[11px] font-black opacity-40 uppercase tracking-widest">Config</div>
                  <div class="text-sm font-black mt-1">{{ t('initializer.database.configTitle') }}</div>
                  <p class="text-xs opacity-40 mt-2 leading-relaxed">{{ t('initializer.database.configDesc') }}</p>
                </div>
                <div class="rounded-2xl bg-base-100 border border-base-content/5 p-4">
                  <div class="text-[11px] font-black opacity-40 uppercase tracking-widest">Admin</div>
                  <div class="text-sm font-black mt-1">{{ t('initializer.database.adminTitle') }}</div>
                  <p class="text-xs opacity-40 mt-2 leading-relaxed">{{ t('initializer.database.adminDesc') }}</p>
                </div>
              </div>
            </div>
            <div class="flex justify-end">
              <button
                type="button"
                class="btn btn-primary rounded-2xl px-8 font-black uppercase tracking-widest shadow-xl shadow-primary/20"
                :class="{ loading: databaseLoading }"
                :disabled="databaseLoading || databaseInitialized"
                @click="handleInitDatabase"
              >
                {{ databaseLoading ? t('initializer.database.initInProgress') : databaseInitialized ? t('initializer.database.initialized') : t('initializer.database.initButton') }}
              </button>
            </div>
          </section>

          <section v-else-if="currentStep === 2" key="admin" class="space-y-6">
            <div class="rounded-3xl bg-base-200/40 border border-base-content/5 overflow-hidden">
              <div class="px-6 py-4 border-b border-base-content/5">
                <h3 class="text-xl font-black tracking-tight">{{ t('initializer.admin.title') }}</h3>
                <p class="text-xs opacity-50 mt-1">{{ t('initializer.admin.description') }}</p>
              </div>

              <form @submit.prevent="handleInit" class="p-6 space-y-5">
                <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
                  <div class="form-control w-full relative">
                    <label class="label py-1">
                      <span class="label-text text-xs font-black opacity-50 uppercase tracking-widest">
                        {{ t('common.username') }} <span class="text-error">*</span>
                      </span>
                    </label>
                    <input
                      type="text"
                      v-model="username"
                      :placeholder="t('login.usernamePlaceholder')"
                      :class="[
                        'input input-bordered w-full rounded-2xl bg-base-100/70 border-base-content/10 transition-all duration-300',
                        usernameError ? 'input-error' : 'focus:input-primary'
                      ]"
                      :disabled="isLoading"
                      required
                    />
                    <transition name="slide-down">
                      <span v-if="usernameError" class="text-error text-xs mt-1">{{ usernameError }}</span>
                    </transition>
                  </div>

                  <div class="form-control w-full relative">
                    <label class="label py-1">
                      <span class="label-text text-xs font-black opacity-50 uppercase tracking-widest">
                        {{ t('common.password') }} <span class="text-error">*</span>
                      </span>
                    </label>
                    <input
                      type="password"
                      v-model="password"
                      :placeholder="t('login.passwordPlaceholder')"
                      :class="[
                        'input input-bordered w-full rounded-2xl bg-base-100/70 border-base-content/10 transition-all duration-300',
                        passwordError ? 'input-error' : 'focus:input-primary'
                      ]"
                      :disabled="isLoading"
                      required
                    />
                    <transition name="slide-down">
                      <span v-if="passwordError" class="text-error text-xs mt-1">{{ passwordError }}</span>
                    </transition>
                  </div>

                  <div class="form-control w-full relative">
                    <label class="label py-1">
                      <span class="label-text text-xs font-black opacity-50 uppercase tracking-widest">
                        {{ t('initializer.confirmPassword') }} <span class="text-error">*</span>
                      </span>
                    </label>
                    <input
                      type="password"
                      v-model="confirmPassword"
                      :placeholder="t('initializer.confirmPasswordPlaceholder')"
                      :class="[
                        'input input-bordered w-full rounded-2xl bg-base-100/70 border-base-content/10 transition-all duration-300',
                        confirmPasswordError ? 'input-error' : 'focus:input-primary'
                      ]"
                      :disabled="isLoading"
                      required
                    />
                    <transition name="slide-down">
                      <span v-if="confirmPasswordError" class="text-error text-xs mt-1">{{ confirmPasswordError }}</span>
                    </transition>
                  </div>
                </div>

                <div class="rounded-2xl bg-base-100 border border-base-content/5 p-4">
                  <h4 class="text-xs font-black opacity-50 uppercase tracking-widest mb-3">{{ t('initializer.passwordRequirements') }}</h4>
                  <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-xs">
                    <div class="flex items-center gap-2 font-black" :class="isUsernameValid ? 'text-success' : 'opacity-40'">
                      <span>{{ isUsernameValid ? '✓' : '○' }}</span>
                      <span>{{ t('initializer.usernameRequirement') }}</span>
                    </div>
                    <div class="flex items-center gap-2 font-black" :class="isPasswordValid ? 'text-success' : 'opacity-40'">
                      <span>{{ isPasswordValid ? '✓' : '○' }}</span>
                      <span>{{ t('validation.minLength', { min: 6 }) }}</span>
                    </div>
                    <div class="flex items-center gap-2 font-black" :class="isConfirmPasswordValid ? 'text-success' : 'opacity-40'">
                      <span>{{ isConfirmPasswordValid ? '✓' : '○' }}</span>
                      <span>{{ t('initializer.passwordMatch') }}</span>
                    </div>
                  </div>
                </div>

                <div class="flex flex-col-reverse gap-3 md:flex-row md:justify-between">
                  <button type="button" class="btn btn-ghost rounded-2xl px-8 font-black uppercase tracking-widest" @click="previousStep">
                    {{ t('common.previous') }}
                  </button>
                  <button
                    type="submit"
                    :disabled="!isFormValid || isLoading"
                    class="btn btn-primary rounded-2xl px-8 font-black uppercase tracking-widest shadow-xl shadow-primary/20"
                    :class="{ 'btn-disabled': !isFormValid || isLoading, loading: isLoading }"
                  >
                    {{ isLoading ? t('initializer.initInProgress') : t('initializer.submitButton') }}
                  </button>
                </div>
              </form>
            </div>
          </section>

          <section v-else key="success">
            <div class="rounded-3xl bg-success/10 border border-success/10 p-8 text-center">
              <div class="w-16 h-16 rounded-3xl bg-success text-success-content flex items-center justify-center font-black text-2xl mx-auto mb-5">✓</div>
              <h3 class="text-2xl font-black tracking-tight">{{ t('initializer.success.title') }}</h3>
              <p class="text-sm opacity-60 mt-3">{{ t('initializer.success.description') }}</p>
              <div class="loading loading-spinner loading-md text-success mt-6"></div>
            </div>
          </section>
        </transition>
      </div>
    </div>
  </div>
</template>

