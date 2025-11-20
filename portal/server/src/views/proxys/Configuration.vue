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
import {computed, markRaw, onMounted, ref} from 'vue';
import Icon from '@/components/icon/Index.vue';
import Drawer from '@/components/drawer/Index.vue';
import Modal from '@/components/modal';
import ConfigFormComponent from './ConfigForm.vue';
import config from '@/service/config';
import {useI18n} from '@/components/lang/useI18n';
import WebConfiguration from "./WebConfiguration.vue";
import Message from '@/components/message';
import DownloadConfig from "./DownloadConfig.vue";

// 定义配置项的类型
interface ConfigItem {
  id: number;
  name: string;
  tag: string;
  remotePort: number;
  proxyId: string;
  protocol: string;
  state: boolean;
  isRunning: boolean;
  runtime: string;
  isExistWeb: boolean;
  clients: number;
  destination: string;
}

const {t} = useI18n();
const ConfigForm = markRaw(ConfigFormComponent);

const configs = ref<ConfigItem[]>([]);
const webItem = ref<ConfigItem>(null);

// 计算统计信息
const totalConfigs = computed(() => configs.value.length);
const enabledConfigs = computed(() => configs.value.filter(item => item.state).length);
const runningConfigs = computed(() => configs.value.filter(item => item.isRunning).length);
const drawerRef = ref<{ open: () => void } | null>(null);
const downloadDrawerRef = ref<{ open: () => void } | null>(null);

const getConfigs = async () => {
  try {
    const res = await config.getProxyConfigs();
    configs.value = res.data || [];
  } catch (error) {
    console.error(error);
  }
};

// 组件挂载时获取数据
onMounted(() => {
  getConfigs();
});

const openWebConfig = (item: ConfigItem) => {
  if (item.protocol !== `HTTPS` && item.protocol !== `HTTP`) {
    return
  }
  webItem.value = item;
  drawerRef.value?.open();
}

const handleAdd = () => {
  let formApi: { handleSubmit: () => Promise<boolean> } | null = null;
  Modal.open(ConfigForm, {
    title: t('configuration.addTunnelConfig'),
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
            await getConfigs();
            return true
          }
        } catch (error) {
          console.error('Failed to submit form:', error);
        }
        return false
      }
    },
  });
};


const handleDelete = (id: number) => {
  Modal.confirm({
    onConfirm: async () => {
      config.delProxyConfig(id).then(e => {
        if (e.success()) {
          getConfigs()
          message.info(t("common.success"));
        }
      })
      return true
    }
  });
};

const handleUpdate = (cfg) => {
  let formApi: { handleSubmit: () => Promise<boolean> } | null = null;
  Modal.open(ConfigForm, {
    title: t('configuration.editTunnelConfig'),
    size: 'auto',
    closable: true,
    maskClosable: false,
    showFooter: true,
    props: {
      onRegister: (api) => {
        formApi = api;
      },
      initialData: cfg,
      isEdit: true,
    },
    onConfirm: async () => {
      if (formApi) {
        try {
          const formData = await formApi.handleSubmit();
          if (formData) {
            await getConfigs();
            return true
          }
        } catch (error) {
          console.error('Failed to submit form:', error);
        }
        return false
      }
    },
  });
};

const handleToggleStatus = async (id: number, state: boolean) => {
  try {
    const res = await config.updateProxyState({
      id: id,
      state: state ? 0 : 1
    });
    if (res.success()) {
      Message.info(t('success.configurationUpdated'))
      getConfigs()
    }
  } catch (error) {
  }
};
const handleClickDownload = () => {
  downloadDrawerRef.value?.open();
}
</script>

