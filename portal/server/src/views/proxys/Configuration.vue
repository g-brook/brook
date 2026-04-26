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
import {computed, markRaw, onMounted, ref} from 'vue';
import * as dayjs from 'dayjs';
import Icon from '@/components/icon/Index.vue';
import Drawer from '@/components/drawer/Index.vue';
import Modal from '@/components/modal';
import ConfigFormComponent from './ConfigForm.vue';
import config from '@/service/config';
import {useI18n} from '@/components/lang/useI18n';
import WebConfiguration from "./WebConfiguration.vue";
import Message from '@/components/message';
import DownloadConfig from "./DownloadConfig.vue";
import message from "@/components/message";
import type { IpRule, IpStrategy } from '@/types/ip';

// 定义配置项的类型
interface ConfigItem {
  id: number;
  name: string;
  tag: string;
  remotePort: number;
  proxyId: string;
  protocol: string;
  state: boolean;
  isRunning: boolean;
  runtime: string;
  isExistWeb: boolean;
  clients: number;
  destination: string;
  strategyId: number | null;
}

const {t} = useI18n();
const ConfigForm = markRaw(ConfigFormComponent);

const configs = ref<ConfigItem[]>([]);
const webItem = ref<ConfigItem>(null);

// 计算统计信息
const totalConfigs = computed(() => configs.value.length);
const enabledConfigs = computed(() => configs.value.filter(item => item.state).length);
const runningConfigs = computed(() => configs.value.filter(item => item.isRunning).length);
const drawerRef = ref<{ open: () => void } | null>(null);
const downloadDrawerRef = ref<{ open: () => void } | null>(null);
const strategyDrawerRef = ref<{ open: () => void } | null>(null);
const selectedStrategy = ref<IpStrategy | null>(null);
const selectedStrategyRules = ref<IpRule[]>([]);
const strategyRulesLoading = ref(false);

const getConfigs = async () => {
  try {
    const res = await config.getProxyConfigs();
    configs.value = res.data || [];
  } catch (error) {
    console.error(error);
  }
};

// IP 策略数据
const strategies = ref<IpStrategy[]>([]);

const getStrategies = async () => {
  try {
    const res = await config.getAllStrategies();
    strategies.value = res.data || [];
  } catch (error) {
    console.error('Failed to fetch strategies:', error);
  }
};

const getStrategyName = (id: number | null) => {
  if (!id) return null;
  return strategies.value.find(s => s.id === id)?.name;
};

const getStrategy = (id: number) => {
  return strategies.value.find(s => s.id === id) || null;
};

const getStrategyTypeText = (type: string) => {
  switch (type) {
    case 'WL': return t('menu.security.strategy.whitelist');
    case 'BL': return t('menu.security.strategy.blacklist');
    case 'IL': return t('menu.security.strategy.privateOnly');
    default: return type || '-';
  }
};

const formatTime = (value: string) => {
  if (!value) return '-';
  if (typeof value === 'string' && value.startsWith('0001-01-01')) return '-';
  const fn = (dayjs as any).default || (dayjs as any);
  const d = fn(value);
  if (!d.isValid()) return value;
  return d.format('YYYY-MM-DD HH:mm:ss');
};

const openStrategyDetail = async (strategyId: number) => {
  if (!strategyId) return;
  if (!strategies.value.length) {
    await getStrategies();
  }
  selectedStrategy.value = getStrategy(strategyId);
  selectedStrategyRules.value = [];
  strategyDrawerRef.value?.open();
  if (!selectedStrategy.value) return;

  strategyRulesLoading.value = true;
  try {
    const res = await config.getIpRules(strategyId);
    if (res.success()) {
      selectedStrategyRules.value = res.data || [];
    } else {
      selectedStrategyRules.value = [];
    }
  } catch (e) {
    selectedStrategyRules.value = [];
  } finally {
    strategyRulesLoading.value = false;
  }
};

// 协议图标映射
const protocolIcons: Record<string, { icon: string, class: string }> = {
  'HTTP': { icon: 'brook-web', class: 'badge-info' },
  'HTTPS': { icon: 'brook-https', class: 'badge-success' },
  'TCP': { icon: 'brook-technology_usb-cable', class: 'badge-warning' },
  'UDP': { icon: 'brook-a-01_UDP-2', class: 'badge-secondary' },
};

