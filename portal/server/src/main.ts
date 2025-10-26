/*
 * Copyright ©  sixh sixh@apache.org
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

import {createApp} from 'vue'
import App from './App.vue'
import './style.css'
import '@/assets/css/iconfont.css'
import router from './components/routes'
import ModalContainer from '@/components/modal/ModalContainer.vue'
// 初始化国际化系统
import '@/i18n'

// 创建 Vue 应用实例
const app = createApp(App)

// 配置路由并挂载应用
app.use(router).mount('#app')

// 创建独立的模态框容器并挂载到 body
const modalContainer = document.createElement('div')
modalContainer.id = 'modal-container'
document.body.appendChild(modalContainer)

const modalApp = createApp(ModalContainer)
modalApp.mount('#modal-container')

