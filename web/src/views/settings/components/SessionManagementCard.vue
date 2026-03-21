<script setup lang="ts">
import { Monitor, Refresh } from '@element-plus/icons-vue'

defineProps<{
  sessions: any[]
  sessionsLoading: boolean
  onLoadSessions: () => void | Promise<void>
  onRevokeAllSessions: () => void | Promise<void>
  onRevokeSession: (id: number) => void | Promise<void>
}>()
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div class="card-header">
        <span class="card-title"><el-icon><Monitor /></el-icon> 活动会话</span>
        <div class="card-header-buttons">
          <el-button @click="onLoadSessions"><el-icon><Refresh /></el-icon>刷新</el-button>
          <el-button type="danger" plain @click="onRevokeAllSessions">撤销所有其他会话</el-button>
        </div>
      </div>
    </template>
    <el-table :data="sessions" v-loading="sessionsLoading" stripe empty-text="暂无数据">
      <el-table-column prop="ip" label="IP地址" width="140" />
      <el-table-column prop="user_agent" label="用户代理" show-overflow-tooltip />
      <el-table-column label="最后活动" width="170">
        <template #default="{ row }">{{ new Date(row.last_active || row.created_at).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="danger" @click="onRevokeSession(row.id)">撤销</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>
</template>

<style scoped lang="scss">
@use './config-card-shared.scss' as *;

.card-header-buttons {
  display: flex;
  gap: 8px;
}
</style>
