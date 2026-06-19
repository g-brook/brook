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
import { computed, markRaw, onMounted, reactive, ref } from 'vue';
import * as dayjs from 'dayjs';
import Icon from '@/components/icon/Index.vue';
import Modal from '@/components/modal';
import FormPageShell from '@/components/page/FormPageShell.vue';
import configService from '@/service/config';
import { useI18n } from '@/components/lang/useI18n';
import Message from '@/components/message';
import type { IpRule, IpStrategy } from '@/types/ip';
import StrategyBindingsModalComponent from './StrategyBindingsModal.vue';

const StrategyBindingsModal = markRaw(StrategyBindingsModalComponent);
const { t } = useI18n();

type PageMode = 'list' | 'add' | 'edit';
type DraftRule = Partial<IpRule> & {
  _key: string;
  ip: string;
  remark: string;
};

const strategies = ref<IpStrategy[]>([]);
const rulesMap = ref<Record<number, IpRule[]>>({});
const rulesLoadingMap = ref<Record<number, boolean>>({});
const pageMode = ref<PageMode>('list');
const editingStrategy = ref<IpStrategy | null>(null);
const originalRules = ref<DraftRule[]>([]);
const draftRules = ref<DraftRule[]>([]);

const strategyForm = reactive({
  id: 0,
  name: '',
  type: 'WL',
  status: 1,
});

const totalStrategies = computed(() => strategies.value.length);
const activeStrategies = computed(() => strategies.value.filter(s => s.status === 1).length);
const isFormPage = computed(() => pageMode.value !== 'list');
const pageTitle = computed(() => pageMode.value === 'edit'
    ? t('menu.security.strategy.edit')
    : t('menu.security.strategy.add'));

const createRuleKey = () => `${Date.now()}-${Math.random().toString(16).slice(2)}`;

const toDraftRule = (rule?: Partial<IpRule>): DraftRule => ({
  id: rule?.id,
  strategyId: rule?.strategyId,
  ip: rule?.ip || '',
  remark: rule?.remark || '',
  _key: createRuleKey(),
});

const cloneRuleForCompare = (rule: DraftRule) => ({
  id: rule.id,
  ip: rule.ip.trim(),
  remark: (rule.remark || '').trim(),
});

const fail = (message?: string) => {
  Message.error(message || t('common.operationFailed'));
};

const request = async <T,>(
  runner: () => Promise<any>,
  options: {
    onSuccess?: (data: T) => void | Promise<void>;
    successMessage?: string;
    errorMessage?: string;
    logLabel: string;
  }
) => {
  try {
    const res = await runner();
    if (res?.success?.()) {
      if (options.onSuccess) {
        await options.onSuccess(res.data as T);
      }
      if (options.successMessage) {
        Message.success(options.successMessage);
      }
      return true;
    }
    fail(res?.message || options.errorMessage);
    return false;
  } catch (error) {
    console.error(options.logLabel, error);
    fail(options.errorMessage);
    return false;
  }
};

const fetchRulesForStrategy = async (strategyId: number) => {
  rulesLoadingMap.value[strategyId] = true;
  const ok = await request<IpRule[]>(
    () => configService.getIpRules(strategyId),
    {
      logLabel: 'getIpRules(map)',
      onSuccess: (data) => {
        rulesMap.value[strategyId] = data || [];
      },
    }
  );
  if (!ok) {
    rulesMap.value[strategyId] = [];
  }
  rulesLoadingMap.value[strategyId] = false;
};

const getStrategies = async () => {
  const ok = await request<IpStrategy[]>(
    () => configService.getAllStrategies(),
    {
      logLabel: 'getAllStrategies',
      onSuccess: async (data) => {
        strategies.value = data || [];
        await Promise.all((strategies.value || []).map(s => fetchRulesForStrategy(s.id)));
      },
    }
  );
  if (!ok) {
    strategies.value = [];
  }
};

onMounted(() => {
  getStrategies();
});

const resetForm = () => {
  strategyForm.id = 0;
  strategyForm.name = '';
  strategyForm.type = 'WL';
  strategyForm.status = 1;
  editingStrategy.value = null;
  originalRules.value = [];
  draftRules.value = [];
};

const handleAddStrategy = () => {
  resetForm();
  pageMode.value = 'add';
};

