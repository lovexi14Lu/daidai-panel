<script setup lang="ts">
import { computed } from 'vue'
import { getDisplayTaskLabels } from '../taskLabels'

const props = defineProps<{
  visible: boolean
  task: any
}>()

const emit = defineEmits<{
  'update:visible': [value: boolean]
}>()

const statusText = computed(() => {
  if (props.task?.status === 0) return '已禁用'
  if (props.task?.status === 0.5) return '排队中'
  if (props.task?.status === 2) return '运行中'
  return '已启用'
})

const statusType = computed(() => {
  if (props.task?.status === 0) return 'info'
  if (props.task?.status === 2) return 'warning'
  return 'success'
})

const displayLabels = computed(() => {
  if (Array.isArray(props.task?.display_labels) && props.task.display_labels.length > 0) {
    return props.task.display_labels
  }
  return getDisplayTaskLabels(props.task?.labels || [])
})

const formatTime = (t: string) => {
  if (!t) return '-'
  return new Date(t).toLocaleString('zh-CN')
}

function handleClose() {
  emit('update:visible', false)
}
</script>

<template>
  <el-dialog
    :model-value="visible"
    title="任务详情"
    width="700px"
    @close="handleClose"
  >
    <el-descriptions v-if="task" :column="2" border>
      <el-descriptions-item label="任务名称" :span="2">
        <div style="display: flex; align-items: center; gap: 8px">
          <el-icon v-if="task.is_pinned" color="var(--el-color-warning)"><Star /></el-icon>
          <span>{{ task.name }}</span>
        </div>
      </el-descriptions-item>
      <el-descriptions-item label="任务ID">{{ task.id }}</el-descriptions-item>
      <el-descriptions-item label="状态">
        <el-tag :type="statusType" size="small">{{ statusText }}</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="定时规则" :span="2">
        <code>{{ task.cron_expression }}</code>
      </el-descriptions-item>
      <el-descriptions-item label="执行命令" :span="2">
        <code style="word-break: break-all">{{ task.command }}</code>
      </el-descriptions-item>
      <el-descriptions-item label="标签" :span="2">
        <el-tag v-for="label in displayLabels" :key="label" size="small" effect="plain" style="margin-right: 6px">
          {{ label }}
        </el-tag>
        <span v-if="displayLabels.length === 0" style="color: var(--el-text-color-placeholder)">无</span>
      </el-descriptions-item>
      <el-descriptions-item label="上次运行状态">
        <el-tag v-if="task.last_run_status === null" type="info" size="small">未运行</el-tag>
        <el-tag v-else-if="task.last_run_status === 0" type="success" size="small">成功</el-tag>
        <el-tag v-else type="danger" size="small">失败</el-tag>
      </el-descriptions-item>
      <el-descriptions-item label="上次运行耗时">
        <span v-if="task.last_running_time != null">{{ task.last_running_time.toFixed(2) }}s</span>
        <span v-else style="color: var(--el-text-color-placeholder)">-</span>
      </el-descriptions-item>
      <el-descriptions-item label="上次运行时间" :span="2">
        {{ formatTime(task.last_run_at) }}
      </el-descriptions-item>
      <el-descriptions-item label="下次运行时间" :span="2">
        {{ formatTime(task.next_run_at) }}
      </el-descriptions-item>
      <el-descriptions-item label="创建时间">{{ formatTime(task.created_at) }}</el-descriptions-item>
      <el-descriptions-item label="更新时间">{{ formatTime(task.updated_at) }}</el-descriptions-item>
      <el-descriptions-item label="备注" :span="2">
        <span v-if="task.remark">{{ task.remark }}</span>
        <span v-else style="color: var(--el-text-color-placeholder)">无</span>
      </el-descriptions-item>
    </el-descriptions>
  </el-dialog>
</template>