// 组件挂载时获取数据
onMounted(() => {
  getConfigs();
  getStrategies();
});

const openWebConfig = (item: ConfigItem) => {
  if (item.protocol !== `HTTPS` && item.protocol !== `HTTP`) {
    return
  }
  webItem.value = item;
  drawerRef.value?.open();
}

const handleAdd = () => {
  let formApi: { handleSubmit: () => Promise<boolean> } | null = null;
  let modalId = '';
  modalId = Modal.open(ConfigForm, {
    title: t('configuration.addTunnelConfig'),
    size: 'auto',
    closable: true,
    maskClosable: true,
    showFooter: true,
    props: {
      onRegister: (api) => {
        formApi = api;
      },
    },
    onConfirm: async () => {
      if (formApi) {
        try {
          const formData = await formApi.handleSubmit();
          if (formData) {
            await getConfigs();
            message.info(t("common.success"));
            Modal.close(modalId);
          }
        } catch (error) {
          console.error('Failed to submit form:', error);
        }
      }
    },
  });
};


const handleDelete = (id: number) => {
  Modal.confirm({
    onConfirm: async () => {
      try {
        const e = await config.delProxyConfig(id);
        if (e.success()) {
          await getConfigs();
          message.info(t("common.success"));
          return true as any;
        }
      } catch (error) {
        console.error(error);
      }
      return false as any;
    }
  });
};

const handleUpdate = (cfg) => {
  let formApi: { handleSubmit: () => Promise<boolean> } | null = null;
  let modalId = '';
  modalId = Modal.open(ConfigForm, {
    title: t('configuration.editTunnelConfig'),
    size: 'auto',
    closable: true,
    maskClosable: false,
    showFooter: true,
    props: {
      onRegister: (api) => {
        formApi = api;
      },
      initialData: cfg,
      isEdit: true,
    },
    onConfirm: async () => {
      if (formApi) {
        try {
          const formData = await formApi.handleSubmit();
          if (formData) {
            await getConfigs();
            message.info(t("common.success"));
            Modal.close(modalId);
          }
        } catch (error) {
          console.error('Failed to submit form:', error);
        }
      }
    },
  });
};

const handleToggleStatus = async (id: number, state: boolean) => {
  try {
    const res = await config.updateProxyState({
      id: id,
      state: state ? 0 : 1
    });
    if (res.success()) {
      Message.info(t('success.configurationUpdated'))
      await getConfigs()
    }
  } catch (error) {
  }
};
const handleClickDownload = () => {
  downloadDrawerRef.value?.open();
}
</script>

