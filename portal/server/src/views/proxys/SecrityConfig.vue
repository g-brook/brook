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
import { computed, ref, onMounted, markRaw } from 'vue';
import * as dayjs from 'dayjs';
import Icon from '@/components/icon/Index.vue';
import Drawer from '@/components/drawer/Index.vue';
import Modal from '@/components/modal';
import configService from '@/service/config';
import { useI18n } from '@/components/lang/useI18n';
import Message from '@/components/message';
import IpStrategyFormComponent from './IpStrategyForm.vue';

const IpStrategyForm = markRaw(IpStrategyFormComponent);

// 定义 IP 策略和规则的类型
interface IpStrategy {
  id: number;
  name: string;
  type: string;
  bind_handler: string;
  status: number;
  created_at: string;
  updated_at: string;
  tunnels?: any[];
}

interface IpRule {
  id: number;
  strategyId: number;
  ip: string;
  remark: string;
}

const { t } = useI18n();

// --- 数据 ---
const strategies = ref<IpStrategy[]>([]);

const currentStrategy = ref<IpStrategy | null>(null);
const rules = ref<IpRule[]>([]);
const rulesMap = ref<Record<number, IpRule[]>>({});
const rulesLoadingMap = ref<Record<number, boolean>>({});
const ruleDrawerRef = ref(null);

const newRuleIp = ref('');
const newRuleRemark = ref('');

const totalStrategies = computed(() => strategies.value.length);
const activeStrategies = computed(() => strategies.value.filter(s => s.status === 1).length);

const getStrategies = async () => {
  try {
    const res = await configService.getAllStrategies();
    if (res.success()) {
      strategies.value = res.data || [];
    } else {
      Message.error(res.message || t('common.operationFailed'));
      strategies.value = [];
    }
    await Promise.all((strategies.value || []).map(s => fetchRulesForStrategy(s.id)));
  } catch (error) {
    console.error('Failed to fetch strategies:', error);
    Message.error(t('common.operationFailed'));
  }
};

const getRules = async (strategyId: number) => {
  try {
    const res = await configService.getIpRules(strategyId);
    if (res.success()) {
      rules.value = res.data || [];
    } else {
      Message.error(res.message || t('common.operationFailed'));
      rules.value = [];
    }
  } catch (error) {
    console.error('Failed to fetch rules:', error);
    Message.error(t('common.operationFailed'));
  }
};

const fetchRulesForStrategy = async (strategyId: number) => {
  try {
    rulesLoadingMap.value[strategyId] = true;
    const res = await configService.getIpRules(strategyId);
    if (res.success()) {
      rulesMap.value[strategyId] = res.data || [];
    } else {
      rulesMap.value[strategyId] = [];
    }
  } catch (error) {
    rulesMap.value[strategyId] = [];
    console.error('Failed to fetch rules:', error);
  } finally {
    rulesLoadingMap.value[strategyId] = false;
  }
};

onMounted(() => {
  getStrategies();
});

const handleAddStrategy = () => {
  let formApi: { handleSubmit: () => Promise<any> } | null = null;
  Modal.open(IpStrategyForm, {
    title: t('menu.security.strategy.add'),
    size: 'auto',
    closable: true,
    maskClosable: false,
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
            const res = await configService.addIpStrategy(formData);
            if (res.success()) {
              Message.success(t('success.operationCompleted'));
              await getStrategies();
              return true;
            }
            Message.error(res.message || t('common.operationFailed'));
          }
        } catch (error) {
          console.error('Failed to add strategy:', error);
          Message.error(t('common.operationFailed'));
        }
      }
      return false;
    },
  });
};

const handleUpdateStrategy = (strategy: IpStrategy) => {
  let formApi: { handleSubmit: () => Promise<any> } | null = null;
  Modal.open(IpStrategyForm, {
    title: t('menu.security.strategy.edit'),
    size: 'auto',
    closable: true,
    maskClosable: false,
    showFooter: true,
    props: {
      onRegister: (api) => {
        formApi = api;
      },
      initialData: strategy,
      isEdit: true,
    },
    onConfirm: async () => {
      if (formApi) {
        try {
          const formData = await formApi.handleSubmit();
          if (formData) {
            const res = await configService.updateIpStrategy(formData);
            if (res.success()) {
              Message.success(t('success.operationCompleted'));
              await getStrategies();
              return true;
            }
            Message.error(res.message || t('common.operationFailed'));
          }
        } catch (error) {
          console.error('Failed to update strategy:', error);
          Message.error(t('common.operationFailed'));
        }
      }
      return false;
    },
  });
};

