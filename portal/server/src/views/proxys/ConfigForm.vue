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
import {computed, onMounted, reactive, ref} from 'vue';
import Icon from "@/components/icon/Index.vue";
import {useI18n} from '@/components/lang/useI18n';
import message from "@/components/message";


// 表单数据类型
interface ConfigForm {
  id: number | null;
  name: string;
  tag: string;
  remotePort: number | null;
  proxyId: string;
  protocol: string;
  destinationAddr: string | null;
  destinationPort: number | null;
  destination: string;
  strategyId: number | null;
}

// 错误信息类型
interface FormErrors {
  name?: string;
  tag?: string;
  remotePort?: string;
  proxyId?: string;
  protocol?: string;
  destinationAddr?: string;
  destinationPort?: string;
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
  {value: 'HTTP', label: 'HTTP', icon: 'brook-web', color: 'text-blue-500'},
  {value: 'HTTPS', label: 'HTTPS', icon: 'brook-https', color: 'text-green-500'},
  {value: 'TCP', label: 'TCP', icon: 'brook-technology_usb-cable', color: 'text-orange-500'},
  {value: 'UDP', label: 'UDP', icon: 'brook-a-01_UDP-2', color: 'text-purple-500'},
];

// 响应式数据
const loading = ref(false);
const form = reactive<ConfigForm>({
  id: props.initialData?.id || null,
  name: props.initialData?.name || '',
  tag: props.initialData?.tag || '',
  remotePort: props.initialData?.remotePort || 10000,
  proxyId: props.initialData?.proxyId || '',
  protocol: props.initialData?.protocol || '',
  destinationAddr: props.initialData?.destination?.split(":")[0] || '',
  destinationPort: props.initialData?.destination?.split(":")[1] ? parseInt(props.initialData.destination.split(":")[1]) : 0,
  destination: props.initialData?.destination || '',
  strategyId: props.initialData?.strategyId || null,
});

const errors = reactive<FormErrors>({});

// 计算属性
const isEdit = computed(() => props.isEdit || false);

const {t} = useI18n();

// IP 策略 Mock 数据 (实际应调用接口获取)
const strategies = ref<any[]>([
  { id: 1, name: 'Default Whitelist' },
  { id: 2, name: 'Block Malicious IPs' },
  { id: 3, name: 'Internal Only' }
]);

// 获取所有策略
const getIpStrategies = async () => {
  // Mock 数据已在上面定义，真实环境应调用接口
  // try {
  //   const res = await config.getIpStrategies();
  //   strategies.value = res.data || [];
  // } catch (e) {}
};

