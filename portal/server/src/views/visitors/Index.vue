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
import { computed, ref } from "vue";
import Icon from "@/components/icon/Index.vue";

type VisitorStatus = "online" | "offline" | "error";
interface VisitorItem {
  id: number;
  name: string;
  serverName: string;
  providerName: string;
  protocol: "stcp" | "xtcp";
  bindAddr: string;
  bindPort: number;
  status: VisitorStatus;
  enabled: boolean;
  lastSeen: string;
  p95Latency: number;
  currentConns: number;
  totalIn: string;
  totalOut: string;
}

const loading = ref(false);
const keyword = ref("");
const statusFilter = ref<"all" | VisitorStatus>("all");
const selectedVisitor = ref<VisitorItem | null>(null);

const visitors = ref<VisitorItem[]>([
  { id: 1, name: "ssh-office", serverName: "proxy-ssh-main", providerName: "edge-shanghai-01", protocol: "stcp", bindAddr: "127.0.0.1", bindPort: 6000, status: "online", enabled: true, lastSeen: "2026-05-10 10:23:11", p95Latency: 26, currentConns: 3, totalIn: "1.2 GB", totalOut: "2.1 GB" },
  { id: 2, name: "mysql-ro", serverName: "proxy-mysql-ro", providerName: "edge-beijing-02", protocol: "stcp", bindAddr: "127.0.0.1", bindPort: 6330, status: "offline", enabled: false, lastSeen: "2026-05-09 23:51:02", p95Latency: 0, currentConns: 0, totalIn: "0 B", totalOut: "0 B" },
  { id: 3, name: "rdp-lab", serverName: "proxy-rdp-lab", providerName: "edge-hk-01", protocol: "xtcp", bindAddr: "127.0.0.1", bindPort: 63389, status: "error", enabled: true, lastSeen: "2026-05-10 09:58:17", p95Latency: 108, currentConns: 0, totalIn: "842 MB", totalOut: "1.6 GB" }
]);

const hasProviderBinding = (v: VisitorItem) => !!v.providerName?.trim() && !!v.serverName?.trim();
const providerVisitors = computed(() => visitors.value.filter(hasProviderBinding));
const filtered = computed(() =>
  providerVisitors.value.filter((v) => {
    const hitStatus = statusFilter.value === "all" || v.status === statusFilter.value;
    const key = keyword.value.trim().toLowerCase();
    const hitKey =
      !key ||
      v.name.toLowerCase().includes(key) ||
      v.serverName.toLowerCase().includes(key) ||
      v.providerName.toLowerCase().includes(key);
    return hitStatus && hitKey;
  })
);
const onlineCount = computed(() => filtered.value.filter(v => v.status === "online").length);

const statusText = (status: VisitorStatus) => (status === "online" ? "在线" : status === "error" ? "异常" : "离线");
const statusBadgeClass = (status: VisitorStatus) => {
  if (status === "online") return "badge-success";
  if (status === "error") return "badge-error";
  return "badge-ghost";
};
const bandClass = (status: VisitorStatus) => {
  if (status === "online") return "from-success/20 text-success border-success/30";
  if (status === "error") return "from-error/20 text-error border-error/30";
  return "from-base-300 text-base-content/60 border-base-content/20";
};

const toggleEnabled = (item: VisitorItem) => {
  item.enabled = !item.enabled;
  item.status = item.enabled ? "online" : "offline";
};

const handleSelect = (item: VisitorItem) => {
  selectedVisitor.value = item;
};
</script>

