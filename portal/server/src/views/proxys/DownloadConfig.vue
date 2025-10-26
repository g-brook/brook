<script setup lang="ts">
import config from '@/service/config';
import { onMounted, ref } from 'vue';
import JsonEditorVue from 'json-editor-vue'
import Icon from "@/components/icon/Index.vue";
import Message from "@/components/message";
const code = ref('')

const getConfigs = async () => {
    try {
        const res = await config.genClientConfig();
        if (res.success()) {
            code.value = res.data;
        }
    } catch (error) {
    }
};
function downloadFile(filename, content) {
  const blob = new Blob([content], { type: 'application/json' });
  const link = document.createElement('a');
  link.href = URL.createObjectURL(blob);
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(link.href);
}

const copyContent = ()=> {
  navigator.clipboard.writeText(JSON.stringify(code.value))
      .then(() => Message.success('复制成功'))
      .catch(() => Message.error("复制失败"))
}

const downFileHandler = async () => {
  if (!code.value) {
    return;
  }
  downloadFile("client.json",JSON.stringify(code.value, null, 2))
}

onMounted(() => {
    getConfigs();
})

</script>

<template>
    <div class="px-2 pt-1">
      <div class="flex gap-2 mb-1">
      <button class="btn btn-primary btn-xs btn-soft" @click="downFileHandler">
        <Icon icon="brook-download" style="font-size: 12px;"/>
        下载
      </button>
      <button class="btn btn-primary btn-xs btn-soft" @click="copyContent">
        复制
      </button>
      </div>
            <JsonEditorVue v-model="code" mode="text"
            :mainMenuBar="false"
            :statusBar="true"
            class="w-full rounded-lg overflow-hidden h-full"
            />
    </div>

</template>