// 表单验证
const validateForm = (): boolean => {
  // 清空之前的错误
  Object.keys(errors).forEach(key => {
    delete errors[key as keyof FormErrors];
  });

  let isValid = true;
  // Name 验证
  if (!form.name.trim()) {
    errors.name = t('validation.required');
    isValid = false;
  } else if (form.name.length > 50) {
    errors.name = t('validation.maxLength', {max: 50});
    isValid = false;
  }
  // Port 验证
  if (!form.remotePort) {
    errors.remotePort = t('validation.required');
    isValid = false;
  } else if (form.remotePort < 10000 || form.remotePort > 65535) {
    errors.remotePort = t('validation.invalidPort');
    isValid = false;
  }
  // ProxyId 验证
  if (!form.proxyId.trim()) {
    errors.proxyId = t('validation.required');
    isValid = false;
  } else if (!/^[a-zA-Z0-9_-]+$/.test(form.proxyId)) {
    errors.proxyId = t('validation.alphanumericDashUnderscore');
    isValid = false;
  }
  // Type 验证
  if (!form.protocol) {
    errors.protocol = t('validation.required');
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
    let res;
    if (form.destinationAddr && form.destinationPort) {
      form.destination = form.destinationAddr + ":" + form.destinationPort
    } else {
      form.destination = ''
    }
    if (!props.isEdit) {
      res = await config.addProxyConfig(form);
    } else {
      res = await config.updateProxyConfig(form)
    }
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

const getPort = async () => {
  try {
    const res = await config.getRandomPort({});
    if (res.success()) {
      form.remotePort = res.data.port;
    }
  } catch (error) {
    console.error(error);
  }
};

// 重置表单
const resetForm = () => {
  form.name = '';
  form.tag = '';
  form.remotePort = 10000;
  form.proxyId = '';
  form.protocol = '';
  form.destinationAddr = '';
  form.destinationPort = null;
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
  props.onRegister({handleSubmit});
}
onMounted(() => {
  if (!props.isEdit) {
    getPort();
  }
  getIpStrategies();
});
</script>
<template>
  <div class="w-[34rem] p-2">
    <form @submit.prevent="handleSubmit" class="space-y-4">
      <!-- 现代化协议选择器 - 极简紧凑卡片 -->
      <div class="grid grid-cols-4 gap-3">
        <label
            v-for="type in protocolTypes"
            :key="type.value"
            :class="[
              'relative flex flex-col items-center justify-center gap-1.5 py-3 px-2 rounded-2xl cursor-pointer transition-all duration-300 border-2 overflow-hidden group',
              form.protocol === type.value
                ? 'bg-primary text-primary-content border-primary shadow-md scale-[1.02]'
                : 'bg-base-200/50 border-transparent hover:bg-base-200/80 hover:border-base-content/10'
            ]"
        >
          <input type="radio" name="types" v-model="form.protocol" :value="type.value" class="hidden"/>

          <!-- 图标容器 -->
          <div
              :class="[
                'flex items-center justify-center w-9 h-9 rounded-xl transition-all duration-300',
                form.protocol === type.value ? 'bg-white/20 text-white' : 'bg-base-100 text-base-content/40'
              ]"
          >
            <Icon :icon="type.icon" style="font-size: 18px;"/>
          </div>

          <!-- 文字内容 -->
          <span :class="['font-black text-[14px] tracking-widest uppercase', form.protocol === type.value ? 'text-white' : 'text-base-content/60']">
            {{ type.label }}
          </span>
        </label>
      </div>

      <!-- 内容区块：统一使用极简淡色背景 -->
      <div class="bg-base-200/40 rounded-3xl p-5 border border-base-content/5 space-y-5">
        <!-- 第一部分：基础信息 -->
        <div class="space-y-3">
          <div class="grid grid-cols-2 gap-5">
            <div class="form-control w-full">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em] flex items-center gap-1">
                  {{ t('configuration.proxyId') }}
                  <span class="text-error font-black">*</span>
                </span>
              </label>
              <div class="relative group">
                <input type="text" v-model="form.proxyId"
                       :class="['input input-bordered focus:input-primary w-full h-11 text-sm font-black tracking-tight pr-10 bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm border-base-content/5', { 'input-error': errors.proxyId }]"
                       :placeholder="t('configuration.form.proxyIdPlaceholder')"/>
                <div class="absolute right-3 top-3 tooltip tooltip-left" :data-tip="t('configuration.proxyIdTip')">
                  <Icon icon="brook-exclamation-circle" class="opacity-20 hover:opacity-100 transition-opacity cursor-help"/>
                </div>
              </div>
            </div>

            <div class="form-control w-full">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em] flex items-center gap-1">
                  {{ t('common.name') }}
                  <span class="text-error font-black">*</span>
                </span>
              </label>
              <input type="text" v-model="form.name"
                     :class="['input input-bordered focus:input-primary w-full h-11 text-sm font-black tracking-tight bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm border-base-content/5', { 'input-error': errors.name }]"
                     :placeholder="t('configuration.form.namePlaceholder')"/>
            </div>
          </div>
        </div>

        <!-- 极细分割线 -->
        <div class="h-px bg-base-content/5 mx-2"></div>

        <!-- 第二部分：网络配置 -->
        <div class="space-y-3">
          <div class="grid grid-cols-12 gap-5 items-end">
            <div class="form-control col-span-4">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em] flex items-center gap-1">
                  {{ t('configuration.remotePort') }}
                  <span class="text-error font-black">*</span>
                </span>
              </label>
              <input type="number" v-model.number="form.remotePort" :disabled="isEdit"
                     class="input input-bordered focus:input-primary w-full h-11 font-mono font-black text-sm bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm border-base-content/5"
                     min="10000" max="65535"/>
            </div>

            <div class="form-control col-span-8">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em]">
                  {{ t('configuration.destination') }}
                </span>
              </label>
              <div class="join w-full h-11 shadow-sm border border-base-content/5 rounded-xl overflow-hidden bg-base-100/30">
                <input type="text" v-model="form.destinationAddr"
                       class="input input-ghost join-item focus:bg-base-100 flex-1 min-w-0 text-sm font-black tracking-tight px-4"
                       :placeholder="t('configuration.form.destAddrPlaceholder')"/>
                <div class="bg-base-content/5 flex items-center px-3 font-mono text-xs font-black opacity-30 border-x border-base-content/5">:</div>
                <input type="number" v-model.number="form.destinationPort"
                       class="input input-ghost join-item focus:bg-base-100 w-24 text-sm font-mono font-black px-4"
                       placeholder="Port" max="65535"/>
              </div>
            </div>
          </div>
        </div>

        <!-- 极细分割线 -->
        <div class="h-px bg-base-content/5 mx-2"></div>

        <!-- 第三部分：高级配置 -->
        <div class="space-y-3">
          <div class="grid grid-cols-2 gap-5">
            <div class="form-control">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em]">{{ t('configuration.form.tagLabel') }}</span>
              </label>
              <input type="text" v-model="form.tag" class="input input-bordered focus:input-primary w-full h-11 bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm font-black text-sm tracking-tight border-base-content/5"
                     :placeholder="t('configuration.form.tagPlaceholder')"/>
            </div>

            <div class="form-control">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.title') }}</span>
              </label>
              <div class="relative">
                <select v-model="form.strategyId" class="select select-bordered focus:select-primary w-full h-11 font-black text-primary bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm appearance-none border-base-content/5 text-sm tracking-tight">
                  <option :value="null">{{ t('common.none') || 'None' }}</option>
                  <option v-for="s in strategies" :key="s.id" :value="s.id">{{ s.name }}</option>
                </select>
                <div class="absolute right-3 top-3.5 pointer-events-none opacity-20">
                  <Icon icon="brook-Down-" style="font-size: 14px;" />
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </form>
  </div>
</template>