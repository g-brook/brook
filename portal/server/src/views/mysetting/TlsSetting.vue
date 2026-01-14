<script setup lang="ts">
import useI18n from '@/components/lang/useI18n'
import Drawer from "@/components/drawer/Index.vue";
import CertInfo from "@/views/mysetting/CertInfo.vue";
import {onMounted, ref} from "vue";
import fun from "@/service/mysetting";

const {t} = useI18n()

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
  try {
    fun.deleteCertificate({id: item.id}).then(() => {
      getAll();
    });

  } catch (error) {
    console.error('删除证书失败:', error)
  }
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
  <Drawer ref="drawerRef" :title="t('mysetting.tls.title')" icon="brook-web" width="50%">
    <CertInfo :cert-id="selectCertItem?.id" />
  </Drawer>
  <div class="space-y-4">
    <div class="border border-base-300 rounded-lg  w-full">
      <div class="card-title">
        <div class="flex w-full items-center p-4">
          <div class="flex-1 flex space-x-2 flex-row items-center ">
            <i class="iconfont brook-token" style="font-size: 24px"></i>
            <h2 class="text-base-content text-xl">{{ t('mysetting.tls.title') }}</h2>
            <p class="text-xs text-base-content/60">{{ t('mysetting.tls.subtitle') }}</p>
          </div>
          <div class=" flex-row items-center">
            <button class="btn btn-primary btn-sm" @click="openAddCert">
              <i class="iconfont brook-plus"></i>
              {{ t('mysetting.tls.actions.create') }}
            </button>
          </div>
        </div>
      </div>
      <div class="bg-base-100 rounded-lg p-4">
        <div class="overflow-x-auto">
          <table class="table table-zebra">
            <thead>
            <tr>
              <th>{{ t('mysetting.tls.table.certName') }}</th>
              <th>{{ t('mysetting.tls.table.desc') }}</th>
              <th>{{ t('mysetting.tls.table.certExpireTime') }}</th>
              <th>{{ t('mysetting.tls.table.certActions') }}</th>
            </tr>
            </thead>
            <tbody>
            <tr v-for="item in itmes">
              <td>{{ item.name }}</td>
              <td>{{ item.desc }}</td>
              <td>{{ item.expireTime }}</td>
              <td class="space-x-2">
                <button class="btn btn-sm btn-ghost" @click="openDetailCert(item)">{{ t('mysetting.tls.actions.view') }}</button>
                <button class="btn btn-sm btn-error btn-outline" @click="deleteCert(item)">
                  {{ t('mysetting.tls.actions.delete') }}
                </button>
              </td>
            </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>