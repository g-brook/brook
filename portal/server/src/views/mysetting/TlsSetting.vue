<script setup lang="ts">
import {useI18n} from '@/components/lang/useI18n'
import Drawer from "@/components/drawer/Index.vue";
import CertInfo from "@/views/mysetting/CertInfo.vue";
import {onMounted, ref} from "vue";
import fun from "@/service/mysetting";
import Modal from "@/components/modal";
import Icon from "@/components/icon/Index.vue";

const {t, locale} = useI18n()

const drawerRef = ref<{ open: () => void } | null>(null);

const itmes = ref([])

const selectCertItem = ref<any>(null);

const openAddCert = () => {
  selectCertItem.value = null;
  drawerRef.value?.open();
}

const openDetailCert = (item) => {
  selectCertItem.value = item
  drawerRef.value?.open();
}


const deleteCert = (item) => {
  Modal.confirm({
    onConfirm: async () => {
      try {
        await fun.deleteCertificate({id: item.id});
        getAll();
      } catch (error) {
        console.error('删除证书失败:', error)
      }
    }
  })
};

const getAll = () => {
  try {
    fun.getCertificates({}).then(e => {
      itmes.value = e.data;
    });
  } catch (error) {
    console.error('加载证书信息失败:', error)
  }
};

onMounted(() => {
  getAll()
})


</script>

<template>
  <Drawer :key="`tls-drawer-${locale}`" ref="drawerRef" :title="t('mysetting.tls.title')" icon="brook-Certificate-1" width="50%">
    <CertInfo :cert-id="selectCertItem?.id" />
  </Drawer>

  <!-- TLS 管理 - 参考 Configuration.vue 风格 -->
  <div :key="`tls-panel-${locale}`" class="bg-base-200/40 rounded-3xl p-6 border border-base-content/5 space-y-6 shadow-sm">
    <div class="flex items-center justify-between">
      <div class="flex items-center gap-3">
        <div class="w-10 h-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary">
          <Icon icon="brook-Certificate-1" style="font-size: 20px" />
        </div>
        <div>
          <h3 class="text-sm font-black uppercase tracking-widest">{{ t('mysetting.tls.title') }}</h3>
          <p class="text-[10px] font-black opacity-30 uppercase tracking-tighter">{{ t('mysetting.tls.subtitle') }}</p>
        </div>
      </div>
      <div class="flex items-center gap-2">
        <button class="btn btn-primary btn-xs h-8 gap-1.5 font-bold px-3 shadow-md shadow-primary/20 text-xs uppercase tracking-widest" @click="openAddCert">
          <Icon icon="brook-add" style="font-size: 12px;"/>
          {{ t('mysetting.tls.actions.create') }}
        </button>
        <div class="divider divider-horizontal mx-0.5 w-px h-4 self-center opacity-10"></div>
        <button class="btn btn-circle btn-xs h-8 w-8 btn-ghost hover:rotate-180 transition-transform duration-500" @click="getAll">
          <Icon icon="brook-refresh" style="font-size: 14px;"/>
        </button>
      </div>
    </div>

    <!-- 证书表格 -->
    <div class="overflow-x-auto rounded-3xl border border-base-content/5 bg-base-100 shadow-sm">
      <table class="table table-md">
        <thead class="bg-base-200/50">
          <tr>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('mysetting.tls.table.certName') }}</th>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('mysetting.tls.table.desc') }}</th>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em]">{{ t('mysetting.tls.table.certExpireTime') }}</th>
            <th class="font-black text-[13px] uppercase opacity-60 tracking-[0.1em] text-center">{{ t('mysetting.tls.table.certActions') }}</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in itmes" :key="item.id" class="hover:bg-base-200/40 transition-colors group">
            <td class="font-black text-sm tracking-tight">{{ item.name }}</td>
            <td class="text-xs font-black opacity-40 tracking-tight">{{ item.desc }}</td>
            <td class="text-xs font-black opacity-40 font-mono tracking-tighter">{{ item.expireTime }}</td>
            <td>
              <div class="flex items-center justify-center gap-1">
                <button class="btn btn-ghost btn-sm btn-square hover:bg-primary hover:text-primary-content transition-all" @click="openDetailCert(item)">
                  <Icon icon="brook-web" style="font-size: 18px;" />
                </button>
                <button class="btn btn-ghost btn-sm btn-square hover:bg-error hover:text-error-content transition-all" @click="deleteCert(item)">
                  <Icon icon="brook-delete" style="font-size: 18px;" />
                </button>
              </div>
            </td>
          </tr>
          <!-- 空状态 -->
          <tr v-if="itmes.length === 0">
            <td colspan="4" class="text-center py-12 opacity-30">
              <div class="flex flex-col items-center gap-2">
                <Icon icon="brook-empty" style="font-size: 40px;" />
                <span class="text-xs font-black uppercase tracking-widest">{{ t('pagination.noData') }}</span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<style scoped>

</style>
