<script setup lang="ts">
import useI18n from '@/components/lang/useI18n'
const { t } = useI18n()
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center gap-3">
      <div class="w-8 h-8 bg-base-200 rounded-md flex items-center justify-center">
        <i class="iconfont brook-token" style="font-size: 24px"></i>
      </div>
      <div>
        <h2 class="text-base-content text-2xl">{{ t('mysetting.tls.title') }}</h2>
        <p class="text-xs text-base-content/60">{{ t('mysetting.tls.subtitle') }}</p>
      </div>
    </div>
    <div class="border border-base-300 rounded-lg  w-full">
      <div class="mt-4 flex justify-end p-2">
        <button class="btn btn-primary">
          <i class="iconfont brook-plus"></i>
          {{ t('mysetting.tls.actions.create') }}
        </button>
      </div>
      <div class="bg-base-100 rounded-lg p-4">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
              <tr>
                <th>{{ t('mysetting.tls.table.certName') }}</th>
                <th>{{ t('mysetting.tls.table.certStatus') }}</th>
                <th>{{ t('mysetting.tls.table.certExpireTime') }}</th>
                <th>{{ t('mysetting.tls.table.certActions') }}</th>
              </tr>
            </thead>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>