<template>
  <div class="overflow-hidden">
    <Drawer ref="strategyDrawerRef"
            :title="selectedStrategy ? `${t('menu.security.strategy.title')} - ${selectedStrategy.name}` : t('menu.security.strategy.title')"
            icon="brook-security"
            width="50%">
      <div class="p-6 flex flex-col gap-6">
        <div v-if="selectedStrategy" class="bg-base-200/40 rounded-3xl p-5 border border-base-content/5">
          <div class="grid grid-cols-2 gap-4">
            <div class="flex flex-col gap-1">
              <span class="text-[11px] font-black opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.name') }}</span>
              <span class="text-sm font-black tracking-tight">{{ selectedStrategy.name }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-[11px] font-black opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.type') }}</span>
              <span class="text-sm font-black tracking-tight">{{ getStrategyTypeText(selectedStrategy.type) }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-[11px] font-black opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.status') }}</span>
              <span class="text-sm font-black tracking-tight">{{ selectedStrategy.status === 1 ? t('configuration.enabled') : t('configuration.disabled') }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-[11px] font-black opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.createdAt') }}</span>
              <span class="text-xs font-mono font-black opacity-60">{{ formatTime(selectedStrategy.created_at) }}</span>
            </div>
            <div class="flex flex-col gap-1">
              <span class="text-[11px] font-black opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.strategy.updatedAt') }}</span>
              <span class="text-xs font-mono font-black opacity-60">{{ formatTime(selectedStrategy.updated_at) }}</span>
            </div>
          </div>
        </div>

        <div class="flex-1 overflow-hidden flex flex-col rounded-3xl border border-base-content/5 bg-base-100 shadow-sm">
          <div class="px-5 py-3 border-b border-base-content/5 flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Icon icon="brook-web" style="font-size: 16px;" class="opacity-40" />
              <span class="text-xs font-black uppercase tracking-widest opacity-60">{{ t('menu.security.rules.title') }}</span>
            </div>
            <button class="btn btn-ghost btn-xs btn-circle" :class="{ loading: strategyRulesLoading }"
                    @click="selectedStrategy && openStrategyDetail(selectedStrategy.id)">
              <Icon icon="brook-refresh" style="font-size: 14px;" />
            </button>
          </div>
          <div class="overflow-y-auto flex-1">
            <table class="table table-md table-pin-rows">
              <thead class="bg-base-200/50">
              <tr>
                <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.ip') }}</th>
                <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.remark') }}</th>
              </tr>
              </thead>
              <tbody>
              <tr v-if="strategyRulesLoading">
                <td colspan="2" class="text-center py-10 opacity-30">
                  {{ t('common.loading') || 'Loading' }}
                </td>
              </tr>
              <tr v-else v-for="rule in selectedStrategyRules" :key="rule.id" class="hover:bg-base-200/40 transition-colors">
                <td class="font-mono font-black text-sm text-primary tracking-tight">{{ rule.ip }}</td>
                <td class="text-xs font-black opacity-40 tracking-tight">{{ rule.remark }}</td>
              </tr>
              <tr v-if="!strategyRulesLoading && selectedStrategyRules.length === 0">
                <td colspan="2" class="text-center py-10 opacity-30">
                  {{ t('pagination.noData') }}
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Drawer>

    <Drawer ref="drawerRef" :title="t('configuration.webInfoConfig')" icon="brook-web" width="50%" @close="getConfigs">
      <WebConfiguration :refProxyId="webItem?.id" :protocol="webItem?.protocol"/>
    </Drawer>
    <Drawer ref="downloadDrawerRef" :title="t('configuration.template')" icon="brook-empty" width="50%">
      <DownloadConfig/>
    </Drawer>
    <!-- 极简页头：整合统计与操作 -->
    <div class="flex sticky top-0 items-center h-14 justify-between gap-4 mb-3 px-5 py-2 rounded-xl bg-base-100/80 backdrop-blur-md z-30 border border-base-content/5 shadow-sm mx-1">
      <div class="flex items-center gap-6">
        <!-- 垂直分割线 -->
        <!-- 整合后的微缩统计 -->
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5 group cursor-help" :title="t('configuration.totalConfigsTitle')">
            <div class="w-1.5 h-1.5 rounded-full bg-primary opacity-40"></div>
            <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('common.total') || 'Total' }}</span>
            <span class="text-sm font-black tracking-tighter">{{ totalConfigs }}</span>
          </div>
          <div class="flex items-center gap-1.5 group cursor-help" :title="t('configuration.enabledConfigsTitle')">
            <div class="w-1.5 h-1.5 rounded-full bg-success opacity-40"></div>
            <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('configuration.enabled') || 'Enabled' }}</span>
            <span class="text-sm font-black tracking-tighter text-success">{{ enabledConfigs }}</span>
          </div>
          <div class="flex items-center gap-1.5 group cursor-help" :title="t('configuration.runningConfigsTitle')">
            <div class="w-1.5 h-1.5 rounded-full bg-info opacity-40"></div>
            <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('common.running') || 'Running' }}</span>
            <span class="text-sm font-black tracking-tighter text-info">{{ runningConfigs }}</span>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-1.5">
        <button class="btn btn-ghost btn-xs h-8 gap-1.5 font-bold px-2 hover:bg-base-content/5 text-xs uppercase tracking-widest opacity-60 hover:opacity-100" @click="handleClickDownload">
          <Icon icon="brook-empty" style="font-size: 12px;"/>
          {{ t('configuration.template') }}
        </button>
        <button class="btn btn-primary btn-xs h-8 gap-1.5 font-bold px-3 shadow-md shadow-primary/20 text-xs uppercase tracking-widest" @click="handleAdd">
          <Icon icon="brook-add" style="font-size: 12px;"/>
          {{ t('common.add') }}
        </button>
        <div class="divider divider-horizontal mx-0.5 w-px h-4 self-center opacity-10"></div>
        <button class="btn btn-circle btn-xs h-8 w-8 btn-ghost hover:rotate-180 transition-transform duration-500" @click="getConfigs">
          <Icon icon="brook-refresh" style="font-size: 14px;"/>
        </button>
      </div>
    </div>

    <!-- 配置表格 - 优化列宽与视觉样式 -->
    <div class="overflow-x-auto rounded-3xl border border-base-content/5 bg-base-100 shadow-sm mx-1">
      <table class="table table-md">
        <!-- head -->
        <thead class="bg-base-200/50">
        <tr>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 40px">#</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('configuration.nameAndTag') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 220px">{{ t('configuration.proxyId') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 100px">{{ t('configuration.remotePort') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 180px">{{ t('configuration.destination') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 100px">{{ t('configuration.protocol') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 200px">{{ t('menu.security.strategy.title') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 190px">{{ t('configuration.status') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 150px">{{ t('server.runtime') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center">{{ t('configuration.actions') }}</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(config, index) in configs" :key="config.id" class="hover:bg-base-200/40 transition-colors group">
          <th class="text-center opacity-30 font-mono text-xs">{{ index + 1 }}</th>
          <td>
            <div class="flex flex-col gap-0.5">
              <div class="flex items-center gap-2">
                <div class="relative flex items-center justify-center">
                  <div v-if="config.isRunning" class="absolute w-2 h-2 bg-success rounded-full animate-ping opacity-75"></div>
                  <div :class="['w-2 h-2 rounded-full relative z-10', config.isRunning ? 'bg-success' : 'bg-base-300']"></div>
                </div>
                <span class="font-black text-sm tracking-tight">{{ config.name }}</span>
                <span v-if="config.tag" :class="['badge badge-xs font-black px-1.5 py-2 scale-95', config.isRunning ? 'badge-success' : 'badge-ghost opacity-50']">
                  {{ config.tag }}
                </span>
              </div>
              <div v-if="config.strategyId" class="flex items-center gap-1 opacity-40 scale-95 origin-left">
                <Icon icon="brook-security" style="font-size: 10px;" />
                <span class="text-[10px] font-black uppercase tracking-widest">{{ getStrategyName(config.strategyId) }}</span>
              </div>
            </div>
          </td>
          <td>
            <div class="flex items-center gap-2">
              <span class="text-[10px] font-black opacity-20 uppercase tracking-tighter">ID:</span>
              <code class="text-sm font-black text-primary tracking-tight">{{ config.proxyId }}</code>
            </div>
          </td>
          <td class="font-mono font-black text-sm tracking-tighter">{{ config.remotePort }}</td>
          <td>
            <div class="flex flex-col gap-0">
              <span class="text-sm font-black tracking-tight opacity-70">{{ config.destination }}</span>
              <span class="text-[10px] opacity-30 uppercase font-black tracking-widest">Target</span>
            </div>
          </td>
          <td>
            <div :class="['badge badge-soft flex items-center gap-1.5 w-fit px-3 py-2.5 border border-current/5', protocolIcons[config.protocol]?.class || 'badge-ghost']">
              <Icon :icon="protocolIcons[config.protocol]?.icon || 'brook-Down-'" style="font-size: 14px;" />
              <span class="font-black text-[10px] tracking-widest uppercase">{{ config.protocol }}</span>
            </div>
          </td>
          <td>
            <button v-if="config.strategyId"
                    class="btn btn-ghost btn-xs h-8 gap-2 px-2 cursor-pointer"
                    @click="openStrategyDetail(config.strategyId)">
              <Icon icon="brook-security" style="font-size: 14px;" class="opacity-40" />
              <span class="text-xs font-black tracking-tight underline underline-offset-4">
                {{ getStrategyName(config.strategyId) }}
              </span>
            </button>
            <span v-else class="text-xs opacity-20 font-black">-</span>
          </td>
          <td>
            <div class="flex flex-col gap-1.5">
              <div class="flex items-center gap-2">
                <input type="checkbox" class="toggle toggle-primary toggle-sm"
                       :checked="config.state" @change="handleToggleStatus(config.id, config.state)"/>
                <span :class="['text-[10px] font-black uppercase tracking-[0.1em]', config.state ? 'text-primary' : 'opacity-20']">
                  {{ config.state ? t('configuration.enabled') : t('configuration.disabled') }}
                </span>
              </div>
              <div :class="['badge badge-xs gap-1 font-black border-none py-2 origin-left scale-90', config.isRunning ? 'bg-success/10 text-success' : 'bg-error/10 text-error']">
                <div :class="['w-1 h-1 rounded-full', config.isRunning ? 'bg-success' : 'bg-error']"></div>
                <span class="text-[10px] uppercase tracking-tighter">{{ config.isRunning ? t('server.start') : t('server.stop') }}</span>
                <span class="opacity-40 text-[10px] ml-0.5">({{ config.clients }})</span>
              </div>
            </div>
          </td>
          <td class="text-[11px] font-black opacity-40 font-mono tracking-tighter">{{ config.runtime }}</td>
          <td>
            <div class="flex items-center justify-center gap-1">
              <button v-if="config.protocol === 'HTTP' || config.protocol === 'HTTPS'"
                      class="btn btn-ghost btn-sm btn-square hover:bg-primary hover:text-primary-content transition-all"
                      @click="openWebConfig(config)">
                <div v-if="!config.isExistWeb" class="indicator">
                  <span class="indicator-item badge badge-warning badge-[8px] animate-bounce border-none"></span> 
                  <Icon icon="brook-web" style="font-size: 18px;" />
                </div>
                <Icon v-else icon="brook-web" style="font-size: 18px;" />
              </button>
              <div v-else class="w-8"></div>
              
              <button class="btn btn-ghost btn-sm btn-square hover:bg-info hover:text-info-content transition-all" @click="handleUpdate(config)">
                <Icon icon="brook-edit" style="font-size: 18px;" />
              </button>
              <button class="btn btn-ghost btn-sm btn-square hover:bg-error hover:text-error-content transition-all" @click="handleDelete(config.id)">
                <Icon icon="brook-delete" style="font-size: 18px;" />
              </button>
            </div>
          </td>
        </tr>

        <!-- 空状态提示 -->
        <tr v-if="configs.length === 0">
          <td colspan="11" class="text-center py-20 bg-base-100">
            <div class="flex flex-col items-center justify-center max-w-xs mx-auto">
              <div class="w-20 h-20 bg-base-200 rounded-3xl flex items-center justify-center mb-6 rotate-12 group-hover:rotate-0 transition-transform duration-500">
                <Icon icon="brook-technology_usb-cable" class="text-primary/20" style="font-size: 40px;"/>
              </div>
              <h3 class="text-lg font-black tracking-tight mb-2 opacity-80">{{ t('configuration.noConfigurations') }}</h3>
              <p class="text-xs font-medium opacity-40 leading-relaxed mb-8">
                {{ t('configuration.noConfigurationsDesc') }}
              </p>
              <button class="btn btn-primary btn-md gap-3 px-8 shadow-xl shadow-primary/20 font-black uppercase tracking-widest text-xs" @click="handleAdd">
                <Icon icon="brook-add" style="font-size: 18px;"/>
                {{ t('configuration.addConfiguration') }}
              </button>
            </div>
          </td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>
/* 表格行悬停效果 */
.table tbody tr:hover {
  background-color: hsl(var(--b2));
}

/* 操作按钮悬停效果 */
.btn:hover {
  transform: translateY(-1px);
  transition: transform 0.2s ease;
}

/* 状态切换动画 */
.toggle {
  transition: all 0.3s ease;
}

/* 强化表头磁吸效果 */
.table thead {
  position: sticky !important;
  top: 0 !important;
  z-index: 20 !important;
  background-color: hsl(var(--b1)) !important;
}

.table thead th {
  background-color: hsl(var(--b1)) !important;
  position: relative !important;
  border-bottom: 1px solid hsl(var(--bc) / 0.1) !important;
}
</style>
