<script setup lang="ts">

import baseInfo from "@/service/baseInfo";
import config from "@/service/config";
import {ref} from "vue";

interface Info {
  lastTime: string;
  host: string;

}

const props = defineProps({
  proxyId: {
    type: String,
    default: ""
  }
})

const configs = ref<Info[]>([]);

const getServerInfos = async () => {
  const response = await baseInfo.getServerInfoByProxyId({proxyId: props.proxyId});
  configs.value = response.data || []
}

</script>

<template>
  <div class="m-1">
  <!-- name of each tab group should be unique -->
  <div class="tabs tabs-border duration-75">
    <label class="tab">
      <input type="radio" name="my_tabs_4" checked="checked" @click="getServerInfos" />
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4 me-2"><path stroke-linecap="round" stroke-linejoin="round" d="M5.25 5.653c0-.856.917-1.398 1.667-.986l11.54 6.347a1.125 1.125 0 0 1 0 1.972l-11.54 6.347a1.125 1.125 0 0 1-1.667-.986V5.653Z" /></svg>
      客户端详情
    </label>
    <div class="tab-content bg-base-100 border-base-300 p-6">
      <table class="table">
        <!-- head -->
        <thead class="sticky top-0 z-20 bg-base-100">
        <tr>
          <th class="bg-base-100 font-semibold" style="width: 10px">#
          </th>
          <th class="bg-base-100 font-semibold" style="width: 80px">地址
          </th>
          <th class="bg-base-100 font-semibold" style="width: 80px">连接时间
          </th>
        </tr>
        </thead>
        <tbody>
        <tr v-for="(item, index) in configs" :key="index">
          <th>
            <div class="flex items-center gap-2">
              {{ index + 1 }}
            </div>

          </th>
          <td>{{item.host}}</td>
          <td>{{item.lastTime}}</td>
        </tr>
        </tbody>
      </table>
      　
    </div>

    <label class="tab">
      <input type="radio" name="my_tabs_4"/>
      <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-4 me-2"><path stroke-linecap="round" stroke-linejoin="round" d="M15.182 15.182a4.5 4.5 0 0 1-6.364 0M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0ZM9.75 9.75c0 .414-.168.75-.375.75S9 10.164 9 9.75 9.168 9 9.375 9s.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Zm5.625 0c0 .414-.168.75-.375.75s-.375-.336-.375-.75.168-.75.375-.75.375.336.375.75Zm-.375 0h.008v.015h-.008V9.75Z" /></svg>
      日志
    </label>
    <div class="tab-content bg-base-100 border-base-300 p-6">Tab content 2</div>
  </div>
  </div>
</template>

<style scoped>

</style>