<template>
  <div class="overflow-hidden">
    <Drawer ref="drawerRef" :title="t('configuration.webInfoConfig')" icon="brook-web" width="50%">
      <WebConfiguration :refProxyId="webItem?.id" :protocol="webItem?.protocol"/>
    </Drawer>
    <Drawer ref="downloadDrawerRef" :title="t('configuration.template')" icon="brook-empty" width="50%">
      <DownloadConfig/>
    </Drawer>
    <!-- 操作栏 -->
    <div
        class="flex sticky border-t-1 border-base-300/50 top-0 items-center h-auto min-h-12 justify-between gap-2 mb-1 p-2 rounded-b-2xl bg-base-100/85 backdrop-blur-sm z-0 shadow-xs">
      <!-- 左侧信息 -->
      <div class="flex items-center gap-4">
        <div class="text-sm text-base-content/60">
          {{ t('configuration.totalConfigs', {count: totalConfigs}) }}
        </div>
        <div class="text-sm text-base-content/60">
          {{ t('configuration.enabledConfigs', {count: enabledConfigs}) }}
        </div>
        <div class="text-sm text-base-content/60">
          {{ t('configuration.runningConfigs', {count: runningConfigs}) }}
        </div>
      </div>

      <!-- 右侧操作按钮 -->
      <div class="flex items-center">
        <button class="btn  btn-ghost btn-sm" @click="handleClickDownload">
          <Icon icon="brook-empty" style="font-size: 12px;"/>
          {{ t('configuration.template') }}
        </button>
        <button class="btn  btn-ghost btn-sm" @click="handleAdd">
          <Icon icon="brook-add" style="font-size: 12px;"/>
          {{ t('common.add') }}
        </button>

        <button class="btn btn-circle " @click="getConfigs">
          <Icon icon="brook-refresh"/>
        </button>
      </div>
    </div>

    <!-- 配置表格 -->
    <div class="overflow-x-auto rounded-box border border-base-content/5 bg-base-100 h-full overflow-y-auto mx-1">
      <table class="table">
        <!-- head -->
        <thead class="sticky top-0 z-20 bg-base-100">
        <tr>
          <th class="bg-base-100 font-semibold" style="width: 80px">{{ t('configuration.serialNumber') }}
          </th>
          <th class="bg-base-100 font-semibold">{{ t('configuration.nameAndTag') }}</th>
          <th class="bg-base-100 font-semibold" style="width: 100px">{{ t('configuration.remotePort') }}
          </th>
          <th class="bg-base-100 font-semibold" style="width: 140px">{{ t('configuration.destination') }}(IP:PORT)</th>
          <th class="bg-base-100 font-semibold" style="width: 140px">{{ t('configuration.proxyId') }}</th>
          <th class="bg-base-100 font-semibold" style="width: 100px">{{ t('configuration.protocol') }}
          </th>
          <th class="bg-base-100 font-semibold" style="width: 100px">{{ t('configuration.status') }}</th>
          <th class="bg-base-100 font-semibold" style="width: 180px">{{ t('server.runtime') }}</th>
          <th class="bg-base-100 font-semibold" style="width: 140px">{{ t('common.running') }}</th>
          <th class="bg-base-100 font-semibold" style="width: 240px">{{ t('configuration.actions') }}</th>
        </tr>
        </thead>
        <tbody>
        <!-- 动态渲染配置数据 -->
        <tr v-for="(config, index) in configs" :key="config.id" class="hover">
          <th>
            <div class="flex items-center gap-2">
              {{ index + 1 }}
            </div>

          </th>
          <td>
            <div class="flex items-center">

              <div class="text-sm font-bold">
                <div class="inline-grid *:[grid-area:1/1]" v-if="config.isRunning">
                  <div class="status status-success animate-ping"></div>
                  <div class="status status-success"></div>
                </div>
                <div class="status status-error" v-else/>
                {{ config.name }}
              </div>
              <div class="ml-2">
                                    <span class="badge badge-xs"
                                          :class="config.isRunning ? 'badge-success' : 'badge-warning'">{{
                                        config.tag
                                      }} </span>
              </div>
              　
            </div>
          </td>
          <td>{{ config.remotePort }}</td>
          <td>{{ config.destination }}</td>
          <td>{{ config.proxyId }}</td>
          <td>
            <div class="badge badge-soft badge-secondary w-16">{{ config.protocol }}</div>
          </td>
          <td>
            <div class="form-control">
              <label class="cursor-pointer label gap-2">
                <input type="checkbox" class="toggle toggle-primary toggle-sm "
                       :checked="config.state" @change="handleToggleStatus(config.id, config.state)"/>
                <span class="label-text text-xs">
                                        {{ config.state ? t('configuration.enabled') : t('configuration.disabled') }}
                                    </span>
              </label>
            </div>
          </td>
          <td>
                            <span>
                                {{ config.runtime }}
                            </span>
          </td>
          <td>
            <div class="badge text-xs text-base-100"
                 :class="config.isRunning ? 'badge-primary' : 'badge-error'">
              <Icon icon="brook-Right-1" style="font-size: 12px;" v-if="config.isRunning"/>
              {{ config.isRunning ? t('server.start') : t('server.stop') }}
            </div>
            <p class="list-col-wrap text-xs">
              {{ t('server.fields.clients') }}:{{ config.clients }}
            </p>
          </td>
          <td>
            <div class="flex items-center gap-1">

              <button class="btn btn-ghost btn-sm btn-square" @click="openWebConfig(config)"
                      :title="t('configuration.delete')"
                      v-if="config.protocol === `HTTP` || config.protocol === `HTTPS`">
                <div class="tooltip tooltip-open tooltip-warning animate-bounce tooltip-left"
                     :data-tip="t('configuration.webNotSet')" v-if="!config.isExistWeb">
                  <Icon icon="brook-web"/>
                </div>
                <div v-else>
                  <Icon icon="brook-web"/>
                </div>
              </button>
              <div v-else>
                <button class="btn btn-ghost btn-sm btn-square btn-disabled"></button>
              </div>
              <button class="btn btn-ghost btn-sm btn-square" @click="handleUpdate(config)">
                <Icon icon="brook-edit"/>
              </button>
              <button class="btn btn-ghost btn-sm btn-square" @click="handleDelete(config.id)"
                      :title="t('configuration.delete')">
                <Icon icon="brook-delete"/>
              </button>
            </div>
          </td>
        </tr>

        <!-- 空状态提示 -->
        <tr v-if="configs.length === 0">
          <td colspan="10" class="text-center py-12">
            <div
                class="w-18 h-18 bg-base-200 rounded-full flex items-center justify-center mx-auto mb-4">
              <Icon icon="brook-technology_usb-cable" class="text-base-content/40"
                    style="font-size: 48px;"/>
            </div>
            <h3 class="text-lg font-medium text-base-content/60 mb-2">{{
                t('configuration.noConfigurations')
              }}</h3>
            <p class="text-sm text-base-content/40 mb-4">{{ t('configuration.noConfigurationsDesc') }}
            </p>
            <button class="btn btn-primary gap-2" @click="handleAdd">
              <Icon icon="brook-add" style="font-size: 16px;"/>
              {{ t('configuration.addConfiguration') }}
            </button>
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