const handleDeleteStrategy = (id: number) => {
  Modal.confirm({
    title: t('menu.security.strategy.delete'),
    onConfirm: async () => {
      try {
        const res = await configService.delIpStrategy(id);
        if (res.success()) {
          Message.success(t('success.operationCompleted'));
          await getStrategies();
          return true as any;
        }
        Message.error(res.message || t('common.operationFailed'));
      } catch (error) {
        console.error('Failed to delete strategy:', error);
        Message.error(t('common.operationFailed'));
      }
      return false as any;
    }
  });
};

const handleToggleStatus = async (strategy: IpStrategy) => {
  const prev = strategy.status;
  try {
    const newStatus = strategy.status === 1 ? 0 : 1;
    const res = await configService.updateIpStrategy({ ...strategy, status: newStatus });
    if (res.success()) {
      Message.success(t('success.configurationUpdated'));
      await getStrategies();
      return;
    }
    Message.error(res.message || t('common.operationFailed'));
    strategy.status = prev;
  } catch (error) {
    console.error('Failed to toggle status:', error);
    Message.error(t('common.operationFailed'));
    strategy.status = prev;
  }
};

const openRuleManager = (strategy: IpStrategy) => {
  currentStrategy.value = strategy;
  getRules(strategy.id);
  ruleDrawerRef.value?.open();
};

const handleAddRule = async () => {
  if (!newRuleIp.value) {
    Message.error(t('validation.required'));
    return;
  }
  if (!currentStrategy.value) return;

  try {
    const res = await configService.addIpRule({
      strategyId: currentStrategy.value.id,
      ip: newRuleIp.value,
      remark: newRuleRemark.value
    });
    if (res.success()) {
      Message.success(t('success.operationCompleted'));
      newRuleIp.value = '';
      newRuleRemark.value = '';
      getRules(currentStrategy.value.id);
      fetchRulesForStrategy(currentStrategy.value.id);
      return;
    }
    Message.error(res.message || t('common.operationFailed'));
  } catch (error) {
    console.error('Failed to add rule:', error);
    Message.error(t('common.operationFailed'));
  }
};

const handleDeleteRule = async (id: number) => {
  if (!currentStrategy.value) return;
  try {
    const res = await configService.delIpRule(id);
    if (res.success()) {
      Message.success(t('success.operationCompleted'));
      getRules(currentStrategy.value.id);
      fetchRulesForStrategy(currentStrategy.value.id);
      return;
    }
    Message.error(res.message || t('common.operationFailed'));
  } catch (error) {
    console.error('Failed to delete rule:', error);
    Message.error(t('common.operationFailed'));
  }
};

const getStrategyTypeBadge = (type: string) => {
  switch (type) {
    case 'WL': return 'badge-success';
    case 'BL': return 'badge-error';
    case 'IL': return 'badge-info';
    default: return 'badge-ghost';
  }
};

