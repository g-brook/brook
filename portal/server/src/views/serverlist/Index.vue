<script setup lang="ts">
import baseInfo from "@/service/baseInfo";
import {onMounted, ref} from "vue";
import DataInfo from "@/views/serverlist/DataInfo.vue";
import Icon from "@/components/icon/Index.vue";
import dayjs from 'dayjs'
import useI18n from '@/components/lang/useI18n'

const list = ref<any[]>([])

const selectRef = ref<any>()

const selectProxyId = ref<string>('')

const rightRef = ref(null)

const initData = async () => {
  const response = await baseInfo.getServerInfo({})
  list.value = response.data || []
  if (response.data.length > 0) {
    showHandler(list.value[0])
  }
}

const showHandler = (server) => {
  selectRef.value = server
  selectProxyId.value = server.proxyId
  rightRef.value?.refresh(server.proxyId)
}

onMounted(async () => {
  await initData()
})

const { t } = useI18n()

</script>

<template>
  <div v-if="list.length === 0" class="justify-center flex flex-col items-center">
    <div class="w-18 h-18 bg-base-300/59 rounded-full flex items-center justify-center mx-auto mb-4">
      <Icon icon="brook-Diagram-" class="text-base-content/30" style="font-size: 48px;"/>
    </div>
    <h3 class="text-lg font-medium text-base-content/30 mb-2">{{ t('server.listEmpty') }}</h3>
  </div>
  <div v-else class="flex flex-row items-center overflow-hidden 　justify-center h-full w-full">
    <div class="w-80 h-full">
      <div class="list bg-base-100 rounded-box shadow-md h-full overflow-hidden">
        <div
            class="card mx-3 my-1  card-xs shadow-sm hover:scale-105 hover:shadow-xl transition-all duration-300 group cursor-pointer"
            v-for="(server, index) in list" :key="index" @click="showHandler(server)"
            :class="server.proxyId==selectProxyId ?'bg-success text-base-100':'border-1 border-base-300'">
          <div class="card-body">
            <div class="card-title">
              <div>
                {{ server.tunnelType }}
              </div>
              {{ server.name }}
              <div class="badge badge-sm "
                   :class="server.proxyId==selectProxyId ?'badge-dash':'badge-primary badge-outline'">{{ server.tag }}
              </div>
            </div>
            <div class="flex flex-row justify-between pt-2">
              <div class="font-serif border-1 border-dashed rounded-lg px-2 py-2">
                {{ dayjs(server.runtime).format('YYYY-MM-DD HH:mm:ss') }}
              </div>
              <div class="flex flex-col items-center">
                <div class="badge badge-xs badge-soft">{{ t('server.fields.port') }}</div>
                <p class="font-bold">{{ server.port ? server.port : 'N/A' }}</p>
              </div>
              <div class="flex flex-col items-center">
                <div class="badge badge-xs badge-soft">{{ t('server.fields.connections') }}</div>
                {{ server.connections ? server.connections : '0' }}
              </div>
              <div class="flex flex-col items-center">
                <div class="badge badge-xs badge-soft">{{ t('server.fields.clients') }}</div>
                {{ server.users ? server.users : '0' }}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="flex-1 items-start justify-start h-full overflow-y-auto">
      <DataInfo ref="rightRef" :proxyId="selectProxyId"/>
    </div>
  </div>
</template>

<style scoped>

</style>