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
import {computed, onMounted, ref} from "vue";
import {useI18n} from '@/components/lang/useI18n';
import Icon from '@/components/icon/Index.vue';
import config from '@/service/config';
import message from "@/components/message";
import fun from "@/service/mysetting";
import {inflate} from "node:zlib";
import Modal from "@/components/modal";

const {t} = useI18n();

const props = defineProps({
  refProxyId: {
    type: Number,
    default: 0
  },
  protocol: {
    type: String,
    default: ""
  },
  inline: {
    type: Boolean,
    default: false
  },
  managed: {
    type: Boolean,
    default: false
  },
})

const isHttps = computed(() => props.protocol === 'HTTPS');

const properties = ref<{
  keyFile: string,
  certFile: string
  certId : number
}>({
  keyFile: "",
  certFile: "",
  certId:0
})

const itmes = ref([])

// 代理列表
interface ProxyItem {
  id: string;
  domain: string;
  paths: string[];
  isEditing: Boolean;
  isNew: Boolean;
}

const proxyList = ref<ProxyItem[]>([]);

const createDefaultProxy = (): ProxyItem => ({
  id: "default",
  domain: "*",
  paths: ["/*"],
  isEditing: false,
  isNew: false
});

// 添加新行
const addNewRow = () => {
  proxyList.value.push({
    id: "",
    domain: "",
    paths: ["/*"],
    isEditing: true,
    isNew: true
  });
};

// 添加新路径
const addPath = (proxy) => {
  proxy.paths.push('');
};

// 删除路径
const removePath = (proxy, index) => {
  proxy.paths.splice(index, 1);
  if (proxy.paths?.length === 0) {
    proxy.paths.push('/*');
  }
};

// 删除代理
const deleteProxy = (index) => {

  Modal.confirm({
    onConfirm: async () => {
      proxyList.value.splice(index, 1);
      return true
    }
  })
};

// 编辑代理
const editProxy = (proxy) => {
  proxy.isEditing = true;
  // 创建备份以便取消时恢复
  proxy._backup = {
    id: proxy.id,
    domain: proxy.domain,
    paths: [...proxy.paths]
  };
};

// 取消编辑
const cancelEdit = (proxy, index) => {
  if (proxy.isNew) {
    // 如果是新行，直接删除
    proxyList.value.splice(index, 1);
  } else if (proxy._backup) {
    // 恢复备份数据
    proxy.id = proxy._backup.id;
    proxy.domain = proxy._backup.domain;
    proxy.paths = [...proxy._backup.paths];
    proxy.isEditing = false;
    delete proxy._backup;
  } else {
    proxy.isEditing = false;
  }
};

// 保存代理
const saveProxy = (proxy) => {
  // 验证表单
  if (!isProxyComplete(proxy)) {
    message.warning(t('configuration.proxyFormIncomplete'));
    return;
  }

  proxy.isEditing = false;
  proxy.isNew = false;

  if (proxy._backup) {
    delete proxy._backup;
  }
};

const isProxyComplete = (proxy) => {
  return !!proxy.id && !!proxy.domain && proxy.paths?.length > 0 && proxy.paths.every(path => !!path);
};

const validateProxyList = () => {
  if (proxyList.value?.length === 0) {
    message.warning(t('configuration.atLeastOneProxy'));
    return false;
  }
  const invalid = proxyList.value.some(proxy => !isProxyComplete(proxy));
  if (invalid) {
    message.warning(t('configuration.proxyFormIncomplete'));
    return false;
  }
  return true;
};

const saveProxyToRemote = async (refProxyId = props.refProxyId, options: { silent?: boolean } = {}) => {
  if (!validateProxyList()) {
    return false;
  }
  const res = await config.addWebConfigs({
    refProxyId,
    certFile: properties.value.certFile,
    keyFile: properties.value.keyFile,
    certId: properties.value.certId,
    proxy: proxyList.value
  });
  if (res.success()) {
    if (!options.silent) {
      message.success(t('success.dataSaved'));
    }
    return true;
  }
  if (!options.silent) {
      message.error(t('errors.operationFailed'));
  }
  return false;
};

const getWebConfigs = () => {
  if (!props.refProxyId) {
    proxyList.value = [createDefaultProxy()];
    return;
  }
  config.getWebConfigs({
    refProxyId: props.refProxyId,
  }).then((res) => {
    var rt = res.data;
    proxyList.value = rt.proxy?.length ? rt.proxy : [createDefaultProxy()];
    properties.value.certFile = rt.certFile || "";
    properties.value.keyFile = rt.keyFile || "";
    properties.value.certId = rt.certId || 0;
  })
}

const getCertificates = () => {
  try {
    fun.getCertificates({}).then(e => {
      itmes.value = e.data;
    });
  } catch (error) {
    console.error('加载证书信息失败:', error)
  }
};

onMounted(() => {
  getWebConfigs()
  getCertificates()
})

defineExpose({
  saveProxyToRemote
});
</script>

