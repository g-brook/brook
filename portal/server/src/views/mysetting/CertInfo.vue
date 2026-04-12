<script setup lang="ts">
import {ref, reactive, onMounted} from 'vue'
import fun from '@/service/mysetting'
import message from "@/components/message";
import {useI18n} from '@/components/lang/useI18n';

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
const {t} = useI18n();

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
        message.success(t('mysetting.tls.form.messages.addSuccess'))
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
              <span class="label-text">{{ t('mysetting.tls.form.nameLabel') }}</span>
              <div class="label">
                <span class="label-text-alt">{{ certForm.name.length }}/20</span>
              </div>
            </label>
            <input
                v-model="certForm.name"
                type="text"
                :placeholder="t('mysetting.tls.form.namePlaceholder')"
                :disabled="!certForm.isAdd"
                maxlength="20"
                class="input input-bordered w-full"
            />
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">{{ t('mysetting.tls.form.contentLabel') }}</span>
            </label>
            <textarea
                v-model="certForm.content"
                :readonly="!certForm.isAdd"
                :placeholder="t('mysetting.tls.form.contentPlaceholder')"
                rows="6"
                class="textarea textarea-bordered w-full font-mono text-sm resize-none"
            ></textarea>
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">{{ t('mysetting.tls.form.privateKeyLabel') }}</span>
            </label>
            <textarea
                :readonly="!certForm.isAdd"
                v-model="certForm.privateKey"
                :placeholder="t('mysetting.tls.form.privateKeyPlaceholder')"
                rows="6"
                class="textarea textarea-bordered w-full font-mono text-sm resize-none"
            ></textarea>
          </div>

          <div class="form-control w-full">
            <label class="label">
              <span class="label-text">{{ t('mysetting.tls.form.descLabel') }}</span>
              <div class="label">
                <span class="label-text-alt">{{ certForm.desc.length }}/50</span>
              </div>
            </label>
            <input
                v-model="certForm.desc"
                :readonly="!certForm.isAdd"
                :placeholder="t('mysetting.tls.form.descPlaceholder')"
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
              {{ t('mysetting.tls.form.save') }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<style scoped>

</style>