<template>
  <div class="brook-page">
    <div class="brook-toolbar">
      <div class="flex items-center gap-4">
        <div class="brook-stat">
          <div class="brook-stat-dot bg-primary"></div>
          <span class="text-xs font-black uppercase opacity-50 tracking-tighter">Visitors</span>
          <span class="text-sm font-black tracking-tighter">{{ providerVisitors.length }}</span>
        </div>
        <div class="brook-stat">
          <div class="brook-stat-dot bg-success"></div>
          <span class="text-xs font-black uppercase opacity-50 tracking-tighter">Online</span>
          <span class="text-sm font-black tracking-tighter text-success">{{ onlineCount }}</span>
        </div>
      </div>
      <div class="flex items-center gap-1.5">
        <input v-model="keyword" type="text" placeholder="搜索 visitor / provider / proxy" class="input input-bordered input-xs w-56" />
        <select v-model="statusFilter" class="select select-bordered select-xs w-24">
          <option value="all">全部</option>
          <option value="online">在线</option>
          <option value="offline">离线</option>
          <option value="error">异常</option>
        </select>
        <button class="btn btn-circle btn-xs btn-ghost brook-action-icon" :class="{ loading }">
          <Icon icon="brook-refresh" style="font-size: 14px;" />
        </button>
      </div>
    </div>

    <section v-if="providerVisitors.length === 0" class="alert rounded-3xl border border-base-content/20 bg-base-100 shadow">
      <Icon icon="brook-empty" style="font-size: 16px" />
      <span>暂无可展示 Visitors</span>
    </section>

    <section v-else class="grid grid-cols-1 xl:grid-cols-2 2xl:grid-cols-3 gap-3 px-1">
      <article
        v-for="item in filtered"
        :key="item.id"
        class="brook-card brook-card-interactive overflow-hidden"
        @click="handleSelect(item)"
      >
        <div
          class="h-9 px-3 rounded-t-3xl border-b border-base-content/5 bg-gradient-to-r to-transparent flex items-center gap-2 text-[10px] font-black tracking-wider uppercase"
          :class="bandClass(item.status)"
        >
          <span class="inline-block w-2 h-2 rounded-full bg-current"></span>
          <span>{{ statusText(item.status) }}</span>
          <span class="ml-auto opacity-70">{{ item.protocol.toUpperCase() }}</span>
        </div>

        <div class="card-body p-4 gap-3">
          <div>
            <h3 class="text-base font-black tracking-tight">{{ item.name }}</h3>
            <p class="text-[11px] opacity-40 font-mono mt-1">{{ item.bindAddr }}:{{ item.bindPort }}</p>
          </div>

          <div class="grid grid-cols-4 gap-2">
            <div class="rounded-2xl bg-base-200/60 p-2">
              <div class="text-[10px] opacity-50">P95</div>
              <div class="text-xs font-black">{{ item.p95Latency }}ms</div>
            </div>
            <div class="rounded-2xl bg-base-200/60 p-2">
              <div class="text-[10px] opacity-50">连接</div>
              <div class="text-xs font-black">{{ item.currentConns }}</div>
            </div>
            <div class="rounded-2xl bg-base-200/60 p-2">
              <div class="text-[10px] opacity-50">入流量</div>
              <div class="text-xs font-black">{{ item.totalIn }}</div>
            </div>
            <div class="rounded-2xl bg-base-200/60 p-2">
              <div class="text-[10px] opacity-50">出流量</div>
              <div class="text-xs font-black">{{ item.totalOut }}</div>
            </div>
          </div>

          <div class="rounded-2xl border border-base-content/5 bg-base-100 p-3 flex items-center gap-2 text-xs">
            <div class="min-w-0 flex-1">
              <div class="opacity-50 text-[10px]">服务提供方</div>
              <div class="font-black text-secondary truncate">{{ item.providerName }}</div>
            </div>
            <Icon icon="brook-Right-" class="opacity-30" style="font-size: 12px" />
            <div class="min-w-0 flex-1">
              <div class="opacity-50 text-[10px]">关联 Proxy</div>
              <code class="text-primary truncate block">{{ item.serverName }}</code>
            </div>
          </div>

          <div class="flex items-center justify-between">
            <span class="text-[10px] opacity-40">最后心跳：{{ item.lastSeen }}</span>
            <div class="flex items-center gap-2">
              <span class="badge badge-soft" :class="statusBadgeClass(item.status)">{{ statusText(item.status) }}</span>
              <button class="btn btn-xs" :class="item.enabled ? 'btn-warning btn-soft' : 'btn-success btn-soft'" @click.stop="toggleEnabled(item)">
                {{ item.enabled ? '停用' : '启用' }}
              </button>
            </div>
          </div>
        </div>
      </article>

      <article v-if="selectedVisitor" class="col-span-full brook-card">
        <div class="px-4 py-3 border-b border-base-content/5 flex items-center justify-between">
          <div class="flex items-center gap-2">
            <Icon icon="brook-Diagram-" style="font-size: 14px;" class="opacity-40" />
            <span class="text-xs font-black uppercase tracking-widest opacity-60">Visitor Detail</span>
          </div>
          <button class="btn btn-ghost btn-xs" @click="selectedVisitor = null">关闭</button>
        </div>
        <div class="p-4 grid grid-cols-1 md:grid-cols-2 xl:grid-cols-4 gap-3">
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">名称</div>
            <div class="text-sm font-black">{{ selectedVisitor.name }}</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">协议</div>
            <div class="text-sm font-black uppercase">{{ selectedVisitor.protocol }}</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">服务提供方</div>
            <div class="text-sm font-black text-secondary truncate">{{ selectedVisitor.providerName }}</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">关联 Proxy</div>
            <div class="text-sm font-black text-primary truncate">{{ selectedVisitor.serverName }}</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">监听地址</div>
            <div class="text-sm font-mono font-black">{{ selectedVisitor.bindAddr }}:{{ selectedVisitor.bindPort }}</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">状态</div>
            <span class="badge badge-soft mt-1" :class="statusBadgeClass(selectedVisitor.status)">{{ statusText(selectedVisitor.status) }}</span>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">P95 延迟</div>
            <div class="text-sm font-black">{{ selectedVisitor.p95Latency }}ms</div>
          </div>
          <div class="rounded-2xl bg-base-200/50 p-3">
            <div class="text-[10px] opacity-50 uppercase">当前连接</div>
            <div class="text-sm font-black">{{ selectedVisitor.currentConns }}</div>
          </div>
        </div>
      </article>

      <div v-if="filtered.length === 0" class="col-span-full">
        <div class="alert border border-dashed border-base-content/20 bg-base-100 shadow-sm text-sm opacity-70">当前筛选条件下无结果</div>
      </div>
    </section>
  </div>
</template>
