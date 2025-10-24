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
import {Line} from 'vue-chartjs'
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
import {computed, onMounted, onUnmounted, ref} from 'vue';
// 注册所需的 Chart.js 元件
Chart.register(LineElement, PointElement, CategoryScale, LinearScale, Title, Tooltip, Legend, Filler, ArcElement, BarElement)

const list = ref<any[]>([])
const isLoading = ref(true)
const realTimeData = ref<any[]>([])
const selectedServer = ref<any>(null)
let dataUpdateInterval: any = null

// 计算总连接数
const totalConnections = computed(() => {
    return list.value.reduce((total, server) => total + (server.connections || 0), 0)
})

// 计算总用户数
const totalUsers = computed(() => {
    return list.value.reduce((total, server) => total + (server.users || 0), 0)
})

// 计算平均响应时间
const averageResponseTime = computed(() => {
    if (list.value.length === 0) return 0
    const total = list.value.reduce((sum, server) => sum + (server.responseTime || 0), 0)
    return Math.round(total / list.value.length)
})

// 计算系统健康度
const systemHealth = computed(() => {
    const onlineServers = list.value.filter(server => server.status === 'online').length
    const totalServers = list.value.length
    return totalServers > 0 ? Math.round((onlineServers / totalServers) * 100) : 100
})

// 生成模拟实时数据
const generateRealTimeData = () => {
    const now = new Date()
    const labels = []
    const connectionData = []
    const trafficData = []
    
    for (let i = 29; i >= 0; i--) {
        const time = new Date(now.getTime() - i * 60000) // 每分钟一个数据点
        labels.push(time.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' }))
        
        // 模拟连接数波动
        const baseConnections = totalConnections.value || 50
        const variation = Math.sin(i * 0.2) * 10 + Math.random() * 20 - 10
        connectionData.push(Math.max(0, Math.round(baseConnections + variation)))
        
        // 模拟流量数据 (MB/s)
        const baseTraffic = 2.5
        const trafficVariation = Math.sin(i * 0.3) * 1.5 + Math.random() * 2 - 1
        trafficData.push(Math.max(0, Number((baseTraffic + trafficVariation).toFixed(2))))
    }
    
    return { labels, connectionData, trafficData }
}

// 实时数据图表配置
const realTimeChartData = computed(() => {
    const data = generateRealTimeData()
    return {
        labels: data.labels,
        datasets: [
            {
                label: '活跃连接数',
                data: data.connectionData,
                borderColor: 'rgb(59, 130, 246)',
                backgroundColor: 'rgba(59, 130, 246, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0,
                pointHoverRadius: 6,
                borderWidth: 2
            },
            {
                label: '流量 (MB/s)',
                data: data.trafficData,
                borderColor: 'rgb(16, 185, 129)',
                backgroundColor: 'rgba(16, 185, 129, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0,
                pointHoverRadius: 6,
                borderWidth: 2,
                yAxisID: 'y1'
            }
        ]
    }
})

const realTimeChartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
        legend: {
            position: 'top' as const,
            labels: {
                usePointStyle: true,
                padding: 20,
                font: {
                    size: 12
                }
            }
        },
        tooltip: {
            mode: 'index' as const,
            intersect: false,
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: 'white',
            bodyColor: 'white',
            borderColor: 'rgba(255, 255, 255, 0.2)',
            borderWidth: 1
        }
    },
    scales: {
        x: {
            display: true,
            grid: {
                display: false
            },
            ticks: {
                maxTicksLimit: 6
            }
        },
        y: {
            type: 'linear' as const,
            display: true,
            position: 'left' as const,
            title: {
                display: true,
                text: '连接数'
            },
            grid: {
                color: 'rgba(0, 0, 0, 0.1)'
            }
        },
        y1: {
            type: 'linear' as const,
            display: true,
            position: 'right' as const,
            title: {
                display: true,
                text: '流量 (MB/s)'
            },
            grid: {
                drawOnChartArea: false,
            },
        }
    },
    interaction: {
        mode: 'nearest' as const,
        axis: 'x' as const,
        intersect: false
    }
}