const handleUpdateStrategy = async (strategy: IpStrategy) => {
  resetForm();
  editingStrategy.value = strategy;
  strategyForm.id = strategy.id;
  strategyForm.name = strategy.name;
  strategyForm.type = strategy.type;
  strategyForm.status = strategy.status;
  pageMode.value = 'edit';

  const rules = rulesMap.value[strategy.id] || [];
  const draft = rules.map(rule => toDraftRule(rule));
  draftRules.value = draft;
  originalRules.value = draft.map(rule => ({ ...rule, _key: createRuleKey() }));
  if (!rulesMap.value[strategy.id]) {
    await fetchRulesForStrategy(strategy.id);
    const refreshed = (rulesMap.value[strategy.id] || []).map(rule => toDraftRule(rule));
    draftRules.value = refreshed;
    originalRules.value = refreshed.map(rule => ({ ...rule, _key: createRuleKey() }));
  }
};

const handleBackToList = () => {
  pageMode.value = 'list';
  resetForm();
};

const addRuleRow = () => {
  draftRules.value.push(toDraftRule());
};

const removeRuleRow = (index: number) => {
  draftRules.value.splice(index, 1);
};

const validateStrategyPage = () => {
  if (!strategyForm.name.trim()) {
    Message.error(t('validation.required'));
    return false;
  }
  const invalidRule = draftRules.value.find(rule => !rule.ip.trim());
  if (invalidRule) {
    Message.error(t('menu.security.rules.placeholder'));
    return false;
  }
  return true;
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

const findSavedStrategy = () => {
  if (strategyForm.id) {
    return strategies.value.find(item => item.id === strategyForm.id) || null;
  }
  const candidates = strategies.value
      .filter(item => item.name === strategyForm.name && item.type === strategyForm.type)
      .sort((a, b) => b.id - a.id);
  return candidates[0] || null;
};

const saveStrategyBase = async () => {
  const payload = {
    id: strategyForm.id,
    name: strategyForm.name.trim(),
    type: strategyForm.type,
    status: strategyForm.status,
  };
  const ok = await request<any>(
    () => pageMode.value === 'edit'
        ? configService.updateIpStrategy(payload)
        : configService.addIpStrategy(payload),
    {
      logLabel: pageMode.value === 'edit' ? 'updateIpStrategy' : 'addIpStrategy',
    }
  );
  if (!ok) return null;

  await getStrategies();
  const saved = findSavedStrategy();
  if (!saved) {
    fail(t('common.operationFailed'));
    return null;
  }
  strategyForm.id = saved.id;
  return saved;
};

const syncRules = async (strategyId: number) => {
  const current = draftRules.value.map(cloneRuleForCompare).filter(rule => rule.ip);
  const original = originalRules.value.map(cloneRuleForCompare).filter(rule => rule.ip);

  const currentById = new Map(current.filter(rule => rule.id).map(rule => [rule.id, rule]));
  const originalById = new Map(original.filter(rule => rule.id).map(rule => [rule.id, rule]));

  for (const rule of original) {
    const next = rule.id ? currentById.get(rule.id) : null;
    const changed = next && (next.ip !== rule.ip || next.remark !== rule.remark);
    if (rule.id && (!next || changed)) {
      const ok = await request<any>(
        () => configService.delIpRule(rule.id!),
        { logLabel: 'delIpRule(sync)' }
      );
      if (!ok) return false;
    }
  }

  for (const rule of current) {
    const old = rule.id ? originalById.get(rule.id) : null;
    const changed = old && (old.ip !== rule.ip || old.remark !== rule.remark);
    if (!rule.id || changed) {
      const ok = await request<any>(
        () => configService.addIpRule({
          strategyId,
          ip: rule.ip,
          remark: rule.remark,
        }),
        { logLabel: 'addIpRule(sync)' }
      );
      if (!ok) return false;
    }
  }

  return true;
};

const handleSavePage = async () => {
  if (!validateStrategyPage()) return;
  const saved = await saveStrategyBase();
  if (!saved) return;
  const rulesSaved = await syncRules(saved.id);
  if (!rulesSaved) return;
  Message.success(t('success.operationCompleted'));
  await getStrategies();
  handleBackToList();
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
        const bindings = Array.isArray(res.data) ? res.data : [];
        if (bindings.length > 0) {
          Modal.info(StrategyBindingsModal, {
            title: t('menu.security.strategy.deleteBlockedTitle'),
            props: { bindings },
            size: 'xl',
            closable: true,
            maskClosable: true,
          });
          return true as any;
        }
        Message.error(res.message || t('common.operationFailed'));
      } catch (error) {
        console.error('delIpStrategy', error);
        Message.error(t('common.operationFailed'));
      }
      return false as any;
    }
  });
};

