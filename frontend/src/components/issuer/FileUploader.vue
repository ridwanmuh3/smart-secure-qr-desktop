<script lang="ts" setup>
import { Icon } from '@iconify/vue'

defineProps<{
  fileName: string
  fileSize: number
  documentHash: string
}>()

const emit = defineEmits<{
  select: []
}>()

function formatSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
</script>

<template>
  <div class="bg-white rounded-xl border border-gray-200 p-5 shadow-sm">
    <div class="flex items-center justify-between mb-3">
      <h3 class="text-sm font-semibold text-gray-700 flex items-center gap-2">
        <span class="w-6 h-6 bg-emerald-100 text-emerald-700 rounded-full flex items-center justify-center text-xs font-bold">1</span>
        Pilih Dokumen
      </h3>
    </div>

    <div v-if="!fileName"
      @click="emit('select')"
      class="border-2 border-dashed border-gray-300 rounded-lg p-8 text-center cursor-pointer hover:border-emerald-400 hover:bg-emerald-50/30 transition-all">
      <Icon icon="lucide:cloud-upload" class="w-10 h-10 text-gray-400 mx-auto mb-2" />
      <p class="text-sm text-gray-500">Klik untuk memilih dokumen</p>
    </div>

    <div v-else class="space-y-3">
      <div class="flex items-center gap-3 p-3 bg-emerald-50 rounded-lg border border-emerald-200">
        <Icon icon="lucide:file-text" class="w-8 h-8 text-emerald-600 shrink-0" />
        <div class="flex-1 min-w-0">
          <p class="font-medium text-gray-800 truncate">{{ fileName }}</p>
          <p class="text-xs text-gray-500">{{ formatSize(fileSize) }}</p>
        </div>
        <button @click="emit('select')" class="text-xs text-emerald-600 hover:text-emerald-700 font-medium">Ganti</button>
      </div>

      <div v-if="documentHash" class="p-3 bg-gray-50 rounded-lg">
        <p class="text-xs text-gray-500 mb-1">Hash Dokumen</p>
        <p class="text-xs font-mono text-gray-700 break-all">{{ documentHash }}</p>
      </div>
    </div>
  </div>
</template>
