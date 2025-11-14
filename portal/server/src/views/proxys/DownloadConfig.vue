<script setup lang="ts">
import config from '@/service/config';
import { onMounted, ref } from 'vue';
import JsonEditorVue from 'json-editor-vue'
import Icon from "@/components/icon/Index.vue";
import Message from "@/components/message";
import useI18n from '@/components/lang/useI18n'
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
      .then(() => Message.success(t('success.copied')))
      .catch(() => Message.error(t('errors.copyFailed')))
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

const { t } = useI18n()

</script>

<template>
    <div class="px-2 pt-1">
      <div class="flex gap-2 mb-1">
      <button class="btn btn-primary btn-xs btn-soft" @click="downFileHandler">
        <Icon icon="brook-download" style="font-size: 12px;"/>
        {{ t('common.download') }}
      </button>
      <button class="btn btn-primary btn-xs btn-soft" @click="copyContent">
        {{ t('common.copy') }}
      </button>
      </div>
            <JsonEditorVue v-model="code" mode="text"
            :mainMenuBar="false"
            :statusBar="true"
            class="w-full rounded-lg overflow-hidden h-full json-editor-minimal"
            />
    </div>

</template>

<style>
/* ===== 🎨 Minimal Designer Theme for json-editor-vue ===== */

.json-editor-minimal {
  --jse-font-family: "JetBrains Mono", "Fira Code", monospace;
  --jse-font-size: 14px;
  --jse-line-height: 22px;
  --jse-background-color: var(--color-base-200);
  --jse-panel-background: var(--color-base-300);
  --jse-key-color: var(--color-base-content);
  --jse-value-color-string: var(--color-info);
  --jse-value-color-number: var(--color-error);
  --jse-value-color-boolean: var(--color-info);
  --jse-value-color-null: var(--color-info);
  --jse-value-color-url: var(--color-base-200);
  --jse-delimiter-color: var(--color-base-content);
  transition: all 0.2s ease-in-out;
  padding: 8px; /* 可选，避免紧贴边缘 */
}

　
.json-editor-minimal:hover {
  box-shadow: 0 3px 12px rgba(0,0,0,0.08);
}


</style>