<template>
  <div :class="props.inline ? 'w-full space-y-4' : 'container mx-auto p-4'">
    <!-- HTTPS配置区域 -->
    <div v-if="isHttps"
         :class="props.inline ? 'rounded-2xl border border-base-content/10 bg-base-100 p-4' : 'card bg-base-100 shadow-sm hover:shadow-md transition-shadow duration-300 mb-6'"
    >
      <div :class="props.inline ? 'p-0' : 'card-body p-4'">
        <h2 class="card-title text-sm mb-2">{{ t('configuration.httpsCertConfig') }}</h2>
        <div class="form-control">
          <label class="label">
            <span class="label-text">Select Certificate </span>
          </label>
          <div class="control grid items-center gap-2 grid-cols-2 ">
            <select
                class="select select-bordered w-full"
                v-model="properties.certId">
              <option v-for="item in itmes" :class="properties.certId === item.id ? 'selected' : ''" :value="item.id" :key="item.id">
               {{ item.name }} - {{ item.expireTime }}
              </option>
            </select>
            <div class="flex-1">
              <span v-if="itmes?.length === 0" class="text-sm text-red-500">
            {{ t('configuration.certInfoEmpty') }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 代理配置列表 -->
    <div :class="props.inline ? 'rounded-2xl border border-base-content/10 bg-base-100 p-4' : 'card w-full bg-base-100 shadow-sm hover:shadow-md transition-shadow duration-300'">
      <div :class="props.inline ? 'p-0' : 'card-body p-4'">
        <div class="flex justify-between gap-2 items-center mb-1">
          <h2 class="card-title text-sm">{{ t('configuration.proxyList') }}</h2>
          <div v-if="!props.managed" class="flex gap-2">
            <button type="button" @click="saveProxyToRemote()"
                    class="btn btn-primary btn-sm gap-1 transition-transform duration-200 hover:scale-105">
              <Icon icon="brook-add" style="font-size: 14px;"/>
              {{ t('configuration.saveConfig') }}
            </button>
          </div>
        </div>

        <div class="overflow-x-auto rounded-lg">
          <table class="table table-zebra w-full">
            <thead>
            <tr class="bg-base-200/70">
              <th class="w-1/6 rounded-tl-lg">ID</th>
              <th class="w-1/4">{{ t('configuration.domain') }}</th>
              <th>{{ t('configuration.paths') }}</th>
              <th class="w-48 rounded-tr-lg">
                {{ t('configuration.actions') }}
                <button type="button" @click="addNewRow" class="btn btn-outline btn-xs">
                  <Icon icon="brook-add" style="font-size: 12px;"/>
                  {{ t('configuration.addRow') }}
                </button>
              </th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="(proxy, index) in proxyList" :key="index"
                class="hover:bg-base-200/50 transition-colors duration-200">
              <!-- 编辑模式 -->
              <template v-if="proxy.isEditing">
                <td>
                  <input type="text" v-model="proxy.id" :placeholder="t('configuration.egProxyId')"
                         class="input input-bordered input-sm w-full focus:ring focus:ring-primary/20 transition-all duration-200"/>
                </td>
                <td>
                  <input type="text" v-model="proxy.domain" :placeholder="t('configuration.egDomain')"
                         class="input input-bordered input-sm w-full focus:ring focus:ring-primary/20 transition-all duration-200"/>
                </td>
                <td>
                  <div class="flex flex-col gap-2">
                    <div v-for="(path, pathIndex) in proxy.paths" :key="pathIndex" class="flex gap-2 items-center">
                      <input type="text" v-model="proxy.paths[pathIndex]" :placeholder="t('configuration.egPath')"
                             class="input input-bordered input-sm flex-1 focus:ring focus:ring-primary/20 transition-all duration-200"/>
                      <button type="button" @click="removePath(proxy, pathIndex)" class="btn btn-ghost btn-circle btn-sm">
                        <Icon icon="brook-delete" style="font-size: 12px;"/>
                      </button>
                    </div>
                    <button type="button" @click="addPath(proxy)"
                            class="btn btn-outline btn-sm gap-1 self-start mt-1 hover:bg-primary/10 transition-colors duration-200">
                      <Icon icon="brook-add" style="font-size: 12px;"/>
                      {{ t('configuration.addPath') }}
                    </button>
                  </div>
                </td>
                <td>
                  <div class="flex gap-2">
                    <button type="button" @click="saveProxy(proxy)" class="btn btn-sm btn-soft">
                      {{ t('common.confirm') }}
                    </button>
                    <button type="button" @click="cancelEdit(proxy, index)" class="btn btn-sm btn-soft">
                      {{ t('common.cancel') }}
                    </button>
                  </div>
                </td>
              </template>

              <!-- 查看模式 -->
              <template v-else>
                <td class="font-medium">{{ proxy.id }}</td>
                <td>{{ proxy.domain }}</td>
                <td>
                  <div class="flex flex-wrap gap-2">
                      <span v-for="(path, pathIndex) in proxy.paths" :key="pathIndex"
                            class="badge badge-outline hover:badge-primary transition-colors duration-200">
                        {{ path }}
                      </span>
                  </div>
                </td>
                <td>
                  <div class="flex flex-row">
                    <button type="button" @click="editProxy(proxy)" class="btn  btn-sm btn-ghost">
                      <Icon icon="brook-setting" style="font-size: 12px;"/>
                      {{ t('common.edit') }}
                    </button>
                    <button type="button" @click="deleteProxy(index)" class="btn  btn-sm btn-ghost">
                      <Icon icon="brook-delete" style="font-size: 12px;"/>
                      {{ t('common.delete') }}
                    </button>
                  </div>
                </td>
              </template>
            </tr>
            <tr v-if="proxyList?.length === 0">
              <td colspan="4" class="text-center py-4 text-base-content/60">
                {{ t('configuration.noProxyTip') }}
              </td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>
