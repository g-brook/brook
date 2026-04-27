<script lang="ts" setup>
import { useI18n } from '@/components/lang/useI18n';

defineProps<{
  bindings: Array<{
    id: number;
    name: string;
    proxyId: string;
    remotePort: number;
    protocol: string;
  }>;
}>();

const { t } = useI18n();
</script>

<template>
  <div class="p-6 space-y-5 min-w-[48rem]">
    <div class="text-sm font-black opacity-70">
      {{ t('menu.security.strategy.deleteBlockedDesc') }}
    </div>
    <div class="overflow-x-auto rounded-2xl border border-base-content/5 bg-base-100">
      <table class="table table-sm">
        <thead class="bg-base-200/40">
        <tr>
          <th class="font-black text-xs uppercase opacity-60 tracking-widest">#</th>
          <th class="font-black text-xs uppercase opacity-60 tracking-widest">{{ t('configuration.nameAndTag') }}</th>
          <th class="font-black text-xs uppercase opacity-60 tracking-widest">{{ t('configuration.proxyId') }}</th>
          <th class="font-black text-xs uppercase opacity-60 tracking-widest">{{ t('configuration.remotePort') }}</th>
          <th class="font-black text-xs uppercase opacity-60 tracking-widest">{{ t('configuration.protocol') }}</th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(b, idx) in bindings" :key="b.id" class="hover:bg-base-200/40 transition-colors">
          <td class="font-mono text-xs opacity-50">{{ idx + 1 }}</td>
          <td class="text-xs font-black">{{ b.name }}</td>
          <td class="text-xs font-mono font-black opacity-70">{{ b.proxyId }}</td>
          <td class="text-xs font-mono font-black opacity-70">{{ b.remotePort }}</td>
          <td class="text-xs font-black opacity-60">{{ b.protocol }}</td>
        </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
