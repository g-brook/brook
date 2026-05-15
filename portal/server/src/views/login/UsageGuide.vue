<script lang="ts" setup>
import Icon from '@/components/icon/Index.vue';
import {useI18n} from '@/components/lang/useI18n';

const { t } = useI18n();

const guideItems = [
  {
    key: 'token',
    order: '01',
    icon: 'brook-token',
    title: 'initializer.guide.tokenTitle',
    desc: 'initializer.guide.tokenDesc',
    path: 'initializer.guide.tokenPath',
  },
  {
    key: 'tunnel',
    order: '02',
    icon: 'brook-technology_usb-cable',
    title: 'initializer.guide.tunnelTitle',
    desc: 'initializer.guide.tunnelDesc',
    path: 'initializer.guide.tunnelPath',
  },
  {
    key: 'template',
    order: '03',
    icon: 'brook-empty',
    title: 'initializer.guide.templateTitle',
    desc: 'initializer.guide.templateDesc',
    path: 'initializer.guide.templatePath',
  },
];
</script>

<template>
  <div class="space-y-6 m-12">
    <div class="relative space-y-5">
      <div class="hidden lg:block absolute left-[38px] top-8 bottom-8 w-px bg-base-content/10"></div>

      <section
        v-for="item in guideItems"
        :key="item.key"
        class="relative grid grid-cols-1 lg:grid-cols-[76px_320px_1fr] gap-4 items-stretch"
      >
        <div class="hidden lg:flex justify-center pt-6">
          <div class="w-14 h-14 rounded-3xl bg-primary text-primary-content shadow-xl shadow-primary/20 flex items-center justify-center text-sm font-black z-10">
            {{ item.order }}
          </div>
        </div>

        <div class="rounded-3xl border border-base-content/5 bg-base-200/40 p-5 flex flex-col justify-between gap-5">
          <div>
            <div class="flex items-center gap-3 mb-4">
              <div class="w-11 h-11 rounded-2xl bg-primary/10 text-primary flex items-center justify-center lg:hidden">
                <Icon :icon="item.icon" style="font-size: 22px;" />
              </div>
              <div class="hidden lg:flex w-11 h-11 rounded-2xl bg-primary/10 text-primary items-center justify-center">
                <Icon :icon="item.icon" style="font-size: 22px;" />
              </div>
              <div>
                <div class="text-[11px] font-black opacity-35 tracking-[0.2em] uppercase">{{ t('initializer.guide.stepLabel', { step: item.order }) }}</div>
                <h4 class="text-xl font-black tracking-tight mt-0.5">{{ t(item.title) }}</h4>
              </div>
            </div>
            <p class="text-sm opacity-55 leading-relaxed">{{ t(item.desc) }}</p>
          </div>

          <div class="rounded-2xl bg-base-100 border border-base-content/5 p-3">
            <div class="text-[10px] font-black opacity-40 uppercase tracking-widest">{{ t('initializer.guide.pathLabel') }}</div>
            <div class="text-xs font-black text-primary mt-1 leading-relaxed">{{ t(item.path) }}</div>
          </div>
        </div>

        <div class="rounded-3xl border-2 border-dashed border-base-content/10 bg-base-100/70 min-h-[220px] flex items-center justify-center overflow-hidden shadow-sm">
          <div class="text-center px-8">
            <div class="w-16 h-16 rounded-3xl bg-base-200 mx-auto flex items-center justify-center mb-4">
              <Icon icon="brook-empty" class="opacity-20" style="font-size: 34px;" />
            </div>
            <div class="text-sm font-black opacity-40">{{ t('initializer.guide.imagePlaceholder') }}</div>
            <div class="text-xs opacity-30 mt-1">{{ t(item.title) }}</div>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
