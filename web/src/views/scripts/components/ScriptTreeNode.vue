<script setup lang="ts">
import { Delete, Document, Edit, Folder, MoreFilled } from '@element-plus/icons-vue'
import type { TreeNode } from '../types'

defineProps<{
  data: TreeNode
  onOpenRename: (path: string) => void
  onDelete: (path: string, isDir: boolean) => void | Promise<void>
}>()

function getFileIcon(node: TreeNode) {
  return node.isLeaf ? Document : Folder
}

function getFileIconColor(node: TreeNode): string {
  if (!node.isLeaf) return '#e6a23c'
  const ext = node.title.split('.').pop()?.toLowerCase()
  switch (ext) {
    case 'js':
      return '#f0db4f'
    case 'ts':
      return '#3178c6'
    case 'py':
      return '#4b8bbe'
    case 'sh':
      return '#4eaa25'
    case 'json':
      return '#e37e36'
    case 'yaml':
    case 'yml':
      return '#cb171e'
    case 'md':
      return '#083fa1'
    case 'html':
      return '#e34c26'
    case 'css':
      return '#264de4'
    default:
      return 'var(--el-text-color-secondary)'
  }
}
</script>

<template>
  <div class="tree-node">
    <el-icon size="14" :style="{ color: getFileIconColor(data) }">
      <component :is="getFileIcon(data)" />
    </el-icon>
    <span class="tree-node-label">{{ data.title }}</span>
    <span v-if="data.isLeaf && data.title.includes('.')" class="file-ext-badge">{{ data.title.split('.').pop()?.toUpperCase() }}</span>
    <div class="tree-node-actions" @click.stop>
      <el-dropdown trigger="click" size="small">
        <el-icon class="more-btn" :size="18"><MoreFilled /></el-icon>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="onOpenRename(data.key)">
              <el-icon><Edit /></el-icon>重命名
            </el-dropdown-item>
            <el-dropdown-item divided @click="onDelete(data.key, !data.isLeaf)">
              <el-icon><Delete /></el-icon><span style="color: var(--el-color-danger)">删除</span>
            </el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </div>
</template>

<style scoped lang="scss">
.tree-node {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
  overflow: hidden;

  .tree-node-label {
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 13px;
  }

  .file-ext-badge {
    font-size: 9px;
    font-weight: 700;
    font-family: var(--dd-font-mono);
    padding: 1px 4px;
    border-radius: 3px;
    background: var(--el-fill-color);
    color: var(--el-text-color-secondary);
    flex-shrink: 0;
    letter-spacing: 0.3px;
    line-height: 1.4;
    opacity: 0;
    transition: opacity 0.2s;
  }

  &:hover .file-ext-badge {
    opacity: 1;
  }

  .tree-node-actions {
    opacity: 0;
    transition: opacity 0.2s;
    flex-shrink: 0;

    .more-btn {
      cursor: pointer;
      padding: 4px;
      border-radius: 4px;
      font-size: 18px;
      color: var(--el-text-color-secondary);
      display: flex;
      align-items: center;
      justify-content: center;

      &:hover {
        background: var(--el-fill-color-light);
        color: var(--el-color-primary);
      }
    }
  }

  &:hover .tree-node-actions {
    opacity: 1;
  }
}
</style>
