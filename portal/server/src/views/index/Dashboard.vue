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
import baseInfo from '@/service/baseInfo';
import { Line } from 'vue-chartjs'
import {
    ArcElement,
    BarElement,
    CategoryScale,
    Chart,
    Filler,
    Legend,
    LinearScale,
    LineElement,
    PointElement,
    Title,
    Tooltip
} from 'chart.js'
import { computed, onMounted, onUnmounted, ref } from 'vue';
import Icon from '@/components/icon/Index.vue'
// 注册所需的 Chart.js 元件
Chart.register(LineElement, PointElement, CategoryScale, LinearScale, Title, Tooltip, Legend, Filler, ArcElement, BarElement)

const list = ref<any[]>([])
const realTimeData = ref<any[]>([])
let dataUpdateInterval: any = null

// 计算总连接数
const totalConnections = computed(() => {
    return list.value.reduce((total, server) => total + (server.connections || 0), 0)
})

// 为每个服务器生成独立的图表数据
const getServerChartData = (server: any, index: number) => {
    const now = new Date()
    const labels = []
    const connectionsData = []

    // 生成最近15分钟的数据点
    for (let i = 14; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 60000)
        labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }))

        // 基于服务器索引和类型生成不同的数据模式
        const serverSeed = index * 17 // 为每个服务器创建不同的种子
        const timeOffset = i * 0.2

        // 连接数数据 (基于服务器类型和时间)
        let baseConnections = server.connections || (20 + index * 15)
        if (server.tunnelType === 'HTTPS') baseConnections += 30
        else if (server.tunnelType === 'HTTP') baseConnections += 15

        const connectionVariation = Math.sin(timeOffset + serverSeed) * 10
        const randomNoise = (Math.random() - 0.5) * 8
        connectionsData.push(Math.max(0, Math.floor(baseConnections + connectionVariation + randomNoise)))
    }

    return {
        labels,
        datasets: [
            {
                label: '连接数',
                data: connectionsData,
                borderColor: server.tunnelType === 'HTTPS' ? 'rgb(16, 185, 129)' :
                    server.tunnelType === 'HTTP' ? 'rgb(59, 130, 246)' : 'rgb(245, 158, 11)',
                backgroundColor: server.tunnelType === 'HTTPS' ? 'rgba(16, 185, 129, 0.1)' :
                    server.tunnelType === 'HTTP' ? 'rgba(59, 130, 246, 0.1)' : 'rgba(245, 158, 11, 0.1)',
                tension: 0.4,
                fill: true,
                pointRadius: 2,
                pointHoverRadius: 4,
                borderWidth: 2,
            }
        ]
    }
}

// 服务器图表配置
const getServerChartOptions = () => {
    return {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
            legend: {
                display: false, // 隐藏图例以节省空间
            },
            tooltip: {
                mode: 'index' as const,
                intersect: false,
                backgroundColor: 'rgba(0, 0, 0, 0.8)',
                titleColor: 'white',
                bodyColor: 'white',
                borderColor: 'rgba(255, 255, 255, 0.2)',
                borderWidth: 1,
            }
        },
        scales: {
            x: {
                display: true,
                grid: {
                    display: false,
                },
                ticks: {
                    maxTicksLimit: 5,
                    font: {
                        size: 10,
                    }
                }
            },
            y: {
                display: true,
                beginAtZero: true,
                grid: {
                    color: 'rgba(0, 0, 0, 0.1)',
                },
                ticks: {
                    font: {
                        size: 10,
                    }
                }
            }
        },
        interaction: {
            mode: 'nearest' as const,
            axis: 'x' as const,
            intersect: false
        },
        animation: {
            duration: 750,
            easing: 'easeInOutQuart' as const
        }
    }
}

// 启动实时数据更新
const startRealTimeUpdate = () => {
    dataUpdateInterval = setInterval(() => {
        // 触发图表重新渲染
        realTimeData.value = [...realTimeData.value, Date.now()]
    }, 30000) // 每30秒更新一次
}

