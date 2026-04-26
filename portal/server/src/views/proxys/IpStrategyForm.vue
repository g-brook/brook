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
import { ref, onMounted } from 'vue';
import { useI18n } from '@/components/lang/useI18n';

const props = defineProps<{
  initialData?: any;
  isEdit?: boolean;
  onRegister?: (api: any) => void;
}>();

const { t } = useI18n();

const form = ref({
  id: 0,
  name: '',
  type: 'WL',
  status: 1
});

onMounted(() => {
  if (props.initialData) {
    form.value = {
      id: props.initialData.id || 0,
      name: props.initialData.name || '',
      type: props.initialData.type || 'WL',
      status: props.initialData.status || 1
    };
  }

  if (props.onRegister) {
    props.onRegister({
      handleSubmit: async () => {
        return form.value;
      }
    });
  }
});
</script>

<template>
  <div class="flex flex-col gap-4 p-4 min-w-[400px]">
    <div class="form-control">
      <label class="label">
        <span class="label-text font-bold">{{ t('menu.security.strategy.name') }}</span>
      </label>
      <input type="text" v-model="form.name" class="input input-bordered w-full" :placeholder="t('menu.security.strategy.name')" />
    </div>

    <div class="form-control">
      <label class="label">
        <span class="label-text font-bold">{{ t('menu.security.strategy.type') }}</span>
      </label>
      <select v-model="form.type" class="select select-bordered w-full">
        <option value="WL">{{ t('menu.security.strategy.whitelist') }}</option>
        <option value="BL">{{ t('menu.security.strategy.blacklist') }}</option>
        <option value="IL">{{ t('menu.security.strategy.privateOnly') }}</option>
      </select>
      <div v-if="form.type === 'IL'" class="mt-2 rounded-xl border border-base-content/10 bg-base-200/40 px-3 py-2 text-xs font-black opacity-60">
        {{ t('menu.security.strategy.ilUsesWhitelistTip') }}
      </div>
    </div>

  </div>
</template>
