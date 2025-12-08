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

<script setup lang="ts">

import baseInfo from "@/service/baseInfo";
import {onMounted, ref} from "vue";
import Icon from '@/components/icon/Index.vue';

interface Info {
  lastTime: string;
  host: string;

}

interface WebLog {
  protocol: string;
  path: string;
  host: string;
  method: string;
  status: int;
  proxyId: string;
  httpId: string;
  time: time.Time;
}

const props = defineProps({
  proxyId: {
    type: String,
    default: ""
  }
})

const configs = ref<Info[]>([]);
const webLogs = ref<WebLog[]>([]);

const getServerInfos = async () => {
  const response = await baseInfo.getServerInfoByProxyId({proxyId: props.proxyId});
  configs.value = response.data || []
}

const getWebLogInfos = async () => {
  const response = await baseInfo.getWebLogs({proxyId: props.proxyId});
  webLogs.value = response.data || []
}


onMounted(() => {
  getServerInfos();
})

</script>

<template>
  <div class="m-1">
    <!-- name of each tab group should be unique -->
    <div class="tabs tabs-border duration-75">
      <label class="tab">
        <input type="radio" name="my_tabs_4" checked="checked" @click="getServerInfos"/>
        <Icon icon="brook-client"/>
        客户端详情
      </label>
      <div class="tab-content bg-base-100 border-base-300">
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
            <td>{{ item.host }}</td>
            <td>{{ item.lastTime }}</td>
          </tr>
          </tbody>
        </table>
        　
      </div>

      <label class="tab">
        <input type="radio" name="my_tabs_4" @click="getWebLogInfos"/>
        <Icon icon="brook-a-clipboardnotedocument"/>
        Http日志
      </label>
      <div class="tab-content bg-base-100 border-base-300">
        <div class="fab">
          <button class="btn btn-lg btn-circle btn-primary opacity-80" @click="getWebLogInfos" >
            <Icon icon="brook-refresh" style="font-size: 20px"/>
          </button>
        </div>
        <table class="table ">
          <!-- head -->
          <thead class="sticky top-0 z-20 bg-base-100">
          <tr>
            <th class="bg-base-100 font-semibold" style="width: 10px">#
            </th>
            <th class="bg-base-100 font-semibold">时间
            </th>
            <th class="bg-base-100 font-semibold">Path
            </th>
            <th class="bg-base-100 font-semibold" style="width: 40px">HttpId
            </th>
            <th class="bg-base-100 font-semibold" style="width: 40px">协议
            </th>

            <th class="bg-base-100 font-semibold" style="width: 40px">Method
            </th>
            <th class="bg-base-100 font-semibold" style="width: 40px">状态
            </th>
          </tr>
          </thead>
          <tbody>
          <tr v-for="(item, index) in webLogs" :key="index">
            <th>
              <div class="flex items-center gap-2">
                <div class="badge badge-xs" :class="item.status==200 ? 'badge-success' : 'badge-error'">
                </div>
                {{ index + 1 }}
              </div>

            </th>
            <td>{{ item.time.String }}</td>
            <td>{{ item.path }}</td>
            <td>{{ item.httpId }}</td>
            <td>{{ item.protocol }}</td>
            <td>{{ item.method }}</td>
            <td>
              {{ item.status }}
            </td>
          </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>