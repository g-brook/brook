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
import {computed} from 'vue'

interface Props {
  icon: string
}

const props = defineProps<Props>()

// 拼接 #icon-xxx
const iconNameFun = computed(() => `${props.icon}`)

</script>

<!-- //00c6ff -->
<template>
  <span class="iconfont" :class="iconNameFun"></span>
</template>