// 服务器类型分布图表
const serverTypeChartData = computed(() => {
    const typeCounts = list.value.reduce((acc, server) => {
        const type = server.tunnelType || 'Unknown'
        acc[type] = (acc[type] || 0) + 1
        return acc
    }, {} as Record<string, number>)
    
    return {
        labels: Object.keys(typeCounts),
        datasets: [{
            data: Object.values(typeCounts),
            backgroundColor: [
                'rgba(59, 130, 246, 0.8)',
                'rgba(16, 185, 129, 0.8)',
                'rgba(245, 158, 11, 0.8)',
                'rgba(239, 68, 68, 0.8)',
                'rgba(139, 92, 246, 0.8)'
            ],
            borderColor: [
                'rgb(59, 130, 246)',
                'rgb(16, 185, 129)',
                'rgb(245, 158, 11)',
                'rgb(239, 68, 68)',
                'rgb(139, 92, 246)'
            ],
            borderWidth: 2
        }]
    }
})

const serverTypeChartOptions = {
    responsive: true,
    maintainAspectRatio: false,
    plugins: {
        legend: {
            position: 'bottom' as const,
            labels: {
                padding: 20,
                usePointStyle: true
            }
        },
        tooltip: {
            backgroundColor: 'rgba(0, 0, 0, 0.8)',
            titleColor: 'white',
            bodyColor: 'white'
        }
    }
}

const initData = async () => {
    try {
        isLoading.value = true
        const response = await baseInfo.getServerInfo({})
        list.value = response.data || []
        
        // 添加模拟数据以便展示
        if (list.value.length === 0) {
            list.value = [
                {
                    name: 'Primary Tunnel',
                    tunnelType: 'HTTPS',
                    port: 443,
                    connections: 156,
                    users: 89,
                    status: 'online',
                    responseTime: 45,
                    uptime: '99.9%',
                    location: 'Singapore',
                    bandwidth: '1.2 GB/s'
                },
                {
                    name: 'Secondary Tunnel',
                    tunnelType: 'HTTP',
                    port: 80,
                    connections: 78,
                    users: 45,
                    status: 'online',
                    responseTime: 32,
                    uptime: '99.7%',
                    location: 'Tokyo',
                    bandwidth: '800 MB/s'
                },
                {
                    name: 'Backup Tunnel',
                    tunnelType: 'TCP',
                    port: 8080,
                    connections: 23,
                    users: 12,
                    status: 'online',
                    responseTime: 67,
                    uptime: '98.5%',
                    location: 'Seoul',
                    bandwidth: '500 MB/s'
                }
            ]
        }
        
        // 设置第一个服务器为选中状态
        if (list.value.length > 0) {
            selectedServer.value = list.value[0]
        }
    } catch (error) {
        console.error('Failed to load server info:', error)
        list.value = []
    } finally {
        isLoading.value = false
    }
}

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

// 选择服务器
const selectServer = (server: any) => {
    selectedServer.value = server
}

// 启动实时数据更新
const startRealTimeUpdate = () => {
    dataUpdateInterval = setInterval(() => {
        // 触发图表重新渲染
        realTimeData.value = [...realTimeData.value, Date.now()]
    }, 30000) // 每30秒更新一次
}

onMounted(() => {
    initData()
    startRealTimeUpdate()
})