const initData = async () => {
    const response = await baseInfo.getServerInfo({})
    list.value = response.data || []
}

const avatarCss =(tunnelType) => {
  if (tunnelType === 'https' || tunnelType==="http") {
    return 'bg-gradient-to-br from-primary to-primary/80'
  }else {
    return 'bg-gradient-to-br from-error to-error/80'
  }
}

onMounted(() => {
    startRealTimeUpdate()
    initData()
})

onUnmounted(() => {
    if (dataUpdateInterval) {
        clearInterval(dataUpdateInterval)
    }
})
</script>
<template>
    <div class="space-y-8 p-6">
        <div v-if="list.length === 0" class="justify-center flex flex-col items-center">
            <div class="w-18 h-18 bg-base-300/59 rounded-full flex items-center justify-center mx-auto mb-4">
                <Icon icon="brook-Diagram-" class="text-base-content/30" style="font-size: 48px;" />
            </div>
            <h3 class="text-lg font-medium text-base-content/30 mb-2">暂无服务器通道，当您创建服务器通道后，即可查看</h3>
        </div>
        <!-- 服务器通道卡片网格 -->
        <div v-if="list.length > 0" class="grid grid-cols-1 lg:grid-cols-3 gap-8">
            <div v-for="(server, index) in list" :key="index"
                class="card bg-base-100 shadow-xl border border-base-200/50 hover:shadow-2xl hover:scale-105 transition-all duration-300 group cursor-pointer">
                <div class="card-body p-6">
                    <div class="flex items-center mb-6">
                            <div class="avatar avatar-placeholder">
                                <div
                                    class="text-primary-content rounded-full w-14 h-14" :class="avatarCss(server.tunnelType)">
                                    <span class="text-xl font-bold">{{ server.name?.charAt(0)?.toUpperCase() || 'S'
                                    }}</span>
                                </div>
                            </div>
                            <div class="ml-3 w-full">
                              <div class="flex items-center justify-between">
                                <h3 class="text-xl font-bold text-base-content">{{ server.name }}
                                    <div class="badge badge-sm badge-secondary">{{ server.tag }}</div>
                                </h3>
                              <div class="text-xl font-extralight">{{ server.port || 'N/A' }}
                              </div>
                              </div>
                              <div class="flex items-center space-x-3">
                                <div :class="[
                                        'badge font-medium',
                                        server.tunnelType === 'HTTPS' ? 'badge-success' :
                                            server.tunnelType === 'HTTP' ? 'badge-info' :
                                                'badge-warning'
                                    ]">
                                  {{ server.tunnelType || 'Unknown' }}
                                </div>
                              </div>

                            </div>
                    </div>

                    <!-- 实时监控图表 -->
                    <div class="mb-6">
                        <div class="flex items-center justify-between mb-3">
                            <div class="flex items-center space-x-1">
                                <div class="w-2 h-2 bg-primary rounded-full animate-pulse"></div>
                                <span class="text-xs text-base-content/60">实时更新</span>
                            </div>
                        </div>
                        <div class="h-32 bg-base-200/30 rounded-lg p-3">
                            <Line :data="getServerChartData(server, index)" :options="getServerChartOptions()" />
                        </div>
                    </div>
                    <!-- 关键指标 -->
                    <div class="stats shadow">
                        <div class="stat">
                            <div class="stat-title">连接数</div>
                            <div class="stat-value">{{ server.connections || 0 }}</div>
                        </div>
                        <div class="stat">
                            <div class="stat-title">响应时间</div>
                            <div class="stat-value">{{ server.responseTime || 0 }}</div>
                        </div>
                        <div class="stat">
                            <div class="stat-title">带宽(kb)</div>
                            <div class="stat-value">{{ server.bandwidth || '0' }}</div>
                        </div>
                        <div class="stat">
                            <div class="stat-title">客户端数</div>
                            <div class="stat-value">{{ server.users || 0 }}</div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>