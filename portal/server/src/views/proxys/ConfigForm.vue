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
import config from "@/service/config";
import {computed, reactive, ref} from 'vue';
import Icon from "@/components/icon/Index.vue";


// 表单数据类型
interface ConfigForm {
  id: number | null;
  name: string;
  tag: string;
  remotePort: number | null;
  proxyId: string;
  protocol: string;
}

// 错误信息类型
interface FormErrors {
  name?: string;
  tag?: string;
  remotePort?: string;
  proxyId?: string;
  protocol?: string;
}

// Props
const props = defineProps<{
  isEdit?: boolean;
  initialData?: Partial<ConfigForm>;
  onRegister?: (api: { handleSubmit: () => Promise<boolean> }) => void;
}>();

// 事件定义
defineEmits<{
  close: [];
  submit: [data: ConfigForm];
}>();
// 协议类型选项
const protocolTypes = [
  { value: 'HTTP', label: 'HTTP' },
  { value: 'HTTPS', label: 'HTTPS' },
  { value: 'TCP', label: 'TCP' },
  { value: 'UDP', label: 'UDP' },
];

// 响应式数据
const loading = ref(false);
const form = reactive<ConfigForm>({
  id: props.initialData?.id || null,
  name: props.initialData?.name || '',
  tag: props.initialData?.tag || '',
  remotePort: props.initialData?.remotePort || 30001,
  proxyId: props.initialData?.proxyId || '',
  protocol: props.initialData?.protocol || ''
});

const errors = reactive<FormErrors>({});

// 计算属性
const isEdit = computed(() => props.isEdit || false);

// 表单验证
const validateForm = (): boolean => {
  // 清空之前的错误
  Object.keys(errors).forEach(key => {
    delete errors[key as keyof FormErrors];
  });

  let isValid = true;
  // Name 验证
  if (!form.name.trim()) {
    errors.name = '名称不能为空';
    isValid = false;
  } else if (form.name.length > 50) {
    errors.name = '名称长度不能超过50个字符';
    isValid = false;
  }

  // Port 验证
  if (!form.remotePort) {
    errors.remotePort = '端口不能为空';
    isValid = false;
  } else if (form.remotePort < 30001 || form.remotePort > 65535) {
    errors.remotePort = '端口范围必须在30001-65535之间';
    isValid = false;
  }
  // ProxyId 验证
  if (!form.proxyId.trim()) {
    errors.proxyId = '代理ID不能为空';
    isValid = false;
  } else if (!/^[a-zA-Z0-9_-]+$/.test(form.proxyId)) {
    errors.proxyId = '代理ID只能包含字母、数字、下划线和横线';
    isValid = false;
  }
  // Type 验证
  if (!form.protocol) {
    errors.protocol = '请选择协议类型';
    isValid = false;
  }
  return isValid;
};

// 提交表单
const handleSubmit = async () => {
  if (!validateForm()) {
    return Promise.reject(new Error('Validation failed'));
  }
  loading.value = true;
  try {
    // 发送请求
    const res = await config.addProxyConfig(form);
    if (res.success()) {
      return Promise.resolve(true);
    } else {
      return Promise.reject(false);
    }
  } catch (error) {
    return Promise.reject(false);
  } finally {
    loading.value = false;
  }
};

// 重置表单
const resetForm = () => {
  form.name = '';
  form.tag = '';
  form.remotePort = 30001;
  form.proxyId = '';
  form.protocol = '';
  Object.keys(errors).forEach(key => {
    delete errors[key as keyof FormErrors];
  });
};

// 暴露方法给父组件
defineExpose({
  resetForm,
  handleSubmit
});

if (props.onRegister) {
  props.onRegister({ handleSubmit });
}
</script>
<template>
  <div class="w-[50rem] h-[20rem]">
    <!-- name of each tab group should be unique -->
    <div class="tabs tabs-lift">
      <label class="tab">
        <input type="radio" name="my_tabs_4" checked=true />
        基础
      </label>
      <div class="tab-content bg-base-100 border-base-300 p-1">
        <form @submit.prevent="handleSubmit" class="mt-4">
            <div class="grid grid-cols-7 gap-2">
              <label class="flex p-2 rounded-full border border-base-300 h-10 cursor-pointer w-fit"
                v-for="type in protocolTypes" :key="type.value">
                <input type="radio" name="types" v-model="form.protocol" :value="type.value"
                  class="radio radio-accent radio-sm checked:bg-red-200 checked:text-red-600 checked:border-red-600" />
                <p class="px-2 text-sm">{{ type.label }}</p>
              </label>
            </div>
          <div class="flex flex-row gap-2 justify-between">
            <div class="fieldset border-base-300 w-full p-4">
              <!-- 代理ID -->
              <div class="form-control">
                <label class="label py-1 w-14">
                  <span class="label-text font-medium">代理ID <span class="text-red-500">*</span></span>
                </label>
                <div class="tooltip" data-tip="hello">
                  <Icon icon="brook-exclamation-circle" style="font-size: 14px;" />
                </div>
                <input type="text" v-model="form.proxyId"
                  :class="['input  focus:input-primary w-full', { 'input-error': errors.proxyId }]"
                  placeholder="请输入代理ID" />
                <label v-if="errors.proxyId" class="label py-1">
                  <span class="label-text-alt text-red-500 text-xs">{{ errors.proxyId }}</span>
                </label>
              </div>

              <!-- 名称 -->
              <div class="form-control">
                <label class="label py-1 w-14">
                  <span class="label-text  font-medium">名称 <span class="text-red-500">*</span></span>
                </label>
                <input type="text" v-model="form.name"
                  :class="['input  focus:input-primary w-full', { 'input-error': errors.name }]"
                  placeholder="请输入配置名称" />
                <label v-if="errors.name" class="label py-0">
                  <span class="label-text-alt text-red-500 text-xs">{{ errors.name }}</span>
                </label>
              </div>
              <div class="form-control">
                <label class="label py-1 w-14">
                  <span class="label-text  font-medium">标签</span>
                </label>
                <input type="text" v-model="form.tag" class="input  focus:input-primary w-full"
                  placeholder="请输入标签（可选）" />
              </div>
            </div>
            <div class="fieldset border-base-300 rounded-box w-xs p-4">              <!-- 端口 -->
              <div class="form-control">
                <label class="label py-1 w-14">
                  <span class="label-text  font-medium">端口 <span class="text-red-500">*</span></span>
                </label>
                <div class="tooltip" data-tip="hello">
                  <Icon icon="brook-exclamation-circle" style="font-size: 14px;" />
                </div>
                <input type="number" v-model.number="form.remotePort"
                  :class="['input  focus:input-primary w-full', { 'input-error': errors.remotePort }]"
                  placeholder="请输入端口号" min="30001" max="65535" />
                <label v-if="errors.remotePort" class="label py-1">
                  <span class="label-text-alt text-red-500 text-xs">{{ errors.remotePort }}</span>
                </label>
              </div>
            </div>
          </div>
        </form>
      </div>
    </div>

  </div>
</template>