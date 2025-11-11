<script setup lang="ts" xmlns="">
import baseInfo from "@/service/baseInfo";
import Icon from '@/components/icon/Index.vue';
import {onMounted, ref} from "vue";

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

const proxyId = ref<string>(props.proxyId);

const configs = ref<Info[]>([]);
const webLogs = ref<WebLog[]>([]);

const getServerInfos = async () => {
  const response = await baseInfo.getServerInfoByProxyId({proxyId: proxyId.value});
  configs.value = response.data || []
}

const getWebLogInfos = async () => {
  const response = await baseInfo.getWebLogs({proxyId: proxyId.value});
  webLogs.value = response.data || []
}


onMounted(() => {
  getServerInfos();
})

// 暴露方法给父组件
defineExpose({
  refresh: function (p) {
    proxyId.value = p ? p : props.proxyId;
    getServerInfos(p)
    getWebLogInfos(p);
  },
});
</script>

<template>
  <div class="ml-1" v-if="proxyId!==''">
    <!-- name of each tab group should be unique -->
    <div class="tabs tabs-border tabs-md duration-300 h-full">
      <label class="tab">
        <input type="radio" name="my_tabs_4" checked="checked" @click="getServerInfos"/>
        <Icon icon="brook-client"/>
        <p class="pl-1">端点(Agent)详情</p>
      </label>
      <div class="tab-content bg-base-100 border-base-300">
        <div v-if="configs.length > 0">
          <table class="table">
            <!-- head -->
            <thead class="sticky top-0 z-20 bg-base-100">
            <tr>
              <th class="bg-base-100 font-semibold" style="width: 2px">#
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

        <div class="flex  justify-center" v-else>
          没有数据
        </div>
        　
      </div>

      <label class="tab">
        <input type="radio" name="my_tabs_4" @click="getWebLogInfos"/>
        <Icon icon="brook-calendar"/>
        <p class="pl-1">HTTP日志</p>

      </label>
      <div class="tab-content bg-base-100 border-base-300">
        <div class="fab">
          <button class="btn btn-lg btn-circle btn-primary opacity-80" @click="getWebLogInfos">
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