const getStrategyTypeText = (type: string) => {
  switch (type) {
    case 'WL': return t('menu.security.strategy.whitelist');
    case 'BL': return t('menu.security.strategy.blacklist');
    case 'IL': return t('menu.security.strategy.privateOnly');
    default: return 'Unknown';
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
</script>

<template>
  <div class="overflow-hidden">
    <!-- IP 规则管理抽屉 -->
    <Drawer ref="ruleDrawerRef" :title="`${t('menu.security.rules.title')} - ${currentStrategy?.name}`" icon="brook-security" width="50%">
      <div class="p-6 flex flex-col h-full space-y-6">
        <!-- 添加规则表单 - 参考 ConfigForm 风格 -->
        <div class="bg-base-200/40 rounded-3xl p-5 border border-base-content/5 space-y-4">
          <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.rules.ip') }}</span>
              </label>
              <input type="text" v-model="newRuleIp" 
                     class="input input-bordered focus:input-primary w-full h-11 bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm font-black text-sm tracking-tight border-base-content/5" 
                     :placeholder="t('menu.security.rules.placeholder')" />
            </div>
            <div class="form-control">
              <label class="label py-1">
                <span class="label-text font-black text-[11px] opacity-40 uppercase tracking-[0.15em]">{{ t('menu.security.rules.remark') }}</span>
              </label>
              <input type="text" v-model="newRuleRemark" 
                     class="input input-bordered focus:input-primary w-full h-11 bg-base-100/30 hover:bg-base-100/50 focus:bg-base-100 transition-all shadow-sm font-black text-sm tracking-tight border-base-content/5" 
                     :placeholder="t('menu.security.rules.remarkPlaceholder')" />
            </div>
          </div>
          <button class="btn btn-primary btn-sm h-11 w-full gap-2 font-black uppercase tracking-widest shadow-md shadow-primary/20" @click="handleAddRule">
            <Icon icon="brook-add" style="font-size: 16px;" />
            {{ t('menu.security.rules.add') }}
          </button>
        </div>

        <!-- 规则列表 - 参考 Configuration 风格 -->
        <div class="flex-1 overflow-hidden flex flex-col rounded-3xl border border-base-content/5 bg-base-100 shadow-sm">
          <div class="overflow-y-auto flex-1">
            <table class="table table-md table-pin-rows">
              <thead class="bg-base-200/50">
                <tr>
                  <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.ip') }}</th>
                  <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.remark') }}</th>
                  <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 80px">{{ t('configuration.actions') }}</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="rule in rules" :key="rule.id" class="hover:bg-base-200/40 transition-colors group">
                  <td class="font-mono font-black text-sm text-primary tracking-tight">{{ rule.ip }}</td>
                  <td class="text-xs font-black opacity-40 tracking-tight">{{ rule.remark }}</td>
                  <td class="text-center">
                    <button class="btn btn-ghost btn-xs btn-square hover:bg-error hover:text-error-content transition-all" @click="handleDeleteRule(rule.id)">
                      <Icon icon="brook-delete" style="font-size: 16px;" />
                    </button>
                  </td>
                </tr>
                <tr v-if="rules.length === 0">
                  <td colspan="3" class="text-center py-20 opacity-30">
                    <div class="flex flex-col items-center gap-2">
                      <Icon icon="brook-empty" style="font-size: 40px;" />
                      <span class="text-xs font-black uppercase tracking-widest">{{ t('pagination.noData') }}</span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </Drawer>

    <!-- 极简页头：整合统计与操作 - 参考 Configuration.vue -->
    <div class="flex sticky top-0 items-center h-14 justify-between gap-4 mb-3 px-5 py-2 rounded-xl bg-base-100/80 backdrop-blur-md z-30 border border-base-content/5 shadow-sm mx-1">
      <div class="flex items-center gap-6">
        <!-- 整合后的微缩统计 -->
        <div class="flex items-center gap-4">
          <div class="flex items-center gap-1.5 group cursor-help">
            <div class="w-1.5 h-1.5 rounded-full bg-primary opacity-40"></div>
            <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('common.total') || 'Total' }}</span>
            <span class="text-sm font-black tracking-tighter">{{ totalStrategies }}</span>
          </div>
          <div class="flex items-center gap-1.5 group cursor-help">
            <div class="w-1.5 h-1.5 rounded-full bg-success opacity-40"></div>
            <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('configuration.enabled') || 'Enabled' }}</span>
            <span class="text-sm font-black tracking-tighter text-success">{{ activeStrategies }}</span>
          </div>
        </div>
      </div>

      <div class="flex items-center gap-1.5">
        <button class="btn btn-primary btn-xs h-8 gap-1.5 font-bold px-3 shadow-md shadow-primary/20 text-xs uppercase tracking-widest" @click="handleAddStrategy">
          <Icon icon="brook-add" style="font-size: 12px;"/>
          {{ t('common.add') }}
        </button>
        <div class="divider divider-horizontal mx-0.5 w-px h-4 self-center opacity-10"></div>
        <button class="btn btn-circle btn-xs h-8 w-8 btn-ghost hover:rotate-180 transition-transform duration-500" @click="getStrategies">
          <Icon icon="brook-refresh" style="font-size: 14px;"/>
        </button>
      </div>
    </div>

    <!-- 配置表格 - 参考 Configuration.vue 风格 -->
    <div class="overflow-x-auto rounded-3xl border border-base-content/5 bg-base-100 shadow-sm mx-1">
      <table class="table table-md">
        <!-- head -->
        <thead class="bg-base-200/50">
        <tr>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 40px">#</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.strategy.name') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 160px">{{ t('menu.security.strategy.type') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 200px">{{ t('menu.security.strategy.bindHandler') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 320px">{{ t('menu.security.rules.title') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 180px">{{ t('menu.security.strategy.createdAt') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 180px">{{ t('menu.security.strategy.updatedAt') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 140px">{{ t('menu.security.strategy.status') }}</th>
          <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 150px">{{ t('configuration.actions') }}</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(strategy, index) in strategies" :key="strategy.id" class="hover:bg-base-200/40 transition-colors group">
          <th class="text-center opacity-30 font-mono text-xs">{{ index + 1 }}</th>
          <td>
            <div class="flex flex-col gap-0.5">
              <div class="flex items-center gap-2">
                <div class="relative flex items-center justify-center">
                  <div v-if="strategy.status === 1" class="absolute w-2 h-2 bg-success rounded-full animate-ping opacity-75"></div>
                  <div :class="['w-2 h-2 rounded-full relative z-10', strategy.status === 1 ? 'bg-success' : 'bg-base-300']"></div>
                </div>
                <span class="font-black text-sm tracking-tight">{{ strategy.name }}</span>
              </div>
              <!-- 绑定的隧道统计 -->
              <div v-if="strategy.tunnels?.length" class="flex items-center gap-1 opacity-40 scale-95 origin-left">
                <Icon icon="brook-technology_usb-cable" style="font-size: 10px;" />
                <span class="text-[12px] font-black uppercase tracking-widest">{{ strategy.tunnels.length }} Tunnels</span>
              </div>
            </div>
          </td>
          <td>
            <div :class="['badge badge-soft flex items-center gap-1.5 w-fit px-3 py-2.5 border border-current/5', getStrategyTypeBadge(strategy.type)]">
              <span class="font-black text-[12px] tracking-widest uppercase">{{ getStrategyTypeText(strategy.type) }}</span>
            </div>
          </td>
          <td>
            <div class="flex items-center gap-2">
              <span class="text-[12px] font-black opacity-20 uppercase tracking-tighter">Handler:</span>
              <code class="text-xs font-black tracking-tight opacity-70">{{ strategy.bind_handler }}</code>
            </div>
          </td>
          <td>
            <div class="flex items-start justify-between gap-2">
              <div class="flex flex-col gap-1 flex-1 min-w-0">
                <div v-if="rulesLoadingMap[strategy.id]" class="text-xs font-black opacity-30">
                  {{ t('common.loading') || 'Loading' }}
                </div>
                <div v-else-if="(rulesMap[strategy.id] || []).length === 0" class="text-xs font-black opacity-20">
                  {{ t('pagination.noData') }}
                </div>
                <div v-else class="flex flex-wrap gap-1.5">
                  <div v-for="rule in (rulesMap[strategy.id] || [])" :key="rule.id" class="tooltip" :data-tip="rule.remark || rule.ip">
                    <span class="badge badge-ghost badge-sm font-mono font-black opacity-70 max-w-[280px] truncate">
                      {{ rule.ip }}
                    </span>
                  </div>
                </div>
              </div>
              <button class="btn btn-ghost btn-xs btn-square hover:bg-base-200 transition-all mt-0.5" @click="fetchRulesForStrategy(strategy.id)">
                <Icon icon="brook-refresh" style="font-size: 14px;"/>
              </button>
            </div>
          </td>
          <td class="font-mono text-xs font-black opacity-50">{{ formatTime(strategy.created_at) }}</td>
          <td class="font-mono text-xs font-black opacity-50">{{ formatTime(strategy.updated_at) }}</td>
          <td>
            <div class="flex items-center gap-2">
              <input type="checkbox" class="toggle toggle-primary toggle-sm"
                     :checked="strategy.status === 1" @change="handleToggleStatus(strategy)"/>
              <span :class="['text-[12px] font-black uppercase tracking-[0.1em]', strategy.status === 1 ? 'text-primary' : 'opacity-20']">
                {{ strategy.status === 1 ? t('configuration.enabled') : t('configuration.disabled') }}
              </span>
            </div>
          </td>
          <td>
            <div class="flex items-center justify-center gap-1">
              <button class="btn btn-ghost btn-sm btn-square hover:bg-primary hover:text-primary-content transition-all" @click="openRuleManager(strategy)">
                <Icon icon="brook-web" style="font-size: 18px;" />
              </button>
              <button class="btn btn-ghost btn-sm btn-square hover:bg-info hover:text-info-content transition-all" @click="handleUpdateStrategy(strategy)">
                <Icon icon="brook-edit" style="font-size: 18px;" />
              </button>
              <button class="btn btn-ghost btn-sm btn-square hover:bg-error hover:text-error-content transition-all" @click="handleDeleteStrategy(strategy.id)">
                <Icon icon="brook-delete" style="font-size: 18px;" />
              </button>
            </div>
          </td>
        </tr>

        <!-- 空状态提示 -->
        <tr v-if="strategies.length === 0">
          <td colspan="9" class="text-center py-20 bg-base-100">
            <div class="flex flex-col items-center justify-center max-w-xs mx-auto">
              <div class="w-20 h-20 bg-base-200 rounded-3xl flex items-center justify-center mb-6 rotate-12 group-hover:rotate-0 transition-transform duration-500">
                <Icon icon="brook-security" class="text-primary/20" style="font-size: 40px;"/>
              </div>
              <h3 class="text-lg font-black tracking-tight mb-2 opacity-80">{{ t('menu.security.strategy.noStrategies') }}</h3>
              <p class="text-xs font-medium opacity-40 leading-relaxed mb-8">
                {{ t('menu.security.strategy.noStrategiesDesc') }}
              </p>
              <button class="btn btn-primary btn-md gap-3 px-8 shadow-xl shadow-primary/20 font-black uppercase tracking-widest text-xs" @click="handleAddStrategy">
                <Icon icon="brook-add" style="font-size: 18px;"/>
                {{ t('menu.security.strategy.add') }}
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
