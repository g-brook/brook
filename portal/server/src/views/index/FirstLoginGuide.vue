<script setup lang="ts">
import {computed, ref} from 'vue';
import Icon from '@/components/icon/Index.vue';
import {useI18n} from '@/components/lang/useI18n';

const props = defineProps<{
  visible: boolean;
}>();

const emit = defineEmits<{
  (e: 'close'): void;
}>();

const {t} = useI18n();
const currentStep = ref(1);

const guideSteps = computed(() => [
  {
    step: 1,
    title: t('firstLoginGuide.step1Title'),
    desc: t('firstLoginGuide.step1Desc'),
    action: t('firstLoginGuide.step1Action'),
    icon: 'brook-token',
  },
  {
    step: 2,
    title: t('firstLoginGuide.step2Title'),
    desc: t('firstLoginGuide.step2Desc'),
    action: t('firstLoginGuide.step2Action'),
    icon: 'brook-technology_usb-cable',
  },
  {
    step: 3,
    title: t('firstLoginGuide.step3Title'),
    desc: t('firstLoginGuide.step3Desc'),
    action: t('firstLoginGuide.step3Action'),
    icon: 'brook-empty',
  },
]);

const activeStep = computed(() => guideSteps.value[currentStep.value - 1]);

const nextStep = () => {
  if (currentStep.value < guideSteps.value.length) {
    currentStep.value += 1;
  } else {
    emit('close');
  }
};

const prevStep = () => {
  if (currentStep.value > 1) {
    currentStep.value -= 1;
  }
};

const getStepImage = computed(() => {
  return new URL(`/src/assets/step${currentStep.value}.png`, import.meta.url).href
})

const skipGuide = () => emit('close');

const showPreview = ref(false)
const openPreview = () => (showPreview.value = true)
const closePreview = () => (showPreview.value = false)
</script>

<template>
  <Teleport to="body">
    <div v-if="visible" class="fixed inset-0 z-[100]">
      <div class="absolute inset-0 bg-base-300/60 backdrop-blur-sm" @click="skipGuide"></div>
      <div class="relative h-full w-full flex items-center justify-center p-4">
        <div class="w-full max-w-5xl rounded-3xl border border-base-content/10 bg-base-100 shadow-2xl overflow-hidden">
          <div class="px-6 py-4 border-b border-base-content/5 flex items-center justify-between">
            <div>
              <div class="text-[10px] font-black uppercase tracking-[0.3em] opacity-40">{{
                  t('firstLoginGuide.kicker')
                }}
              </div>
              <h3 class="text-xl font-black tracking-tight mt-1">{{ t('firstLoginGuide.title') }}</h3>
            </div>
            <button class="btn btn-ghost btn-sm btn-square" @click="skipGuide">
              <Icon icon="brook-delete" style="font-size: 16px;"/>
            </button>
          </div>

          <div class="grid grid-cols-1 lg:grid-cols-[280px_1fr] min-h-[520px]">
            <aside class="border-r border-base-content/5 bg-base-200/30 p-6">
              <div class="space-y-3">
                <div v-for="item in guideSteps" :key="item.step"
                     class="flex items-center gap-3 rounded-2xl p-3 transition-colors"
                     :class="currentStep === item.step ? 'bg-primary/10 border border-primary/10' : 'bg-base-100/60 border border-base-content/5'">
                  <div class="w-16 h-10 rounded-2xl flex items-center justify-center font-black"
                       :class="currentStep === item.step ? 'bg-primary text-primary-content' : 'bg-base-200 text-base-content/60'">
                    0{{ item.step }}
                  </div>
                  <div class="min-w-0">
                    <div class="text-sm font-black truncate">{{ item.title }}</div>
                    <div class="text-[11px] opacity-50 truncate">{{ item.action }}</div>
                  </div>
                </div>
              </div>
            </aside>

            <main class="p-6 lg:p-8 flex flex-col gap-6">
              <div class="flex items-start gap-4">
                <div class="w-14 h-14 rounded-3xl bg-primary/10 text-primary flex items-center justify-center shrink-0">
                  <Icon :icon="activeStep.icon" style="font-size: 28px;"/>
                </div>
                <div class="min-w-0">
                  <div class="text-[11px] font-black uppercase tracking-[0.25em] opacity-40">
                    {{ t('firstLoginGuide.stepLabel', {step: `0${currentStep}`}) }}
                  </div>
                  <h4 class="text-2xl font-black tracking-tight mt-1">{{ activeStep.title }}</h4>
                  <p class="text-sm opacity-60 leading-relaxed mt-3 max-w-2xl">{{ activeStep.desc }}</p>
                </div>
              </div>

              <div class="grid grid-cols-1 lg:grid-cols-[1.1fr_0.9fr] gap-5 items-stretch">
                <div
                    class="rounded-3xl border-2 border-dashed border-base-content/10 bg-base-200/40  flex items-center justify-center">
                  <div class="relative group">
                    <!-- 图片 -->
                    <img
                        :src="getStepImage"
                        class="w-full h-full rounded-3xl cursor-pointer"
                        @click="openPreview"
                    />

                    <div class="absolute left-1/2 -translate-x-1/2 bottom-3 bg-black/60 text-white text-xs px-3 py-1.5 rounded-full opacity-0 group-hover:opacity-100 transition-opacity duration-200 whitespace-nowrap">
                      {{t('firstLoginGuide.showImage')}}
                    </div>

                    <!-- 放大预览遮罩 -->
                    <div
                        v-if="showPreview"
                        class="fixed inset-0 bg-black/90 z-50 flex items-center justify-center"
                        @click="closePreview"
                    >
                      <img
                          :src="getStepImage"
                          class="max-w-[90%] max-h-[90vh] object-contain"
                          @click.stop
                      />
                    </div>
                  </div>
                </div>

                <div class="rounded-3xl bg-base-100 border border-base-content/5 p-5 flex flex-col gap-4">
                  <div class="rounded-2xl bg-base-200/40 p-4">
                    <div class="text-[10px] font-black uppercase tracking-widest opacity-40">
                      {{ t('firstLoginGuide.tipTitle') }}
                    </div>
                    <div class="text-sm font-black mt-2">{{ activeStep.action }}</div>
                  </div>
                  <div class="rounded-2xl bg-success/10 border border-success/10 p-4 text-sm leading-relaxed">
                    {{ t('firstLoginGuide.tipDesc') }}
                  </div>
                </div>
              </div>

              <div class="flex items-center justify-between gap-3 mt-auto pt-2">
                <button class="btn btn-ghost rounded-2xl px-6 font-black uppercase tracking-widest" @click="skipGuide">
                  {{ t('common.skip') }}
                </button>
                <div class="flex items-center gap-2">
                  <button class="btn btn-ghost rounded-2xl px-6 font-black uppercase tracking-widest"
                          :disabled="currentStep === 1" @click="prevStep">
                    {{ t('common.previous') }}
                  </button>
                  <button class="btn btn-primary rounded-2xl px-7 font-black uppercase tracking-widest"
                          @click="nextStep">
                    {{ currentStep < guideSteps.length ? t('common.next') : t('common.done') }}
                  </button>
                </div>
              </div>
            </main>
          </div>
        </div>
      </div>
    </div>
  </Teleport>
</template>