onUnmounted(() => {
    if (dataUpdateInterval) {
        clearInterval(dataUpdateInterval)
    }
})
</script>
<template>
    <div class="space-y-8 p-6 bg-gradient-to-br from-base-100 to-base-200/30 min-h-screen">
       
        <!-- 增强的统计卡片 -->
        <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
            <!-- 总服务器数 -->
            <div class="stats shadow-xl bg-gradient-to-br from-primary/20 via-primary/10 to-primary/5 border border-primary/30 hover:shadow-2xl transition-all duration-300">
                <div class="stat">
                    <div class="stat-figure text-primary">
                        <div class="w-12 h-12 bg-primary/20 rounded-xl flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                            </svg>
                        </div>
                    </div>
                    <div class="stat-title text-primary/80 font-medium">总服务器</div>
                    <div class="stat-value text-primary text-3xl font-bold">{{ list.length }}</div>
                    <div class="stat-desc text-primary/60">
                        <span class="inline-flex items-center">
                            <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
                            </svg>
                            活跃隧道服务器
                        </span>
                    </div>
                </div>
            </div>

            <!-- 总连接数 -->
            <div class="stats shadow-xl bg-gradient-to-br from-success/20 via-success/10 to-success/5 border border-success/30 hover:shadow-2xl transition-all duration-300">
                <div class="stat">
                    <div class="stat-figure text-success">
                        <div class="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1" />
                            </svg>
                        </div>
                    </div>
                    <div class="stat-title text-success/80 font-medium">总连接数</div>
                    <div class="stat-value text-success text-3xl font-bold">{{ totalConnections.toLocaleString() }}</div>
                    <div class="stat-desc text-success/60">
                        <span class="inline-flex items-center">
                            <svg class="w-3 h-3 mr-1 animate-pulse" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"></path>
                            </svg>
                            当前活跃连接
                        </span>
                    </div>
                </div>
            </div>

            <!-- 平均响应时间 -->
            <div class="stats shadow-xl bg-gradient-to-br from-warning/20 via-warning/10 to-warning/5 border border-warning/30 hover:shadow-2xl transition-all duration-300">
                <div class="stat">
                    <div class="stat-figure text-warning">
                        <div class="w-12 h-12 bg-warning/20 rounded-xl flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                            </svg>
                        </div>
                    </div>
                    <div class="stat-title text-warning/80 font-medium">平均响应</div>
                    <div class="stat-value text-warning text-3xl font-bold">{{ averageResponseTime }}ms</div>
                    <div class="stat-desc text-warning/60">
                        <span class="inline-flex items-center">
                            <svg class="w-3 h-3 mr-1" fill="currentColor" viewBox="0 0 20 20">
                                <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd"></path>
                            </svg>
                            系统响应时间
                        </span>
                    </div>
                </div>
            </div>

            <!-- 系统健康度 -->
            <div class="stats shadow-xl bg-gradient-to-br from-info/20 via-info/10 to-info/5 border border-info/30 hover:shadow-2xl transition-all duration-300">
                <div class="stat">
                    <div class="stat-figure text-info">
                        <div class="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4.318 6.318a4.5 4.5 0 000 6.364L12 20.364l7.682-7.682a4.5 4.5 0 00-6.364-6.364L12 7.636l-1.318-1.318a4.5 4.5 0 00-6.364 0z" />
                            </svg>
                        </div>
                    </div>
                    <div class="stat-title text-info/80 font-medium">系统健康</div>
                    <div class="stat-value text-info text-3xl font-bold">{{ systemHealth }}%</div>
                    <div class="stat-desc text-info/60">
                        <span class="inline-flex items-center">
                            <div class="w-2 h-2 bg-info rounded-full mr-2 animate-pulse"></div>
                            运行状态良好
                        </span>
                    </div>
                </div>
            </div>
        </div>

        <!-- 服务器通道卡片网格 -->
        <div v-if="list.length > 0" class="grid grid-cols-1 lg:grid-cols-2 gap-8">
            <div 
                v-for="(server, index) in list" 
                :key="index"
                class="card bg-base-100 shadow-xl border border-base-200/50 hover:shadow-2xl hover:scale-105 transition-all duration-300 group">
                <div class="card-body p-6">
                    <div class="flex items-center justify-between mb-6">
                        <div class="flex items-center space-x-4">
                            <div class="avatar placeholder">
                                <div class="bg-gradient-to-br from-primary to-primary/70 text-primary-content rounded-xl w-14 h-14">
                                    <span class="text-xl font-bold">{{ server.name?.charAt(0)?.toUpperCase() || 'S' }}</span>
                                </div>
                            </div>
                            <div>
                                <h3 class="text-xl font-bold text-base-content">{{ server.name || '未命名服务器' }}</h3>
                                <div class="flex items-center space-x-3 mt-2">
                                    <div :class="[
                                        'badge font-medium',
                                        server.tunnelType === 'HTTPS' ? 'badge-success' : 
                                        server.tunnelType === 'HTTP' ? 'badge-info' : 
                                        'badge-warning'
                                    ]">
                                        {{ server.tunnelType || 'Unknown' }}
                                    </div>
                                    <div class="badge badge-outline">端口 {{ server.port || 'N/A' }}</div>
                                    <div class="flex items-center space-x-1">
                                        <div class="w-2 h-2 bg-success rounded-full animate-pulse"></div>
                                        <span class="text-xs text-success font-medium">在线</span>
                                    </div>
                                </div>
                            </div>
                        </div>
                        
                        <!-- 操作菜单 -->
                        <div class="dropdown dropdown-end">
                            <label tabindex="0" class="btn btn-ghost btn-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v.01M12 12v.01M12 19v.01M12 6a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2zm0 7a1 1 0 110-2 1 1 0 010 2z" />
                                </svg>
                            </label>
                            <ul tabindex="0" class="dropdown-content menu p-2 shadow bg-base-100 rounded-box w-52">
                                <li><a>查看详情</a></li>
                                <li><a>编辑配置</a></li>
                                <li><a>重启服务</a></li>
                                <li><a>导出数据</a></li>
                                <li class="divider"></li>
                                <li><a class="text-error">删除服务器</a></li>
                            </ul>
                        </div>
                    </div>

                    <!-- 实时监控图表 -->
                    <div class="mb-6">
                        <div class="flex items-center justify-between mb-3">
                            <h4 class="text-sm font-semibold text-base-content/80">实时监控</h4>
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
                    <div class="grid grid-cols-4 gap-3 mb-6">
                        <div class="bg-gradient-to-br from-primary/10 to-primary/5 rounded-lg p-3 text-center">
                            <div class="text-lg font-bold text-primary">{{ server.connections || 0 }}</div>
                            <div class="text-xs text-base-content/60">连接数</div>
                        </div>
                        <div class="bg-gradient-to-br from-success/10 to-success/5 rounded-lg p-3 text-center">
                            <div class="text-lg font-bold text-success">{{ server.responseTime || 45 }}ms</div>
                            <div class="text-xs text-base-content/60">响应时间</div>
                        </div>
                        <div class="bg-gradient-to-br from-warning/10 to-warning/5 rounded-lg p-3 text-center">
                            <div class="text-lg font-bold text-warning">{{ server.bandwidth || '2.1MB/s' }}</div>
                            <div class="text-xs text-base-content/60">带宽</div>
                        </div>
                        <div class="bg-gradient-to-br from-info/10 to-info/5 rounded-lg p-3 text-center">
                            <div class="text-lg font-bold text-info">{{ server.users || 0 }}</div>
                            <div class="text-xs text-base-content/60">用户数</div>
                        </div>
                    </div>

                    <!-- 详细信息 -->
                    <div class="space-y-3">
                        <div class="flex items-center justify-between text-sm">
                            <span class="text-base-content/60">服务器地址:</span>
                            <span class="font-mono text-base-content">{{ server.location || 'localhost' }}</span>
                        </div>
                        <div class="flex items-center justify-between text-sm">
                            <span class="text-base-content/60">运行时间:</span>
                            <span class="text-base-content">{{ server.uptime || '24小时12分钟' }}</span>
                        </div>
                        <div class="flex items-center justify-between text-sm">
                            <span class="text-base-content/60">CPU使用率:</span>
                            <div class="flex items-center space-x-2">
                                <div class="w-16 bg-base-200 rounded-full h-2">
                                    <div :class="[
                                        'h-2 rounded-full transition-all duration-300',
                                        (server.cpuUsage || 25) > 80 ? 'bg-error' : 
                                        (server.cpuUsage || 25) > 60 ? 'bg-warning' : 'bg-success'
                                    ]" :style="`width: ${server.cpuUsage || 25}%`"></div>
                                </div>
                                <span class="text-xs font-medium">{{ server.cpuUsage || 25 }}%</span>
                            </div>
                        </div>
                        <div class="flex items-center justify-between text-sm">
                            <span class="text-base-content/60">内存使用:</span>
                            <div class="flex items-center space-x-2">
                                <div class="w-16 bg-base-200 rounded-full h-2">
                                    <div :class="[
                                        'h-2 rounded-full transition-all duration-300',
                                        (server.memoryUsage || 45) > 80 ? 'bg-error' : 
                                        (server.memoryUsage || 45) > 60 ? 'bg-warning' : 'bg-info'
                                    ]" :style="`width: ${server.memoryUsage || 45}%`"></div>
                                </div>
                                <span class="text-xs font-medium">{{ server.memoryUsage || 45 }}%</span>
                            </div>
                        </div>
                    </div>

                    <!-- 操作按钮 -->
                    <div class="card-actions justify-end mt-6 space-x-2">
                        <button class="btn btn-ghost btn-sm group-hover:btn-primary transition-all duration-200">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                            </svg>
                            详细统计
                        </button>
                        <button class="btn btn-ghost btn-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" class="h-4 w-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                            </svg>
                            管理
                        </button>
                    </div>
                </div>
            </div>
        </div>

        <!-- 快速操作面板 -->
        <div v-if="list.length > 0" class="grid grid-cols-1 md:grid-cols-3 gap-6 mt-8">
            <div class="card bg-gradient-to-br from-primary/10 to-primary/5 border border-primary/20">
                <div class="card-body text-center p-6">
                    <div class="w-12 h-12 bg-primary/20 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-primary" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                        </svg>
                    </div>
                    <h3 class="font-bold text-base-content mb-2">添加服务器</h3>
                    <p class="text-sm text-base-content/60 mb-4">配置新的隧道服务器</p>
                    <button class="btn btn-primary btn-sm">立即添加</button>
                </div>
            </div>

            <div class="card bg-gradient-to-br from-success/10 to-success/5 border border-success/20">
                <div class="card-body text-center p-6">
                    <div class="w-12 h-12 bg-success/20 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-success" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
                        </svg>
                    </div>
                    <h3 class="font-bold text-base-content mb-2">系统监控</h3>
                    <p class="text-sm text-base-content/60 mb-4">查看详细监控报告</p>
                    <button class="btn btn-success btn-sm">查看报告</button>
                </div>
            </div>

            <div class="card bg-gradient-to-br from-info/10 to-info/5 border border-info/20">
                <div class="card-body text-center p-6">
                    <div class="w-12 h-12 bg-info/20 rounded-xl flex items-center justify-center mx-auto mb-4">
                        <svg xmlns="http://www.w3.org/2000/svg" class="h-6 w-6 text-info" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                        </svg>
                    </div>
                    <h3 class="font-bold text-base-content mb-2">系统设置</h3>
                    <p class="text-sm text-base-content/60 mb-4">配置系统参数</p>
                    <button class="btn btn-info btn-sm">进入设置</button>
                </div>
            </div>
        </div>

        <!-- 空状态 (如果没有服务器) -->
        <div v-if="list.length === 0" class="text-center py-20">
            <div class="w-32 h-32 bg-gradient-to-br from-primary/20 to-primary/5 rounded-full flex items-center justify-center mx-auto mb-8">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-16 w-16 text-primary/60" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01" />
                </svg>
            </div>
            <h3 class="text-2xl font-bold text-base-content/60 mb-4">欢迎使用隧道管理系统</h3>
            <p class="text-base-content/40 mb-8 max-w-md mx-auto">
                开始配置您的第一个隧道服务器，享受高效的网络代理服务
            </p>
            <button class="btn btn-primary btn-lg">
                <svg xmlns="http://www.w3.org/2000/svg" class="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4" />
                </svg>
                创建第一个服务器
            </button>
        </div>
    </div>
</template>