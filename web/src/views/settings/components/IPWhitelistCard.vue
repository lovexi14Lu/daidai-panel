<script setup lang="ts">
import { Connection, Plus, Refresh } from '@element-plus/icons-vue'

const showAddIPDialog = defineModel<boolean>('showAddIPDialog', { required: true })
const newIP = defineModel<string>('newIP', { required: true })
const newIPRemarks = defineModel<string>('newIPRemarks', { required: true })

defineProps<{
  ipWhitelist: any[]
  ipWhitelistLoading: boolean
  onLoadIPWhitelist: () => void | Promise<void>
  onAddIP: () => void | Promise<void>
  onRemoveIP: (id: number) => void | Promise<void>
}>()
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div class="card-header">
        <span class="card-title"><el-icon><Connection /></el-icon> IP白名单</span>
        <div class="card-header-buttons">
          <el-button @click="onLoadIPWhitelist"><el-icon><Refresh /></el-icon>刷新</el-button>
          <el-button type="primary" @click="showAddIPDialog = true">
            <el-icon><Plus /></el-icon>添加IP/网段
          </el-button>
        </div>
      </div>
    </template>
    <p class="card-tip">
      支持单个 IP、CIDR 网段，以及更易输入的 IPv4 通配格式，例如 `203.0.113.7`、`203.0.113.0/24`、`203.0.113.*`。
    </p>
    <el-table :data="ipWhitelist" v-loading="ipWhitelistLoading" stripe empty-text="暂无数据">
      <el-table-column prop="ip" label="IP / 网段" min-width="220" />
      <el-table-column prop="remarks" label="描述" min-width="200" />
      <el-table-column label="状态" width="80">
        <template #default>
          <el-tag type="success" size="small">启用</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="created_at" label="创建时间" width="170">
        <template #default="{ row }">{{ new Date(row.created_at).toLocaleString() }}</template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button size="small" text type="danger" @click="onRemoveIP(row.id)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
  </el-card>

  <el-dialog v-model="showAddIPDialog" title="添加 IP 白名单" width="460px">
    <el-form label-width="80px">
      <el-form-item label="IP / 网段">
        <el-input v-model="newIP" placeholder="如: 203.0.113.7 / 203.0.113.0/24 / 203.0.113.*" />
        <div class="field-hint">
          适合固定公网填单个 IP；动态公网但前缀稳定时，可填网段或 `*` 通配格式。
        </div>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="newIPRemarks" placeholder="备注说明 (可选)" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="showAddIPDialog = false">取消</el-button>
      <el-button type="primary" @click="onAddIP">添加</el-button>
    </template>
  </el-dialog>
</template>

<style scoped lang="scss">
@use './config-card-shared.scss' as *;

.card-header-buttons {
  display: flex;
  gap: 8px;
}

.card-tip {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.6;
  color: var(--el-text-color-secondary);
}

.field-hint {
  margin-top: 6px;
  font-size: 12px;
  line-height: 1.5;
  color: var(--el-text-color-secondary);
}
</style>
