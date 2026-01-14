<script setup lang="ts">
import {ref, reactive, onMounted} from 'vue'
import fun from '@/service/mysetting'
import message from "@/components/message";

// 定义证书表单接口
interface CertForm {
  name: string
  certType: string
  content: string,
  privateKey: string,
  desc: string
  isAdd: boolean
}

// 定义组件的props和emit
interface Props {
  certId?: string
}

const props = withDefaults(defineProps<Props>(), {
  certId: '',
})

// 表单数据
const certForm = reactive<CertForm>({
  name: '',
  certType: '',
  content: '',
  privateKey: '',
  desc: '',
  isAdd: true
})

const initValues = () => {
  certForm.name = ''
  certForm.certType = ''
  certForm.content = ''
  certForm.privateKey = ''
  certForm.desc = ''
}

// 提交加载状态
const submitLoading = ref(false)

// 提交表单
const handleSubmit = async () => {
  if (!validateForm()) {
    return
  }
  submitLoading.value = true
  try {
    // 这里应该调用API保存证书信息
    fun.addCertificate(certForm).then(e => {
      if (e.success()) {
        message.success('Certificate added successfully.')
      }
    })
    // 提交成功后的处理
  } finally {
    submitLoading.value = false
  }
}

// 验证表单
const validateForm = (): boolean => {
  return true
}

onMounted(() => {
  if (props.certId) {
    certForm.isAdd = false
    loadCertInfo(props.certId)
  }else {
    initValues()
    certForm.isAdd = true
  }
})

// 加载证书信息的方法
const loadCertInfo = async (id: string) => {
  try {
    fun.getCertificateById({id: id}).then(e => {
      if (e.success()) {
        certForm.name = e.data.name
        certForm.privateKey = e.data.privateKey
        certForm.desc = e.data.desc
        certForm.content = e.data.content
      }
    });
  } catch (error) {
    console.error('加载证书信息失败:', error)
  }
}
</script>

<template>
  <div>
    <div class="card bg-base-100 shadow-sm">
      <div class="card-body">
        <form class="space-y-4">
          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">证书名称</span>
              <div class="label">
                <span class="label-text-alt">{{ certForm.name.length }}/20</span>
              </div>
            </label>
            <input
                v-model="certForm.name"
                type="text"
                placeholder="请输入证书名称"
                :disabled="!certForm.isAdd"
                maxlength="20"
                class="input input-bordered w-full"
            />
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">证书内容</span>
            </label>
            <textarea
                v-model="certForm.content"
                :readonly="!certForm.isAdd"
                placeholder="-----BEGIN CERTIFICATE-----               -----END CERTIFICATE-----"
                rows="6"
                class="textarea textarea-bordered w-full font-mono text-sm resize-none"
            ></textarea>
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">私钥内容</span>
            </label>
            <textarea
                :readonly="!certForm.isAdd"
                v-model="certForm.privateKey"
                placeholder="-----BEGIN PRIVATE KEY-----              -----END PRIVATE KEY-----"
                rows="6"
                class="textarea textarea-bordered w-full font-mono text-sm resize-none"
            ></textarea>
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">备注信息   <div class="label">
              <span class="label-text-alt">{{ certForm.desc.length }}/50</span>
            </div></span>
            </label>
            <input
                v-model="certForm.desc"
                :readonly="!certForm.isAdd"
                placeholder="请输入备注信息"
                maxlength="50"
                class=" w-full input"
            />
          </div>

          <div class="flex gap-2" v-if="certForm.isAdd">
            <button
                type="button"
                @click="handleSubmit"
                :disabled="submitLoading"
                class="btn btn-primary"
            >
              <span v-if="submitLoading" class="loading loading-spinner"></span>
              保存证书信息
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>