const handleToggleStatus = async (strategy: IpStrategy) => {
  const prev = strategy.status;
  const newStatus = strategy.status === 1 ? 0 : 1;
  const ok = await request<any>(
    () => configService.updateIpStrategy({ ...strategy, status: newStatus }),
    {
      logLabel: 'updateIpStrategy(status)',
      successMessage: t('success.configurationUpdated'),
      onSuccess: async () => {
        await getStrategies();
      }
    }
  );
  if (!ok) {
    strategy.status = prev;
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
  <div class="brook-page">
    <FormPageShell
        v-if="isFormPage"
        :title="pageTitle"
        :cancelText="t('common.cancel')"
        :saveText="t('common.save')"
        @back="handleBackToList"
        @cancel="handleBackToList"
        @save="handleSavePage">
      <div class="mb-10">
        <h2 class="text-2xl font-semibold tracking-tight text-base-content/90">{{ pageTitle }}</h2>
        <p class="mt-3 text-sm font-medium text-base-content/45">
          {{ t('menu.security.description') }}
        </p>
      </div>

      <div class="overflow-hidden rounded-2xl border border-base-content/10 bg-base-100 shadow-[0_1px_2px_rgba(15,23,42,0.04)]">
        <section class="px-4 pb-3 pt-4">
          <div class="mb-3 flex items-center justify-between gap-4">
            <div>
              <h3 class="text-sm font-semibold tracking-tight text-base-content/80">{{ t('menu.security.strategy.title') }}</h3>
              <p class="mt-0.5 text-xs text-base-content/45">{{ t('menu.security.strategy.noStrategiesDesc') }}</p>
            </div>
            <input type="checkbox" v-model="strategyForm.status" :true-value="1" :false-value="0" class="toggle toggle-primary toggle-sm" />
          </div>

          <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
            <div class="form-control">
              <label class="mb-2 block text-sm font-semibold tracking-tight text-base-content/80">
                {{ t('menu.security.strategy.name') }} <span class="text-error">*</span>
              </label>
              <input
                  type="text"
                  v-model="strategyForm.name"
                  class="input input-bordered h-11 w-full rounded-xl border-base-content/10 bg-base-100 px-3 text-sm font-medium shadow-none focus:input-primary"
                  :placeholder="t('menu.security.strategy.name')" />
            </div>

            <div class="form-control">
              <label class="mb-2 block text-sm font-semibold tracking-tight text-base-content/80">
                {{ t('menu.security.strategy.type') }}
              </label>
              <select v-model="strategyForm.type" class="select select-bordered h-11 w-full rounded-xl border-base-content/10 bg-base-100 px-3 text-sm font-medium shadow-none focus:select-primary">
                <option value="WL">{{ t('menu.security.strategy.whitelist') }}</option>
                <option value="BL">{{ t('menu.security.strategy.blacklist') }}</option>
                <option value="IL">{{ t('menu.security.strategy.privateOnly') }}</option>
              </select>
              <div v-if="strategyForm.type === 'IL'" class="mt-2 rounded-xl border border-base-content/10 bg-base-200/40 px-3 py-2 text-xs font-semibold text-base-content/55">
                {{ t('menu.security.strategy.ilUsesWhitelistTip') }}
              </div>
            </div>
          </div>
        </section>

        <section class="border-t border-base-content/10 px-4 py-4">
          <div class="mb-4 flex items-center justify-between gap-4">
            <div>
              <h3 class="text-sm font-semibold tracking-tight text-base-content/80">{{ t('menu.security.rules.title') }}</h3>
              <p class="mt-0.5 text-xs text-base-content/45">{{ t('menu.security.rules.placeholder') }}</p>
            </div>
            <button type="button" class="btn btn-primary btn-sm h-9 rounded-xl gap-2 px-3 font-semibold" @click="addRuleRow">
              <Icon icon="brook-add" style="font-size: 14px;" />
              {{ t('menu.security.rules.add') }}
            </button>
          </div>

          <div class="brook-table-wrap">
            <table class="table table-md">
              <thead class="bg-base-200/50">
              <tr>
                <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.ip') }}</th>
                <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.rules.remark') }}</th>
                <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 80px">{{ t('configuration.actions') }}</th>
              </tr>
              </thead>
              <tbody>
              <tr v-for="(rule, index) in draftRules" :key="rule._key" class="hover:bg-base-200/30">
                <td>
                  <input
                      type="text"
                      v-model="rule.ip"
                      class="input input-bordered h-10 w-full rounded-xl border-base-content/10 bg-base-100 px-3 font-mono text-sm font-semibold shadow-none focus:input-primary"
                      :placeholder="t('menu.security.rules.placeholder')" />
                </td>
                <td>
                  <input
                      type="text"
                      v-model="rule.remark"
                      class="input input-bordered h-10 w-full rounded-xl border-base-content/10 bg-base-100 px-3 text-sm font-medium shadow-none focus:input-primary"
                      :placeholder="t('menu.security.rules.remarkPlaceholder')" />
                </td>
                <td class="text-center">
                  <button type="button" class="btn btn-ghost btn-sm btn-square hover:bg-error hover:text-error-content" @click="removeRuleRow(index)">
                    <Icon icon="brook-delete" style="font-size: 16px;" />
                  </button>
                </td>
              </tr>
              <tr v-if="draftRules.length === 0">
                <td colspan="3" class="py-12 text-center text-xs font-black uppercase tracking-widest opacity-25">
                  {{ t('pagination.noData') }}
                </td>
              </tr>
              </tbody>
            </table>
          </div>
        </section>
      </div>
    </FormPageShell>

    <template v-else>
      <!-- 极简页头：整合统计与操作 - 参考 Configuration.vue -->
      <div class="brook-toolbar">
        <div class="flex items-center gap-6">
          <div class="flex items-center gap-4">
            <div class="brook-stat group cursor-help">
              <div class="brook-stat-dot bg-primary"></div>
              <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('common.total') || 'Total' }}</span>
              <span class="text-sm font-black tracking-tighter">{{ totalStrategies }}</span>
            </div>
            <div class="brook-stat group cursor-help">
              <div class="brook-stat-dot bg-success"></div>
              <span class="text-xs font-black uppercase opacity-50 tracking-tighter">{{ t('configuration.enabled') || 'Enabled' }}</span>
              <span class="text-sm font-black tracking-tighter text-success">{{ activeStrategies }}</span>
            </div>
          </div>
        </div>

        <div class="flex items-center gap-1.5">
          <button class="btn btn-primary btn-xs brook-action-primary" @click="handleAddStrategy">
            <Icon icon="brook-add" style="font-size: 12px;"/>
            {{ t('common.add') }}
          </button>
          <div class="divider divider-horizontal mx-0.5 w-px h-4 self-center opacity-10"></div>
          <button class="btn btn-circle btn-xs btn-ghost brook-action-icon" @click="getStrategies">
            <Icon icon="brook-refresh" style="font-size: 14px;"/>
          </button>
        </div>
      </div>

      <div class="brook-table-wrap mx-1">
        <table class="table table-md">
          <thead class="bg-base-200/50">
          <tr>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center" style="width: 40px">#</th>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('menu.security.strategy.name') }}</th>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]" style="width: 160px">{{ t('menu.security.strategy.type') }}</th>
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
                <button class="btn btn-ghost btn-sm btn-square hover:bg-info hover:text-info-content transition-all" @click="handleUpdateStrategy(strategy)">
                  <Icon icon="brook-edit" style="font-size: 18px;" />
                </button>
                <button class="btn btn-ghost btn-sm btn-square hover:bg-error hover:text-error-content transition-all" @click="handleDeleteStrategy(strategy.id)">
                  <Icon icon="brook-delete" style="font-size: 18px;" />
                </button>
              </div>
            </td>
          </tr>

          <tr v-if="strategies.length === 0">
            <td colspan="8" class="text-center py-20 bg-base-100">
              <div class="flex flex-col items-center justify-center max-w-xs mx-auto">
                <div class="w-20 h-20 bg-base-200 rounded-3xl flex items-center justify-center mb-6 rotate-12 group-hover:rotate-0 transition-transform duration-500">
                  <Icon icon="brook-Firewall-Off" class="text-primary/20" style="font-size: 40px;"/>
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
    </template>
  </div>
</template>

<style scoped>
.btn:hover {
  transform: translateY(-1px);
  transition: transform 0.2s ease;
}

.toggle {
  transition: all 0.3s ease;
}
</style>
